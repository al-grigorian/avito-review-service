package models

import "time"

type PullRequest struct {
	ID        string     `json:"pull_request_id" db:"pull_request_id"`
	Name      string     `json:"pull_request_name" db:"pull_request_name"`
	AuthorID  string     `json:"author_id" db:"author_id"`
	Status    string     `json:"status" db:"status"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	MergedAt  *time.Time `json:"merged_at,omitempty" db:"merged_at"`
	Reviewers []User     `json:"reviewers"`
}

type CreatePullRequestRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type CreatePullRequestResponse struct {
	PullRequest PullRequest `json:"pull_request"`
}
