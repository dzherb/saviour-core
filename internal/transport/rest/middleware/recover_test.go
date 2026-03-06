package middleware_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"saviour/internal/logger"
	"saviour/internal/testkit/testmock"
	"saviour/internal/transport/rest/middleware"
)

func TestRecover_BeforeWrite(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	h := testmock.NewMockSlogHandler()
	log := slog.New(h)

	wrappedHandler := middleware.Recover(log)(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	wrappedHandler.ServeHTTP(w, r)

	logHasPanicAttr := false

	h.Records()[0].Attrs(func(attr slog.Attr) bool {
		if attr.Key == "panic" {
			assert.Equal(t, "test panic", attr.Value.String())

			logHasPanicAttr = true

			return false
		}

		return true
	})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NotContains(
		t, "test panic", w.Body.String(),
		"response body should not contain panic details",
	)
	assert.True(t, logHasPanicAttr)
}

func TestRecover_AfterWrite(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, err := w.Write([]byte("teapot"))

		assert.NoError(t, err)

		panic("test panic")
	})

	wrappedHandler := middleware.Recover(logger.Noop)(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	wrappedHandler.ServeHTTP(w, r)

	assert.Equal(
		t, http.StatusTeapot, w.Code,
		"status code should not be changed if WriteHeader was called",
	)
	assert.Equal(t, []byte("teapot"), w.Body.Bytes())
}
