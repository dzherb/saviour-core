package user

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"

	"saviour/internal/model"
	"saviour/internal/repository"
	"saviour/internal/service/auth"
)

type ServiceImpl struct {
	log      *slog.Logger
	userRepo repository.UserRepository
}

func NewService(
	log *slog.Logger,
	userRepo repository.UserRepository,
) *ServiceImpl {
	return &ServiceImpl{
		log:      log,
		userRepo: userRepo,
	}
}

func (s ServiceImpl) CreateUser(
	ctx context.Context,
	params CreateUserParams,
) (*model.User, error) {
	passwordHash, err := auth.HashPassword(params.Password)
	if err != nil {
		return nil, err
	}

	return s.userRepo.CreateUser(
		ctx,
		repository.CreateUserParams{
			Username:     params.Username,
			PasswordHash: passwordHash,
			IsAdmin:      params.IsAdmin,
		},
	)
}

func (s ServiceImpl) DeactivateUser(ctx context.Context, uuid uuid.UUID) error {
	err := s.userRepo.DeactivateUser(ctx, uuid)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
	}

	return nil
}
