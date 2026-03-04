package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"saviour/internal/logger"
)

type ErrorType string

const (
	InternalErrorType             ErrorType = "INTERNAL_ERROR"
	TimeoutErrorType              ErrorType = "RESPONSE_TIMEOUT"
	ResourceDoesNotExistErrorType ErrorType = "RESOURCE_DOES_NOT_EXIST"
	EntityNotFoundErrorType       ErrorType = "ENTITY_NOT_FOUND"
	MethodNotAllowedErrorType     ErrorType = "METHOD_NOT_ALLOWED"
	ValidationFailedErrorType     ErrorType = "VALIDATION_FAILED"
)

type errorResponse interface {
	error
	StatusCode() int32
}

var _ errorResponse = new(ErrorResponse)

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorBody.Type, e.ErrorBody.Message)
}

func (e *ErrorResponse) StatusCode() int32 {
	return e.ErrorBody.StatusCode
}

func NewErrorResponse(
	status int32,
	errType ErrorType,
	msg string,
) *ErrorResponse {
	return &ErrorResponse{ErrorBody: struct {
		Type       string `json:"type"`
		StatusCode int32  `json:"status_code"`
		Message    string `json:"message,omitempty"`
	}{
		Message:    msg,
		StatusCode: status,
		Type:       string(errType)},
	}
}

var DefaultInternalError = NewErrorResponse( //nolint:errname
	http.StatusInternalServerError,
	InternalErrorType,
	"Something went wrong",
)

type ErrorHandler struct {
	log *slog.Logger
}

func NewErrorHandler(log *slog.Logger) *ErrorHandler {
	return &ErrorHandler{log: log}
}

func (h *ErrorHandler) HandleRequestError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	h.log.ErrorContext(
		r.Context(),
		"error handling request data",
		logger.ErrorAttr(err),
		slog.String("url", r.URL.String()),
		slog.String("method", r.Method),
	)

	WriteErrorResponse(
		h.log,
		w, r,
		DefaultInternalError,
	)
}

func (h *ErrorHandler) HandleResponseError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	var apiErr errorResponse

	if errors.As(err, &apiErr) {
		WriteErrorResponse(h.log, w, r, apiErr)

		return
	}

	h.log.ErrorContext(
		r.Context(),
		"handler returned an error",
		slog.String("error", fmt.Sprintf("%s", err)),
		slog.String("url", r.URL.String()),
		slog.String("method", r.Method),
	)

	WriteErrorResponse(
		h.log, w, r,
		DefaultInternalError,
	)
}

// WriteErrorResponse is a convenience wrapper over WriteResponse
// for writing API errors.
func WriteErrorResponse(
	log *slog.Logger,
	w http.ResponseWriter,
	r *http.Request,
	resp errorResponse,
) {
	WriteResponse(
		log, w, r,
		resp, resp.StatusCode(),
	)
}
