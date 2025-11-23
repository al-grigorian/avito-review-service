package repositories

import (
	"context"
	"database/sql"
	"math/rand"
	"time"

	"github.com/al-grigorian/avito-review-service/internal/domain"
	"github.com/al-grigorian/avito-review-service/internal/models"
	"github.com/jmoiron/sqlx"
)

type PRRepository struct {
	db *sqlx.DB
}

func NewPRRepository(db *sqlx.DB) *PRRepository {
	return &PRRepository{db: db}
}

func (r *PRRepository) CreatePR(ctx context.Context, pr models.PullRequest) (*models.PullRequest, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var teamName string
	err = tx.GetContext(ctx, &teamName,
		`SELECT team_name FROM users WHERE user_id = $1`, pr.AuthorID)
	if err != nil {
		return nil, err
	}

	var candidates []models.User
	err = tx.SelectContext(ctx, &candidates,
		`SELECT user_id, username, is_active FROM users 
         WHERE team_name = $1 AND is_active = true AND user_id != $2`,
		teamName, pr.AuthorID)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return nil, domain.ErrNoActiveReviewers
	}

	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	count := 2
	if len(candidates) < 2 {
		count = len(candidates)
	}
	reviewers := candidates[:count]

	now := time.Now()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
         VALUES ($1, $2, $3, 'OPEN', $4)`,
		pr.ID, pr.Name, pr.AuthorID, now)
	if err != nil {
		return nil, err
	}

	for _, rev := range reviewers {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1, $2)`, pr.ID, rev.UserID)
		if err != nil {
			return nil, err
		}
	}

	pr.Reviewers = reviewers
	pr.CreatedAt = now
	pr.Status = "OPEN"

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &pr, nil
}

func (r *PRRepository) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()

	var pr models.PullRequest
	err = tx.QueryRowxContext(ctx,
		`UPDATE pull_requests 
         SET status = 'MERGED', merged_at = $1 
         WHERE pull_request_id = $2 AND status = 'OPEN'
         RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at`,
		now, prID).StructScan(&pr)

	if err == sql.ErrNoRows {
		err = tx.QueryRowxContext(ctx,
			`SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at 
             FROM pull_requests WHERE pull_request_id = $1`, prID).StructScan(&pr)
		if err == sql.ErrNoRows {
			return nil, domain.ErrPRNotFound
		}
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	err = tx.SelectContext(ctx, &pr.Reviewers,
		`SELECT u.user_id, u.username, u.is_active 
         FROM pr_reviewers pr 
         JOIN users u ON u.user_id = pr.user_id 
         WHERE pr.pull_request_id = $1`, prID)
	if err != nil {
		return nil, err
	}

	pr.MergedAt = &now

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &pr, nil
}

func (r *PRRepository) ReassignReviewer(
	ctx context.Context,
	prID, oldUserID string,
) (newUserID string, pr *models.PullRequest, err error) {

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", nil, err
	}
	defer tx.Rollback()

	// проверяем статус PR
	var status string
	err = tx.GetContext(ctx, &status,
		`SELECT status FROM pull_requests WHERE pull_request_id = $1`, prID)
	if err == sql.ErrNoRows {
		return "", nil, domain.ErrPRNotFound
	}
	if err != nil {
		return "", nil, err
	}
	if status == "MERGED" {
		return "", nil, domain.ErrPRMerged
	}

	// проверяем, что oldUserID действительно ревьювер
	var exists bool
	err = tx.GetContext(ctx, &exists,
		`SELECT EXISTS(
            SELECT 1 FROM pr_reviewers 
            WHERE pull_request_id = $1 AND user_id = $2
         )`, prID, oldUserID)
	if err != nil {
		return "", nil, err
	}
	if !exists {
		return "", nil, domain.ErrNotAssigned
	}

	// получаем команду автора PR
	var teamName string
	err = tx.GetContext(ctx, &teamName,
		`SELECT u.team_name 
         FROM pull_requests pr 
         JOIN users u ON u.user_id = pr.author_id 
         WHERE pr.pull_request_id = $1`, prID)
	if err != nil {
		return "", nil, err
	}

	// кандидаты на замену: активные, из той же команды, не автор, не oldUserID
	var candidates []models.User
	err = tx.SelectContext(ctx, &candidates,
		`SELECT user_id, username, is_active 
         FROM users 
         WHERE team_name = $1 
           AND is_active = true 
           AND user_id != (SELECT author_id FROM pull_requests WHERE pull_request_id = $2)
           AND user_id != $3`,
		teamName, prID, oldUserID)
	if err != nil {
		return "", nil, err
	}
	if len(candidates) == 0 {
		return "", nil, domain.ErrNoCandidate
	}

	// случайный кандидат
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	newUser := candidates[0]

	// заменяем в pr_reviewers
	_, err = tx.ExecContext(ctx,
		`DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2`, prID, oldUserID)
	if err != nil {
		return "", nil, err
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1, $2)
		ON CONFLICT (pull_request_id, user_id) DO NOTHING`, prID, newUser.UserID)
	if err != nil {
		return "", nil, err
	}

	// возвращаем обновлённый PR
	prModel := &models.PullRequest{ID: prID}
	err = tx.QueryRowxContext(ctx,
		`SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
         FROM pull_requests WHERE pull_request_id = $1`, prID).StructScan(prModel)
	if err != nil {
		return "", nil, err
	}
	_ = tx.SelectContext(ctx, &prModel.Reviewers,
		`SELECT u.user_id, u.username, u.is_active 
         FROM pr_reviewers pr 
         JOIN users u ON u.user_id = pr.user_id 
         WHERE pr.pull_request_id = $1`, prID)

	if err = tx.Commit(); err != nil {
		return "", nil, err
	}

	return newUser.UserID, prModel, nil
}

func (r *PRRepository) GetReviewPRs(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	var prs []models.PullRequestShort
	err := r.db.SelectContext(ctx, &prs,
		`SELECT 
			pr.pull_request_id,
			pr.pull_request_name,
			pr.author_id,
			pr.status
		 FROM pull_requests pr
		 JOIN pr_reviewers rev ON rev.pull_request_id = pr.pull_request_id
		 WHERE rev.user_id = $1 AND pr.status = 'OPEN'
		 ORDER BY pr.created_at DESC`, userID)
	return prs, err
}
