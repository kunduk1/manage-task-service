package user

import (
	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/repository"
)

type repo struct {
	db db.Client
}

func NewRepository(dbc db.Client) repository.UserRepository {
	return &repo{db: dbc}
}
