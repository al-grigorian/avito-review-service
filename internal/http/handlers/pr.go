package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/al-grigorian/avito-review-service/internal/domain"
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
	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"invalid json"}}`, http.StatusBadRequest)
		return
	}
	if req.PullRequestID == "" {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"pull_request_id required"}}`, http.StatusBadRequest)
		return
	}

	resp, err := h.service.MergePullRequest(r.Context(), req.PullRequestID)
	if err != nil {
		if errors.Is(err, domain.ErrPRNotFound) {
			http.Error(w, `{"error":{"code":"NOT_FOUND","message":"resource not found"}}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":{"code":"INTERNAL","message":"server error"}}`, http.StatusInternalServerError)
		return
	}

	// Ответ строго по спецификации
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"pr": map[string]any{
			"pull_request_id":    resp.PullRequest.ID,
			"pull_request_name":  resp.PullRequest.Name,
			"author_id":          resp.PullRequest.AuthorID,
			"status":             resp.PullRequest.Status,
			"assigned_reviewers": extractUserIDs(resp.PullRequest.Reviewers),
			"mergedAt":           resp.PullRequest.MergedAt,
		},
	})
}

func extractUserIDs(users []models.User) []string {
	ids := make([]string, len(users))
	for i, u := range users {
		ids[i] = u.UserID
	}
	return ids
}

func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldReviewerID string `json:"old_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"invalid json"}}`, http.StatusBadRequest)
		return
	}
	if req.PullRequestID == "" || req.OldReviewerID == "" {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"pull_request_id and old_user_id required"}}`, http.StatusBadRequest)
		return
	}

	newUserID, pr, err := h.service.ReassignReviewer(r.Context(), req.PullRequestID, req.OldReviewerID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPRNotFound):
			http.Error(w, `{"error":{"code":"NOT_FOUND","message":"resource not found"}}`, http.StatusNotFound)
		case errors.Is(err, domain.ErrPRMerged):
			http.Error(w, `{"error":{"code":"PR_MERGED","message":"cannot reassign on merged PR"}}`, http.StatusConflict)
		case errors.Is(err, domain.ErrNotAssigned):
			http.Error(w, `{"error":{"code":"NOT_ASSIGNED","message":"reviewer is not assigned to this PR"}}`, http.StatusConflict)
		case errors.Is(err, domain.ErrNoCandidate):
			http.Error(w, `{"error":{"code":"NO_CANDIDATE","message":"no active replacement candidate in team"}}`, http.StatusConflict)
		default:
			http.Error(w, `{"error":{"code":"INTERNAL","message":"server error"}}`, http.StatusInternalServerError)
		}
		return
	}

	resp := map[string]any{
		"pr": map[string]any{
			"pull_request_id":    pr.ID,
			"pull_request_name":  pr.Name,
			"author_id":          pr.AuthorID,
			"status":             pr.Status,
			"assigned_reviewers": extractUserIDs(pr.Reviewers),
			"createdAt":          pr.CreatedAt,
			"mergedAt":           pr.MergedAt,
		},
		"replaced_by": newUserID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
