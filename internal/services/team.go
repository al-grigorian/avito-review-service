package services

import (
	"context"
	"errors"

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
	if _, exists := s.teamRepo.GetTeamByName(ctx, team.TeamName); exists {
		return ErrTeamExists
	}
	return s.teamRepo.UpsertTeam(ctx, team)
}
