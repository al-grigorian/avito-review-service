package domain

import "errors"

var (
	ErrTeamExists        = errors.New("team_name already exists")
	ErrNoActiveReviewers = errors.New("no active reviewers in team")
	ErrTeamNotFound      = errors.New("team not found")
	ErrPRNotFound        = errors.New("pull request not found")
	ErrAlreadyMerged     = errors.New("pull request already merged")
	ErrUserNotFound      = errors.New("user not found")
	ErrPRMerged          = errors.New("cannot reassign on merged PR")
	ErrNotAssigned       = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate       = errors.New("no active replacement candidate in team")
)
