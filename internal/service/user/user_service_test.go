package user_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"saviour/internal/logger"
	"saviour/internal/model"
	"saviour/internal/repository"
	"saviour/internal/service/auth"
	"saviour/internal/service/user"
	"saviour/internal/testkit/testmock"
)

func TestUserService_CreateUser(t *testing.T) {
	t.Parallel()

	userRepo := testmock.NewMockUserRepository(t)

	userService := user.NewService(
		logger.Noop,
		userRepo,
	)

	userRepo.EXPECT().
		CreateUser(
			mock.Anything,
			mock.MatchedBy(
				func(params repository.CreateUserParams) bool {
					return params.UUID == uuid.Nil &&
						params.Username == "test_user" &&
						auth.CheckPassword(
							"password",
							params.PasswordHash,
						) == nil &&
						params.IsAdmin == true
				},
			),
		).
		RunAndReturn(
			func(
				ctx context.Context,
				params repository.CreateUserParams,
			) (*model.User, error) {
				return &model.User{
					UUID:         uuid.New(),
					CreatedAt:    time.Now(),
					Username:     params.Username,
					PasswordHash: params.PasswordHash,
					IsAdmin:      params.IsAdmin,
				}, nil
			},
		)

	usr, err := userService.CreateUser(
		context.Background(),
		user.CreateUserParams{
			Username: "test_user",
			Password: "password",
			IsAdmin:  true,
		},
	)

	require.NoError(t, err)
	require.NotEmpty(t, usr.UUID)
	require.Equal(t, "test_user", usr.Username)
	require.True(t, usr.IsAdmin)
}

func TestUserService_DeactivateUser_OK(t *testing.T) {
	t.Parallel()

	userRepo := testmock.NewMockUserRepository(t)

	userService := user.NewService(
		logger.Noop,
		userRepo,
	)

	userUUID := uuid.New()

	userRepo.EXPECT().
		DeactivateUser(mock.Anything, userUUID).
		Return(nil)

	err := userService.DeactivateUser(context.Background(), userUUID)

	require.NoError(t, err)
}

func TestUserService_DeactivateUser_NotFound(t *testing.T) {
	t.Parallel()

	userRepo := testmock.NewMockUserRepository(t)

	userService := user.NewService(
		logger.Noop,
		userRepo,
	)

	userUUID := uuid.New()

	userRepo.EXPECT().
		DeactivateUser(mock.Anything, userUUID).
		Return(repository.ErrUserNotFound)

	err := userService.DeactivateUser(
		context.Background(),
		userUUID,
	)

	require.ErrorIs(t, err, user.ErrUserNotFound)
}
