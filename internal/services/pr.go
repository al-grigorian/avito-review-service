package services

import (
	"context"

	"github.com/al-grigorian/avito-review-service/internal/models"
	"github.com/al-grigorian/avito-review-service/internal/repositories"
)

type PRService struct {
	prRepo *repositories.PRRepository
}

func NewPRService(prRepo *repositories.PRRepository) *PRService {
	return &PRService{prRepo: prRepo}
}

func (s *PRService) CreatePullRequest(ctx context.Context, req models.CreatePullRequestRequest) (*models.CreatePullRequestResponse, error) {
	pr := models.PullRequest{
		ID:       req.PullRequestID,
		Name:     req.PullRequestName,
		AuthorID: req.AuthorID,
	}

	created, err := s.prRepo.CreatePR(ctx, pr)
	if err != nil {
		return nil, err
	}

	return &models.CreatePullRequestResponse{PullRequest: *created}, nil
}

func (s *PRService) MergePullRequest(ctx context.Context, prID string) (*models.CreatePullRequestResponse, error) {
	pr, err := s.prRepo.MergePR(ctx, prID)
	if err != nil {
		return nil, err
	}
	return &models.CreatePullRequestResponse{PullRequest: *pr}, nil
}

func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (
	newUserID string, pr *models.PullRequest, err error) {
	return s.prRepo.ReassignReviewer(ctx, prID, oldUserID)
}
