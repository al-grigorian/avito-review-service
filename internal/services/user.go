package services

import (
	"context"

	"github.com/al-grigorian/avito-review-service/internal/models"
	"github.com/al-grigorian/avito-review-service/internal/repositories"
)

type UserService struct {
	userRepo *repositories.UserRepository
	prRepo   *repositories.PRRepository
}

func NewUserService(userRepo *repositories.UserRepository, prRepo *repositories.PRRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func (s *UserService) SetActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	return s.userRepo.SetActive(ctx, userID, isActive)
}

func (s *UserService) GetReviewPRs(ctx context.Context, userID string) (*models.GetReviewResponse, error) {
	prs, err := s.prRepo.GetReviewPRs(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &models.GetReviewResponse{
		UserID:       userID,
		PullRequests: prs,
	}, nil
}
