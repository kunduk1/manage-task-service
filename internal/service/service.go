package service

//go:generate go tool mockgen -source=service.go -destination=mocks/service_mock.go -package=mocks

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/model"
)

type AuthService interface {
	Register(ctx context.Context, in model.RegisterInput) (*model.User, error)
	Login(ctx context.Context, in model.LoginInput) (*model.AuthTokens, error)
}

type TeamsService interface {
	Create(ctx context.Context, in model.CreateTeamInput) (*model.Team, error)
	List(ctx context.Context, userID int64) ([]model.Team, error)
	Invite(ctx context.Context, in model.InviteInput) error
}
