package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/al-grigorian/avito-review-service/internal/models"
	"github.com/al-grigorian/avito-review-service/internal/services"
)

type PRHandler struct {
	service *services.PRService
}

func NewPRHandler(s *services.PRService) *PRHandler {
	return &PRHandler{service: s}
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePullRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"invalid json"}}`, http.StatusBadRequest)
		return
	}

	resp, err := h.service.CreatePullRequest(r.Context(), req)
	if err != nil {
		log.Printf("ERROR creating PR: %v", err)
		http.Error(w, `{"error":{"code":"INTERNAL","message":"server error"}}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	// ... код merge
}

func (h *PRHandler) GetReviewers(w http.ResponseWriter, r *http.Request) {
	// ... код reviewers
}
