package domain

import "errors"

var (
	ErrTeamExists        = errors.New("team_name already exists")
	ErrNoActiveReviewers = errors.New("no active reviewers in team")
	ErrTeamNotFound      = errors.New("team not found")
)
