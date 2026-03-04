package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"saviour/internal/logger"
	"saviour/internal/transport/rest/middleware"
)

func TestWithTraceIDCtx(t *testing.T) {
	t.Parallel()

	traceSet := map[uuid.UUID]struct{}{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		traceID := logger.TraceIDFromCtx(ctx)

		assert.NotNil(t, traceID)

		traceSet[traceID] = struct{}{}
	})

	wrappedHandler := middleware.WithTraceIDCtx()(handler)

	wrappedHandler.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/resource1", nil),
	)

	wrappedHandler.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/resource2", nil),
	)

	assert.Len(t, traceSet, 2)
}
