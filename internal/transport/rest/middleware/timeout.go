package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"saviour/internal/transport/rest/api"
)

func Timeout( //nolint:gocognit
	log *slog.Logger,
	timeout time.Duration,
) func(http.Handler) http.Handler {
	if timeout == 0 {
		return noop
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			tw := newTimeoutRW(w)
			done := make(chan any, 1)

			go func() {
				defer func() {
					if p := recover(); p != nil {
						done <- p
					}
				}()

				next.ServeHTTP(tw, r.WithContext(ctx))
				close(done)
			}()

			select {
			case p, panicked := <-done:
				if panicked {
					panic(p)
				}

			case <-ctx.Done():
				tw.Close()

				log.ErrorContext(
					ctx,
					"request timeout",
					slog.String("method", r.Method),
					slog.String("url", r.URL.String()),
				)

				if tw.HasStartedWriting() {
					return
				}

				api.WriteErrorResponse(
					log, w, r,
					api.NewErrorResponse(
						http.StatusGatewayTimeout,
						api.TimeoutErrorType,
						"Response took too long, canceled",
					),
				)
			}
		})
	}
}

type timeoutRW struct {
	http.ResponseWriter

	mu             sync.Mutex
	startedWriting bool
	closed         bool
}

func newTimeoutRW(w http.ResponseWriter) *timeoutRW {
	return &timeoutRW{ResponseWriter: w}
}

func (w *timeoutRW) Close() {
	w.mu.Lock()
	w.closed = true
	w.mu.Unlock()
}

func (w *timeoutRW) HasStartedWriting() bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.startedWriting
}

func (w *timeoutRW) WriteHeader(status int) {
	w.mu.Lock()

	if w.closed {
		w.mu.Unlock()
		return
	}

	w.startedWriting = true

	w.mu.Unlock()

	w.ResponseWriter.WriteHeader(status)
}

func (w *timeoutRW) Write(p []byte) (int, error) {
	w.mu.Lock()

	if w.closed {
		w.mu.Unlock()
		// The response has already been timed out.
		// Silently discard the write to avoid double writes and duplicate logging.
		return len(p), nil
	}

	w.startedWriting = true

	w.mu.Unlock()

	return w.ResponseWriter.Write(p)
}

func (w *timeoutRW) Flush() {
	w.mu.Lock()

	if w.closed {
		w.mu.Unlock()

		return
	}

	w.mu.Unlock()

	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

var _ http.Flusher = new(timeoutRW)
