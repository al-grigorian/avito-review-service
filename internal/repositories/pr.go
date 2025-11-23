package repositories

import (
	"context"
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
