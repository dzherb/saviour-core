package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"saviour/internal/service/auth"
)

func TestAccessToken_IssueAndValidate(t *testing.T) {
	t.Parallel()

	secret := "access-secret"
	userID := uuid.New()
	sessionID := uuid.New()
	ttl := time.Minute

	token, err := auth.IssueToken(
		secret,
		userID,
		sessionID,
		nil,
		ttl,
	)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	parsed, err := auth.ValidateToken(token, secret)
	require.NoError(t, err)
	require.NotNil(t, parsed)

	require.Equal(t, userID, parsed.UserUUID)
	require.Equal(t, sessionID, parsed.SessionUUID)
}

func TestAccessToken_ContainsRoles(t *testing.T) {
	t.Parallel()

	secret := "access-secret"
	userID := uuid.New()
	sessionID := uuid.New()
	ttl := time.Minute

	token, err := auth.IssueToken(
		secret,
		userID,
		sessionID,
		[]auth.Role{auth.AdminRole, "someone"},
		ttl,
	)
	require.NoError(t, err)

	parsed, err := auth.ValidateToken(token, secret)
	require.NoError(t, err)

	require.Contains(t, parsed.Roles, auth.AdminRole)
	require.Contains(t, parsed.Roles, auth.Role("someone"))
}

func TestToken_InvalidSecret(t *testing.T) {
	t.Parallel()

	token, err := auth.IssueToken(
		"correct-secret",
		uuid.New(),
		uuid.New(),
		nil,
		time.Minute,
	)
	require.NoError(t, err)

	_, err = auth.ValidateToken(
		token,
		"wrong-secret",
	)
	require.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestAccessToken_Expired(t *testing.T) {
	t.Parallel()

	secret := "secret-123"

	token, err := auth.IssueToken(
		secret,
		uuid.New(),
		uuid.New(),
		nil,
		-time.Second,
	)
	require.NoError(t, err)

	_, err = auth.ValidateToken(token, secret)
	require.ErrorIs(t, err, auth.ErrTokenExpired)
}

func TestToken_InvalidSID(t *testing.T) {
	t.Parallel()

	secret := "secret-321"
	now := time.Now()

	claims := auth.SessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.New().String(),
			Issuer:    auth.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
		},
		SessionID: "not-a-uuid",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	_, err = auth.ValidateToken(signed, secret)
	require.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestToken_InvalidSubject(t *testing.T) {
	t.Parallel()

	secret := "secret-000"
	now := time.Now()

	claims := auth.SessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "not-a-uuid",
			Issuer:    auth.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
		},
		SessionID: uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	_, err = auth.ValidateToken(signed, secret)
	require.ErrorIs(t, err, auth.ErrInvalidToken)
}
