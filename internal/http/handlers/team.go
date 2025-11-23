package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/al-grigorian/avito-review-service/internal/domain"
	"github.com/al-grigorian/avito-review-service/internal/models"
	"github.com/al-grigorian/avito-review-service/internal/services"
)

type TeamHandler struct {
	service *services.TeamService
}

func NewTeamHandler(s *services.TeamService) *TeamHandler {
	return &TeamHandler{service: s}
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var team models.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"invalid json"}}`, http.StatusBadRequest)
		return
	}

	if team.TeamName == "" || len(team.Members) == 0 {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"team_name and members required"}}`, http.StatusBadRequest)
		return
	}

	if err := h.service.CreateTeam(r.Context(), team); err != nil {
		if errors.Is(err, domain.ErrTeamExists) {
			http.Error(w, `{"error":{"code":"TEAM_EXISTS","message":"team_name already exists"}}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":{"code":"INTERNAL","message":"server error"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"team": team})
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"team_name required"}}`, http.StatusBadRequest)
		return
	}

	team, err := h.service.GetTeam(r.Context(), teamName)
	if err != nil {
		if errors.Is(err, domain.ErrTeamNotFound) {
			http.Error(w, `{"error":{"code":"TEAM_NOT_FOUND","message":"team not found"}}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":{"code":"INTERNAL","message":"server error"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}
