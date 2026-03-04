package middleware

import "net/http"

func noop(next http.Handler) http.Handler {
	return next
}
