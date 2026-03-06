package auth

import (
	"context"

	"saviour/internal/model"
)

type Service interface {
	CreateSession(
		ctx context.Context,
		params CreateSessionParams,
	) (*model.TokenPair, error)
	RefreshSession(
		ctx context.Context,
		params RefreshSessionParams,
	) (*model.TokenPair, error)
	RevokeSession(
		ctx context.Context,
		params RevokeSessionParams,
	) error
}

type TokenValidator interface {
	ValidateAccessToken(token string) (*TokenParsed, error)
	ValidateRefreshToken(token string) (*TokenParsed, error)
}
