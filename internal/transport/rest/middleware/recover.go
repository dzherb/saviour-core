package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"sync/atomic"

	"saviour/internal/transport/rest/api"
)

func Recover(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			aw := &writeAwareRW{ResponseWriter: w}

			defer func() {
				if rec := recover(); rec != nil {
					log.ErrorContext(
						r.Context(),
						"recovered from panic",
						slog.String("panic", fmt.Sprint(rec)),
						slog.String("stack", string(debug.Stack())),
						slog.String("url", r.URL.String()),
					)

					if !aw.wrote.Load() {
						api.WriteErrorResponse(
							log,
							w,
							r,
							api.DefaultInternalError,
						)
					}
				}
			}()

			next.ServeHTTP(aw, r)
		})
	}
}

type writeAwareRW struct {
	http.ResponseWriter
	wrote atomic.Bool
}

func (w *writeAwareRW) WriteHeader(status int) {
	w.wrote.Store(true)
	w.ResponseWriter.WriteHeader(status)
}

func (w *writeAwareRW) Write(p []byte) (int, error) {
	w.wrote.Store(true)
	return w.ResponseWriter.Write(p)
}
