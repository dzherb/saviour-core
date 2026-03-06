package sqliterepo_test

import (
	"context"
	"database/sql"
	"net/netip"
	"testing"
	"time"

	"github.com/google/uuid"

	"saviour/internal/model"
	"saviour/internal/repository"
	sqliterepo "saviour/internal/repository/sqlite"
)

func createTestUser(t *testing.T, db *sql.DB) (*model.User, error) {
	t.Helper()

	userRepo := sqliterepo.NewUserRepository(db)
	userUUID := uuid.New()

	return userRepo.CreateUser(
		context.Background(),
		repository.CreateUserParams{
			UUID:         userUUID,
			Username:     "test_user",
			PasswordHash: "password",
		},
	)
}

func (suite *RepositoryTestSuite) TestSessionRepository_CreateSession() {
	user, err := createTestUser(suite.T(), suite.DB)

	suite.Require().Nil(err)

	sessionRepo := sqliterepo.NewSessionsRepository(suite.DB)
	sessionUUID := uuid.New()
	ip := netip.MustParseAddr("0.0.0.0")

	session, err := sessionRepo.CreateSession(
		context.Background(),
		repository.CreateSessionParams{
			UUID:             sessionUUID,
			UserUUID:         user.UUID,
			RefreshTokenHash: "test_hash",
			UserAgent:        "test_agent",
			IP:               ip,
		},
	)

	suite.Require().NoError(err)
	suite.Equal(sessionUUID, session.UUID)
	suite.Equal(user.UUID, session.UserUUID)
	suite.Equal("test_agent", session.UserAgent)
	suite.Equal(ip, session.IP)
	suite.True(session.LastRefreshAt.After(time.Now().Add(-5 * time.Second)))
}

func (suite *RepositoryTestSuite) TestSessionRepository_RefreshSession_OK() {
	user, err := createTestUser(suite.T(), suite.DB)
	suite.Require().Nil(err)

	sessionRepo := sqliterepo.NewSessionsRepository(suite.DB)
	sessionUUID := uuid.New()

	_, err = sessionRepo.CreateSession(
		context.Background(),
		repository.CreateSessionParams{
			UUID:             sessionUUID,
			UserUUID:         user.UUID,
			RefreshTokenHash: "refresh_hash_1",
			UserAgent:        "test_agent",
			IP:               netip.MustParseAddr("0.0.0.0"),
		},
	)

	suite.Require().NoError(err)

	err = sessionRepo.RefreshActiveSession(
		context.Background(),
		repository.RefreshActiveSessionParams{
			UUID:                sessionUUID,
			RefreshTokenHash:    "refresh_hash_1",
			NewRefreshTokenHash: "refresh_hash_2",
			IP:                  netip.MustParseAddr("0.0.0.0"),
			RefreshTTL:          time.Hour * 100,
		},
	)

	suite.Require().NoError(err)

	err = sessionRepo.RefreshActiveSession(
		context.Background(),
		repository.RefreshActiveSessionParams{
			UUID:                sessionUUID,
			RefreshTokenHash:    "refresh_hash_1",
			NewRefreshTokenHash: "refresh_hash_3",
			IP:                  netip.MustParseAddr("0.0.0.0"),
			RefreshTTL:          time.Hour * 100,
		},
	)

	suite.ErrorIs(
		err, repository.ErrSessionNotFound,
		"subsequent refreshes with the same token hash should fail",
	)
}

func (suite *RepositoryTestSuite) TestSessionRepository_RefreshSession_Revoked() {
	user, err := createTestUser(suite.T(), suite.DB)
	suite.Require().Nil(err)

	sessionRepo := sqliterepo.NewSessionsRepository(suite.DB)
	sessionUUID := uuid.New()

	_, err = sessionRepo.CreateSession(
		context.Background(),
		repository.CreateSessionParams{
			UUID:             sessionUUID,
			UserUUID:         user.UUID,
			RefreshTokenHash: "refresh_hash_1",
			UserAgent:        "test_agent",
			IP:               netip.MustParseAddr("0.0.0.0"),
		},
	)

	suite.Require().NoError(err)

	err = sessionRepo.RevokeSession(
		context.Background(),
		repository.RevokeSessionParams{
			UUID:     sessionUUID,
			UserUUID: user.UUID,
		},
	)

	suite.Require().NoError(err)

	err = sessionRepo.RefreshActiveSession(
		context.Background(),
		repository.RefreshActiveSessionParams{
			UUID:                sessionUUID,
			RefreshTokenHash:    "refresh_hash_1",
			NewRefreshTokenHash: "refresh_hash_2",
			IP:                  netip.MustParseAddr("0.0.0.0"),
			RefreshTTL:          time.Hour * 100,
		},
	)

	suite.ErrorIs(
		err, repository.ErrSessionNotFound,
		"revoked session should not be refreshable",
	)
}

func (suite *RepositoryTestSuite) TestSessionRepository_RefreshSession_Expired() {
	user, err := createTestUser(suite.T(), suite.DB)
	suite.Require().Nil(err)

	sessionRepo := sqliterepo.NewSessionsRepository(suite.DB)
	sessionUUID := uuid.New()

	_, err = sessionRepo.CreateSession(
		context.Background(),
		repository.CreateSessionParams{
			UUID:             sessionUUID,
			UserUUID:         user.UUID,
			RefreshTokenHash: "refresh_hash_1",
			UserAgent:        "test_agent",
			IP:               netip.MustParseAddr("0.0.0.0"),
		},
	)

	suite.Require().NoError(err)

	time.Sleep(1100 * time.Millisecond)

	err = sessionRepo.RefreshActiveSession(
		context.Background(),
		repository.RefreshActiveSessionParams{
			UUID:                sessionUUID,
			RefreshTokenHash:    "refresh_hash_1",
			NewRefreshTokenHash: "refresh_hash_2",
			IP:                  netip.MustParseAddr("0.0.0.0"),
			RefreshTTL:          time.Second * 1,
		},
	)

	suite.ErrorIs(
		err, repository.ErrSessionNotFound,
		"expired session should not be refreshable",
	)
}
