package api

import (
	"log/slog"
	"net/http"
)

func NewHTTPHandler(log *slog.Logger, hi HandlerInterface) http.Handler {
	errHandler := NewErrorHandler(log)

	return HandlerWithOptions(
		hi,
		Options{
			BaseRouter: http.NewServeMux(),
			// RequestErrorHandlerFunc should not to be called
			// if OAPIValidator middleware is used.
			// It's already meant to handle all the validation errors.
			RequestErrorHandlerFunc:  errHandler.HandleRequestError,
			ResponseErrorHandlerFunc: errHandler.HandleResponseError,
		},
	)
}
