package repositories

import (
	"context"

	"github.com/al-grigorian/avito-review-service/internal/domain"
	"github.com/al-grigorian/avito-review-service/internal/models"

	"github.com/jmoiron/sqlx"
)

type TeamRepository struct {
	db *sqlx.DB
}

func NewTeamRepository(db *sqlx.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) TeamExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists,
		`SELECT EXISTS (SELECT 1 FROM teams WHERE name = $1)`, name)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *TeamRepository) UpsertTeam(ctx context.Context, team models.Team) error {
	// ← КЛЮЧЕВАЯ ПРОВЕРКА ДО ВСЕГО!
	exists, err := r.TeamExists(ctx, team.TeamName)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrTeamExists
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Теперь точно знаем, что команды нет → вставляем
	_, err = tx.ExecContext(ctx,
		`INSERT INTO teams (name) VALUES ($1)`, team.TeamName)
	if err != nil {
		return err
	}

	for _, m := range team.Members {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO users (user_id, username, team_name, is_active)
             VALUES ($1, $2, $3, $4)
             ON CONFLICT (user_id) DO UPDATE SET
                 username = EXCLUDED.username,
                 team_name = EXCLUDED.team_name,
                 is_active = EXCLUDED.is_active`,
			m.UserID, m.Username, team.TeamName, m.IsActive)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *TeamRepository) GetTeamByName(ctx context.Context, name string) (*models.Team, bool) {
	var t models.Team
	t.TeamName = name
	err := r.db.SelectContext(ctx, &t.Members,
		`SELECT user_id, username, is_active FROM users WHERE team_name = $1`, name)
	if err != nil || len(t.Members) == 0 {
		return nil, false
	}
	return &t, true
}
