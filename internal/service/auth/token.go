package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const Issuer = "saviour.core"

type TokenParsed struct {
	UserUUID    uuid.UUID
	SessionUUID uuid.UUID
}

type SessionClaims struct {
	jwt.RegisteredClaims
	SID string `json:"sid"`
}

func issueToken(
	secret string,
	userUUID uuid.UUID,
	sessionUUID uuid.UUID,
	ttl time.Duration,
) (string, error) {
	now := time.Now()

	// Access and refresh tokens currently use the same claims set.
	// This may change in the future.
	claims := SessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userUUID.String(),
			Issuer:    Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		SID: sessionUUID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

func validateToken(
	token string,
	secret string,
) (*TokenParsed, error) {
	var claims SessionClaims

	_, err := jwt.ParseWithClaims(
		token,
		&claims,
		func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		},
		jwt.WithIssuedAt(),
		jwt.WithIssuer(Issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}

		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	userUUID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("%w: parse sub: %w", ErrInvalidToken, err)
	}

	sessionUUID, err := uuid.Parse(claims.SID)
	if err != nil {
		return nil, fmt.Errorf("%w: parse sid: %w", ErrInvalidToken, err)
	}

	return &TokenParsed{
		UserUUID:    userUUID,
		SessionUUID: sessionUUID,
	}, nil
}
