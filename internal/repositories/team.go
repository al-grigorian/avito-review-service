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
	var team models.Team
	team.TeamName = name

	err := r.db.SelectContext(ctx, &team.Members,
		`SELECT user_id, username, is_active FROM users WHERE team_name = $1`, name)
	if err != nil {
		return nil, false
	}
	if len(team.Members) == 0 {
		return nil, false
	}
	return &team, true
}

func (r *TeamRepository) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	var team models.Team
	team.TeamName = teamName

	err := r.db.SelectContext(ctx, &team.Members,
		`SELECT user_id, username, is_active FROM users WHERE team_name = $1 ORDER BY user_id`, teamName)
	if err != nil {
		return nil, err
	}
	if len(team.Members) == 0 {
		return nil, domain.ErrTeamNotFound
	}
	return &team, nil
}
