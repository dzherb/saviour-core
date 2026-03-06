package user

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"saviour/internal/model"
)

type Service interface {
	CreateUser(
		ctx context.Context,
		params CreateUserParams,
	) (*model.User, error)
	DeactivateUser(ctx context.Context, uuid uuid.UUID) error
}

var (
	ErrUserNotFound = errors.New("user not found")
)

type CreateUserParams struct {
	Username string
	Password string
	IsAdmin  bool
}
