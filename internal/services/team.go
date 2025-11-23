package services

import (
	"context"
	"errors"

	"github.com/al-grigorian/avito-review-service/internal/domain"
	"github.com/al-grigorian/avito-review-service/internal/models"
	"github.com/al-grigorian/avito-review-service/internal/repositories"
)

var ErrTeamExists = errors.New("team already exists")

type TeamService struct {
	teamRepo *repositories.TeamRepository
}

func NewTeamService(repo *repositories.TeamRepository) *TeamService {
	return &TeamService{teamRepo: repo}
}

func (s *TeamService) CreateTeam(ctx context.Context, team models.Team) error {
	if _, err := s.teamRepo.GetTeam(ctx, team.TeamName); err == nil {
		return domain.ErrTeamExists
	}
	return s.teamRepo.UpsertTeam(ctx, team)
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*models.TeamResponse, error) {
	team, err := s.teamRepo.GetTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}
	return &models.TeamResponse{Team: *team}, nil
}
