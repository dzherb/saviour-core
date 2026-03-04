package middleware

import (
	"net/http"

	"saviour/internal/logger"
)

func WithTraceIDCtx() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := logger.NewTraceID()

			ctx := logger.ContextWithTraceID(
				r.Context(),
				traceID,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
