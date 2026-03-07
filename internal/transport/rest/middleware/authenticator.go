package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"saviour/internal/logger"
	"saviour/internal/service/auth"
	"saviour/internal/transport/rest/api"
)

const (
	AuthHeader = "Authorization"
	AuthPrefix = "Bearer "

	bearerAuthSecurityKey = securityCtxKey("BearerAuth")
)

type tokenValidator interface {
	ValidateAccessToken(token string) (*auth.TokenParsed, error)
}

// Authenticator handles requests authentication.
//
// It implicitly depends on OAPIValidator middleware
// to determine if a resource should be protected.
func Authenticator( //nolint:gocognit
	log *slog.Logger,
	tokenValidator tokenValidator,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requiredRoles, ok := isResourceProtected(r, bearerAuthSecurityKey)
			if !ok {
				next.ServeHTTP(w, r)

				return
			}

			header := r.Header.Get(AuthHeader)
			if header == "" {
				api.WriteErrorResponse(
					log, w, r,
					api.NewErrorResponse(
						http.StatusUnauthorized,
						api.AccessTokenNotValidErrorType,
						"Authorization header not provided",
					),
				)

				return
			}

			token := strings.TrimPrefix(header, AuthPrefix)

			tokenParsed, err := tokenValidator.ValidateAccessToken(token)
			if err != nil {
				handleTokenValidationError(log, w, r, err)

				return
			}

			for _, requiredRole := range requiredRoles {
				if !slices.Contains(tokenParsed.Roles, requiredRole) {
					api.WriteErrorResponse(
						log, w, r,
						api.NewErrorResponse(
							http.StatusForbidden,
							api.AccessForbiddenErrorType,
							fmt.Sprintf(
								`User role "%s" is required`,
								requiredRole,
							),
						),
					)

					return
				}
			}

			authCtx := auth.ContextWithToken(r.Context(), tokenParsed)

			next.ServeHTTP(w, r.WithContext(authCtx))
		})
	}
}

func handleTokenValidationError(
	log *slog.Logger,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	if errors.Is(err, auth.ErrTokenExpired) {
		api.WriteErrorResponse(
			log, w, r,
			api.NewErrorResponse(
				http.StatusUnauthorized,
				api.AccessTokenExpiredErrorType,
				"Access token expired, try to refresh",
			),
		)

		return
	}

	if errors.Is(err, auth.ErrInvalidToken) {
		api.WriteErrorResponse(
			log, w, r,
			api.NewErrorResponse(
				http.StatusUnauthorized,
				api.AccessTokenNotValidErrorType,
				"Access token not valid",
			),
		)

		return
	}

	log.ErrorContext(
		r.Context(),
		"unexpected error validating a token",
		logger.ErrorAttr(err),
	)

	api.WriteErrorResponse(
		log, w, r,
		api.DefaultInternalError,
	)
}

func isResourceProtected(
	request *http.Request,
	security securityCtxKey,
) ([]auth.Role, bool) {
	roles, ok := request.Context().Value(security).([]string)

	res := make([]auth.Role, 0, len(roles))
	for _, role := range roles {
		res = append(res, auth.Role(role))
	}

	// probably should cache per path?
	return res, ok
}
