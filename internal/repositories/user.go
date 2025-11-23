package repositories

import (
	"context"
	"database/sql"

	"github.com/al-grigorian/avito-review-service/internal/domain"
	"github.com/al-grigorian/avito-review-service/internal/models"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) SetActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowxContext(ctx,
		`UPDATE users 
         SET is_active = $1 
         WHERE user_id = $2
         RETURNING user_id, username, team_name, is_active`,
		isActive, userID).StructScan(&user)
	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
