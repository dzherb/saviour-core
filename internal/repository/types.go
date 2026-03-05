package repository

import (
	"context"

	"github.com/google/uuid"

	"saviour/internal/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, param CreateUserParams) (*model.User, error)
}

type CreateUserParams struct {
	UUID         uuid.UUID
	Username     string
	PasswordHash string
}
