package middleware_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"saviour/internal/logger"
	"saviour/internal/service/auth"
	"saviour/internal/testkit/testmock"
	"saviour/internal/transport/rest/api"
	"saviour/internal/transport/rest/middleware"
)

func TestAuthenticator_NotProtectedResourceIsAccessible(t *testing.T) {
	t.Parallel()

	mockValidator := testmock.NewMockTokenValidator(t)
	mw := middleware.Authenticator(
		logger.Noop,
		mockValidator,
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusTeapot, w.Code)
}

func TestAuthenticator_ValidToken(t *testing.T) {
	t.Parallel()

	userUUID := uuid.New()
	sessionUUID := uuid.New()

	mockValidator := testmock.NewMockTokenValidator(t)

	mockValidator.EXPECT().
		ValidateAccessToken("test-token").
		Return(
			&auth.TokenParsed{
				UserUUID:    userUUID,
				SessionUUID: sessionUUID,
			},
			nil,
		)

	mw := middleware.Authenticator(
		logger.Noop,
		mockValidator,
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	r.Header.Set("Authorization", "Bearer test-token")

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := auth.MustGetTokenFromContext(r.Context())

		require.Equal(t, token.UserUUID, userUUID)
		require.Equal(t, token.SessionUUID, sessionUUID)

		w.WriteHeader(http.StatusTeapot)
	}))

	h.ServeHTTP(
		w,
		r.WithContext(
			context.WithValue(
				r.Context(),
				middleware.BearerAuthSecurityKey,
				[]string{},
			),
		),
	)

	assert.Equal(t, http.StatusTeapot, w.Code)
}

func TestAuthenticator_UnknownSecurityIsSkipped(t *testing.T) {
	t.Parallel()

	mockValidator := testmock.NewMockTokenValidator(t)
	mw := middleware.Authenticator(
		logger.Noop,
		mockValidator,
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	h.ServeHTTP(
		w,
		r.WithContext(
			context.WithValue(
				r.Context(),
				middleware.SecurityCtxKey("TestAuth"),
				[]string{},
			),
		),
	)

	assert.Equal(t, http.StatusTeapot, w.Code)
}

func TestAuthenticator_NoAuthHeader(t *testing.T) {
	t.Parallel()

	mockValidator := testmock.NewMockTokenValidator(t)
	mw := middleware.Authenticator(
		logger.Noop,
		mockValidator,
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	h.ServeHTTP(
		w,
		r.WithContext(
			context.WithValue(
				r.Context(),
				middleware.BearerAuthSecurityKey,
				[]string{},
			),
		),
	)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthenticator_TokenExpired(t *testing.T) { //nolint:dupl
	t.Parallel()

	mockValidator := testmock.NewMockTokenValidator(t)

	mockValidator.EXPECT().
		ValidateAccessToken("test-token").
		Return(nil, auth.ErrTokenExpired)

	mw := middleware.Authenticator(
		logger.Noop,
		mockValidator,
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	r.Header.Set("Authorization", "Bearer test-token")

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	h.ServeHTTP(
		w,
		r.WithContext(
			context.WithValue(
				r.Context(),
				middleware.BearerAuthSecurityKey,
				[]string{},
			),
		),
	)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), string(api.AccessTokenExpiredErrorType))
}

func TestAuthenticator_TokenNotValid(t *testing.T) { //nolint:dupl
	t.Parallel()

	mockValidator := testmock.NewMockTokenValidator(t)

	mockValidator.EXPECT().
		ValidateAccessToken("test-token").
		Return(nil, auth.ErrInvalidToken)

	mw := middleware.Authenticator(
		logger.Noop,
		mockValidator,
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	r.Header.Set("Authorization", "Bearer test-token")

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	h.ServeHTTP(
		w,
		r.WithContext(
			context.WithValue(
				r.Context(),
				middleware.BearerAuthSecurityKey,
				[]string{},
			),
		),
	)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(
		t,
		w.Body.String(),
		string(api.AccessTokenNotValidErrorType),
	)
}

func TestAuthenticator_UnexpectedError(t *testing.T) {
	t.Parallel()

	mockValidator := testmock.NewMockTokenValidator(t)

	mockValidator.EXPECT().
		ValidateAccessToken("test-token").
		Return(nil, errors.New("test error"))

	mw := middleware.Authenticator(
		logger.Noop,
		mockValidator,
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	r.Header.Set("Authorization", "Bearer test-token")

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	h.ServeHTTP(
		w,
		r.WithContext(
			context.WithValue(
				r.Context(),
				middleware.BearerAuthSecurityKey,
				[]string{},
			),
		),
	)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), string(api.InternalErrorType))
}

func TestAuthenticator_UserRoles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		userRoles      []auth.Role
		securityScopes []string
		isAllowed      bool
	}{
		{
			[]auth.Role{auth.AdminRole},
			[]string{string(auth.AdminRole)},
			true,
		},
		{
			[]auth.Role{"role1", "role2"},
			[]string{"role1"},
			true,
		},
		{
			[]auth.Role{"role"},
			[]string{},
			true,
		},
		{
			[]auth.Role{"role"},
			nil,
			true,
		},
		{
			[]auth.Role{},
			[]string{"role"},
			false,
		},
		{
			nil,
			[]string{"role"},
			false,
		},
		{
			[]auth.Role{"role1"},
			[]string{"role2"},
			false,
		},
		{
			[]auth.Role{"role2"},
			[]string{"role2", "role3"},
			false,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
			t.Parallel()

			userUUID := uuid.New()
			sessionUUID := uuid.New()

			mockValidator := testmock.NewMockTokenValidator(t)

			mockValidator.EXPECT().
				ValidateAccessToken("test-token").
				Return(
					&auth.TokenParsed{
						UserUUID:    userUUID,
						SessionUUID: sessionUUID,
						Roles:       c.userRoles,
					},
					nil,
				)

			mw := middleware.Authenticator(
				logger.Noop,
				mockValidator,
			)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			r.Header.Set("Authorization", "Bearer test-token")

			h := mw(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					token := auth.MustGetTokenFromContext(r.Context())

					require.Equal(t, token.UserUUID, userUUID)
					require.Equal(t, token.SessionUUID, sessionUUID)

					w.WriteHeader(http.StatusTeapot)
				}),
			)

			h.ServeHTTP(
				w,
				r.WithContext(
					context.WithValue(
						r.Context(),
						middleware.BearerAuthSecurityKey,
						c.securityScopes,
					),
				),
			)

			if c.isAllowed {
				assert.Equal(t, http.StatusTeapot, w.Code)

				return
			}

			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}
