// Package v1 — публичный контракт транспортного уровня для домена teams (HTTP/JSON).
package v1

import "time"

type CreateTeamRequest struct {
	Name        string `json:"name"        example:"Platform"`
	Description string `json:"description" example:"Core platform team"`
}

type TeamResponse struct {
	ID          int64     `json:"id"          example:"1"`
	Name        string    `json:"name"        example:"Platform"`
	Description string    `json:"description" example:"Core platform team"`
	CreatedBy   int64     `json:"created_by"  example:"42"`
	CreatedAt   time.Time `json:"created_at"  example:"2026-06-16T12:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at"  example:"2026-06-16T12:00:00Z"`
}

type ListTeamsResponse struct {
	Teams []TeamResponse `json:"teams"`
}

type InviteRequest struct {
	UserID int64  `json:"user_id"        example:"7"`
	Role   string `json:"role,omitempty" example:"member"`
}

type InviteResponse struct {
	TeamID int64  `json:"team_id" example:"1"`
	UserID int64  `json:"user_id" example:"7"`
	Role   string `json:"role"    example:"member"`
}
