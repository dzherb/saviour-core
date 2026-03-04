package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type loggingRW struct {
	http.ResponseWriter
	status int
	bytes  int
	wrote  bool
}

func (w *loggingRW) WriteHeader(status int) {
	if w.wrote {
		return
	}

	w.status = status
	w.wrote = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingRW) Write(p []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}

	n, err := w.ResponseWriter.Write(p)
	w.bytes += n

	return n, err
}

func DebugLogProcessedRequests(
	log *slog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lw := &loggingRW{ResponseWriter: w}

			start := time.Now()

			defer func() {
				log.DebugContext(
					r.Context(),
					"request processed",
					slog.String("method", r.Method),
					slog.String("url", r.URL.String()),
					slog.Int("status", lw.status),
					slog.Int("bytes_written", lw.bytes),
					slog.Duration("duration", time.Since(start)),
				)
			}()

			next.ServeHTTP(lw, r)
		})
	}
}
