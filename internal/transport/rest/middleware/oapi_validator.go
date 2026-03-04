package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/google/uuid"

	"saviour/internal/logger"
	"saviour/internal/transport/rest/api"
)

type securityCtxKey string

// OAPIValidator creates a middleware that validates
// requests against the provided openapi specification.
func OAPIValidator(
	log *slog.Logger,
	spec *openapi3.T,
) func(http.Handler) http.Handler {
	// Disable host validation
	spec.Servers = nil

	router, err := gorillamux.NewRouter(spec)
	if err != nil {
		panic(err)
	}

	registerExtraFormatValidators()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route, pathParams, err := router.FindRoute(r)
			if err != nil {
				handleOAPIError(log, w, r, err)

				return
			}

			ctxWithSecurity := r.Context()

			requestValidationInput := &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
				Options: &openapi3filter.Options{
					MultiError: false,
					AuthenticationFunc: func(
						_ context.Context,
						input *openapi3filter.AuthenticationInput,
					) error {
						// Do not handle authentication here,
						// only enrich the context
						ctxWithSecurity = context.WithValue( //nolint:fatcontext
							ctxWithSecurity,
							securityCtxKey(input.SecuritySchemeName),
							input.Scopes,
						)

						return nil
					},
				},
			}

			err = openapi3filter.ValidateRequest(
				r.Context(),
				requestValidationInput,
			)
			if err != nil {
				handleOAPIError(log, w, r, err)

				return
			}

			next.ServeHTTP(w, r.WithContext(ctxWithSecurity))
		})
	}
}

const URLRegexp = `^https?:\/\/(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*)$` //nolint:lll

var once sync.Once

func registerExtraFormatValidators() {
	once.Do(func() {
		openapi3.DefineStringFormatValidator(
			"uuid",
			openapi3.NewCallbackValidator(uuid.Validate),
		)
		openapi3.DefineStringFormatValidator(
			"uri",
			openapi3.NewRegexpFormatValidator(URLRegexp),
		)
	})
}

func handleOAPIError(
	log *slog.Logger,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	if errors.Is(err, routers.ErrMethodNotAllowed) {
		api.WriteErrorResponse(
			log, w, r,
			api.NewErrorResponse(
				http.StatusMethodNotAllowed,
				api.MethodNotAllowedErrorType,
				fmt.Sprintf(
					"Method '%s' not allowed for resource %s",
					r.Method, r.URL.Path,
				),
			),
		)

		return
	}

	if errors.Is(err, routers.ErrPathNotFound) {
		api.WriteErrorResponse(
			log, w, r,
			api.NewErrorResponse(
				http.StatusNotFound,
				api.ResourceDoesNotExistErrorType,
				fmt.Sprintf(
					"Resource %s does not exist",
					r.RequestURI,
				),
			),
		)

		return
	}

	var reqErr *openapi3filter.RequestError

	if errors.As(err, &reqErr) {
		api.WriteErrorResponse(
			log, w, r,
			api.NewErrorResponse(
				http.StatusBadRequest,
				api.ValidationFailedErrorType,
				// openapi errors seem to be multi-line
				// with a decent message on the first.
				strings.Split(reqErr.Error(), "\n")[0],
			),
		)

		return
	}

	// This should never happen today, but if the upstream code changes,
	// we don't want to crash the server.
	log.ErrorContext(
		r.Context(),
		"unexpected error from openapi3filter.ValidateRequest",
		logger.ErrorAttr(err),
	)

	api.WriteErrorResponse(log, w, r, api.DefaultInternalError)
}
