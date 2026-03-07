package auth_test

import (
	"context"
	"net/netip"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"saviour/internal/logger"
	"saviour/internal/model"
	"saviour/internal/repository"
	"saviour/internal/service/auth"
	"saviour/internal/testkit/testmock"
	"saviour/pkg/secret"
)

func testConfig() auth.Config {
	return auth.Config{
		AccessTokenSecret:  secret.New("test1"),
		RefreshTokenSecret: secret.New("test2"),
	}
}

func TestAuth_CreateSession_OK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	userRepo := testmock.NewMockUserRepository(t)
	sessionRepo := testmock.NewMockSessionRepository(t)

	password := "pass"

	passwordHash, err := auth.HashPassword(password)
	require.NoError(t, err)

	userRepo.EXPECT().
		GetUserByUsername(
			ctx,
			"user1",
		).
		Return(
			&model.User{
				UUID:         uuid.New(),
				PasswordHash: passwordHash,
			},
			nil,
		)

	sessionRepo.EXPECT().
		CreateSession(ctx, mock.Anything).
		Return(&model.Session{}, nil)

	authService := auth.New(
		logger.Noop,
		userRepo,
		sessionRepo,
		testConfig(),
	)

	pair, err := authService.CreateSession(
		ctx,
		auth.CreateSessionParams{
			Username:  "user1",
			Password:  "pass",
			UserAgent: "ua",
			IP:        netip.MustParseAddr("1.1.1.1"),
		},
	)

	require.NoError(t, err)
	require.NotNil(t, pair)
	require.NotEmpty(t, pair.AccessToken)
	require.NotEmpty(t, pair.RefreshToken)
}

func TestAuth_RefreshSession_OK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	userID := uuid.New()
	sessionID := uuid.New()

	cfg := testConfig()
	refreshToken, _ := auth.IssueToken(
		cfg.RefreshTokenSecret.UnsafeValue(),
		userID,
		sessionID,
		nil,
		auth.RefreshTokenTTL,
	)

	userRepo := testmock.NewMockUserRepository(t)
	sessionRepo := testmock.NewMockSessionRepository(t)

	sessionRepo.EXPECT().
		RefreshActiveSession(
			mock.Anything,
			mock.Anything,
		).
		Return(nil)

	authService := auth.New(
		logger.Noop,
		userRepo,
		sessionRepo,
		cfg,
	)

	pair, err := authService.RefreshSession(
		ctx,
		auth.RefreshSessionParams{
			RefreshToken: refreshToken,
			UserAgent:    "ua",
			IP:           netip.MustParseAddr("1.1.1.1"),
		},
	)

	require.NoError(t, err)
	require.NotNil(t, pair)
}

func TestAuth_RefreshSession_SessionNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cfg := testConfig()
	refreshToken, _ := auth.IssueToken(
		cfg.RefreshTokenSecret.UnsafeValue(),
		uuid.New(),
		uuid.New(),
		nil,
		auth.RefreshTokenTTL,
	)

	sessionRepo := testmock.NewMockSessionRepository(t)

	sessionRepo.
		EXPECT().
		RefreshActiveSession(
			mock.Anything,
			mock.Anything,
		).
		Return(repository.ErrSessionNotFound)

	authService := auth.New(
		logger.Noop,
		nil,
		sessionRepo,
		cfg,
	)

	_, err := authService.RefreshSession(
		ctx, auth.RefreshSessionParams{
			RefreshToken: refreshToken,
		},
	)

	require.ErrorIs(t, err, auth.ErrSessionNotFound)
}

func TestAuth_RevokeSession_OK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	userID := uuid.New()
	sessionID := uuid.New()

	sessionRepo := testmock.NewMockSessionRepository(t)

	sessionRepo.
		EXPECT().
		RevokeSession(
			ctx,
			repository.RevokeSessionParams{
				UUID:     sessionID,
				UserUUID: userID,
			},
		).
		Return(nil)

	authService := auth.New(
		logger.Noop,
		nil,
		sessionRepo,
		testConfig(),
	)

	err := authService.RevokeSession(
		ctx,
		auth.RevokeSessionParams{
			UserUUID:    userID,
			SessionUUID: sessionID,
		},
	)

	require.NoError(t, err)
}

func TestAuth_RevokeSession_SessionNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	userID := uuid.New()
	sessionID := uuid.New()

	sessionRepo := testmock.NewMockSessionRepository(t)

	sessionRepo.
		EXPECT().
		RevokeSession(
			ctx,
			repository.RevokeSessionParams{
				UUID:     sessionID,
				UserUUID: userID,
			},
		).
		Return(repository.ErrSessionNotFound)

	authService := auth.New(
		logger.Noop,
		nil,
		sessionRepo,
		testConfig(),
	)

	err := authService.RevokeSession(
		ctx,
		auth.RevokeSessionParams{
			UserUUID:    userID,
			SessionUUID: sessionID,
		},
	)

	require.ErrorIs(t, err, auth.ErrSessionNotFound)
}
