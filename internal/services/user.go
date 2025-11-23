package services

import (
	"context"

	"github.com/al-grigorian/avito-review-service/internal/models"
	"github.com/al-grigorian/avito-review-service/internal/repositories"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: repo}
}

func (s *UserService) SetActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	return s.userRepo.SetActive(ctx, userID, isActive)
}
