package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/netip"
	"time"

	"github.com/google/uuid"

	"saviour/internal/model"
	"saviour/internal/repository"
	"saviour/pkg/secret"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionNotFound    = errors.New("session not found")
)

const (
	AccessTokenTTL  = 10 * time.Minute
	RefreshTokenTTL = 24 * 30 * time.Hour
)

type Config struct {
	AccessTokenSecret  secret.Secret
	RefreshTokenSecret secret.Secret
}

type Auth struct {
	log         *slog.Logger
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository

	cfg Config
}

func New(
	log *slog.Logger,
	userRepo repository.UserRepository,
	sessionsRepo repository.SessionRepository,
	cfg Config,
) *Auth {
	return &Auth{
		log:         log,
		userRepo:    userRepo,
		sessionRepo: sessionsRepo,
		cfg:         cfg,
	}
}

type CreateSessionParams struct {
	Username  string
	Password  string
	UserAgent string
	IP        netip.Addr
}

const fakeHash = "$2a$10$7EqJtq98hPqEX7fNZaFWoO7EqJtq98hPqEX7fNZaFWoO7EqJtq98hPq" //nolint:lll

func (a *Auth) CreateSession(
	ctx context.Context,
	params CreateSessionParams,
) (*model.TokenPair, error) {
	user, err := a.userRepo.GetUserByUsername(
		ctx,
		params.Username,
	)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			_ = CheckPassword(params.Password, fakeHash)

			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	err = CheckPassword(params.Password, user.PasswordHash)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	sessionUUID := uuid.Must(uuid.NewV7())

	tokenPair, err := a.issueTokenPair(
		user.UUID,
		sessionUUID,
		a.userRoles(user),
	)
	if err != nil {
		return nil, err
	}

	_, err = a.sessionRepo.CreateSession(
		ctx,
		repository.CreateSessionParams{
			UUID:             sessionUUID,
			UserUUID:         user.UUID,
			RefreshTokenHash: hashRefreshToken(tokenPair.RefreshToken),
			UserAgent:        params.UserAgent,
			IP:               params.IP,
		},
	)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

type RefreshSessionParams struct {
	RefreshToken string
	UserAgent    string
	IP           netip.Addr
}

func (a *Auth) RefreshSession(
	ctx context.Context,
	params RefreshSessionParams,
) (*model.TokenPair, error) {
	token, err := a.ValidateRefreshToken(params.RefreshToken)
	if err != nil {
		return nil, err
	}

	tokenPair, err := a.issueTokenPair(
		token.UserUUID,
		token.SessionUUID,
		token.Roles,
	)
	if err != nil {
		return nil, err
	}

	err = a.sessionRepo.RefreshActiveSession(
		ctx,
		repository.RefreshActiveSessionParams{
			UUID:                token.SessionUUID,
			RefreshTokenHash:    hashRefreshToken(params.RefreshToken),
			NewRefreshTokenHash: hashRefreshToken(tokenPair.RefreshToken),
			IP:                  params.IP,
			RefreshTTL:          RefreshTokenTTL,
		},
	)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return nil, ErrSessionNotFound
		}

		return nil, err
	}

	a.log.InfoContext(
		ctx,
		"session refreshed",
		slog.String("user_id", token.UserUUID.String()),
		slog.String("session_id", token.SessionUUID.String()),
		slog.String("ip", params.IP.String()),
		slog.String("user_agent", params.UserAgent),
	)

	return tokenPair, nil
}

type RevokeSessionParams struct {
	UserUUID    uuid.UUID
	SessionUUID uuid.UUID
}

func (a *Auth) RevokeSession(
	ctx context.Context,
	params RevokeSessionParams,
) error {
	err := a.sessionRepo.RevokeSession(
		ctx,
		repository.RevokeSessionParams{
			UUID:     params.SessionUUID,
			UserUUID: params.UserUUID,
		},
	)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return ErrSessionNotFound
		}
	}

	return nil
}

func (a *Auth) ValidateAccessToken(token string) (*TokenParsed, error) {
	return validateToken(token, a.cfg.AccessTokenSecret.UnsafeString())
}

func (a *Auth) ValidateRefreshToken(token string) (*TokenParsed, error) {
	return validateToken(token, a.cfg.RefreshTokenSecret.UnsafeString())
}

func (a *Auth) issueTokenPair(
	userUUID, sessionUUID uuid.UUID,
	roles []Role,
) (*model.TokenPair, error) {
	accessToken, err := issueToken(
		a.cfg.AccessTokenSecret.UnsafeString(),
		userUUID,
		sessionUUID,
		roles,
		AccessTokenTTL,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := issueToken(
		a.cfg.RefreshTokenSecret.UnsafeString(),
		userUUID,
		sessionUUID,
		roles,
		RefreshTokenTTL,
	)
	if err != nil {
		return nil, err
	}

	return &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *Auth) userRoles(user *model.User) []Role {
	if user.IsAdmin {
		return []Role{AdminRole}
	}

	return nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
