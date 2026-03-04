package middleware

import "net/http"

// UseIf excludes the middleware if the condition is not met.
func UseIf(
	cond bool,
	m func(http.Handler) http.Handler,
) func(handler http.Handler) http.Handler {
	if cond {
		return m
	}

	return noop
}
