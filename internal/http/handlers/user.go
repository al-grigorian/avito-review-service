package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/al-grigorian/avito-review-service/internal/domain"
	"github.com/al-grigorian/avito-review-service/internal/models"
	"github.com/al-grigorian/avito-review-service/internal/services"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(s *services.UserService) *UserHandler {
	return &UserHandler{service: s}
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"invalid json"}}`, http.StatusBadRequest)
		return
	}
	if req.UserID == "" {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"user_id required"}}`, http.StatusBadRequest)
		return
	}

	user, err := h.service.SetActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			http.Error(w, `{"error":{"code":"NOT_FOUND","message":"user not found"}}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":{"code":"INTERNAL","message":"server error"}}`, http.StatusInternalServerError)
		return
	}

	resp := models.UserResponse{
		User: struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			TeamName string `json:"team_name"`
			IsActive bool   `json:"is_active"`
		}{
			UserID:   user.UserID,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"user_id required"}}`, http.StatusBadRequest)
		return
	}

	resp, err := h.service.GetReviewPRs(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":{"code":"INTERNAL","message":"server error"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
