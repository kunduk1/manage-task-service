package errors

import "errors"

var (
	ErrUserExists           = errors.New("user already exists")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserNotFound         = errors.New("user not found")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrValidation           = errors.New("validation error")
	ErrCacheMiss            = errors.New("cache: key not found")
	ErrTeamNotFound         = errors.New("team not found")
	ErrTaskNotFound         = errors.New("task not found")
	ErrMemberExists         = errors.New("team member already exists")
	ErrNotTeamMember        = errors.New("user is not a team member")
	ErrForbidden            = errors.New("forbidden")
)
