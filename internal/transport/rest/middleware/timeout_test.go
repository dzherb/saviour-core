package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"saviour/internal/logger"
	"saviour/internal/transport/rest/api"
	"saviour/internal/transport/rest/middleware"
)

func TestTimeout_NoTimeout(t *testing.T) {
	t.Parallel()

	h := middleware.Timeout(logger.Noop, 0)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, ew := w.Write([]byte("ok"))

			assert.NoError(t, ew)
		}),
	)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "ok", w.Body.String())
}

func TestTimeout_WritesBeforeTimeout(t *testing.T) {
	t.Parallel()

	h := middleware.Timeout(logger.Noop, 50*time.Millisecond)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, ew := w.Write([]byte("ok"))

			assert.NoError(t, ew)
		}),
	)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "ok", w.Body.String())
}

func TestTimeout_WritesAfterTimeout(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})

	h := middleware.Timeout(logger.Noop, 10*time.Millisecond)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()

			time.Sleep(10 * time.Millisecond)

			w.WriteHeader(http.StatusTeapot)
			_, err := w.Write([]byte("nope"))

			assert.NoError(t, err)

			close(done)
		}),
	)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		assert.FailNow(t, "timed out waiting for handler")
	}

	require.Equal(t, http.StatusGatewayTimeout, w.Code)
	require.Contains(t, w.Body.String(), api.TimeoutErrorType)
	require.NotContains(t, w.Body.String(), "nope")
}

func TestTimeout_WritesBeforeAndAfterTimeout(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})

	h := middleware.Timeout(logger.Noop, 10*time.Millisecond)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("before"))

			assert.NoError(t, err)

			<-r.Context().Done()

			time.Sleep(10 * time.Millisecond)

			_, err = w.Write([]byte("after"))
			assert.NoError(t, err)

			close(done)
		}),
	)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		assert.FailNow(t, "timed out waiting for handler")
	}

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "before", w.Body.String())
}

func TestTimeout_PanicBeforeTimeout(t *testing.T) {
	t.Parallel()

	h := middleware.Timeout(logger.Noop, 50*time.Millisecond)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, ew := w.Write([]byte("ok"))

			assert.NoError(t, ew)

			panic("oops")
		}),
	)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	assert.PanicsWithValue(t, "oops", func() {
		h.ServeHTTP(w, r)
	})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestTimeout_PanicAfterTimeout(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})

	h := middleware.Timeout(logger.Noop, 10*time.Millisecond)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer close(done)

			<-r.Context().Done()

			time.Sleep(10 * time.Millisecond)

			panic("oops")
		}),
	)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		h.ServeHTTP(w, r)
	})

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		assert.FailNow(t, "timed out waiting for handler")
	}

	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	assert.Contains(t, w.Body.String(), api.TimeoutErrorType)
}
