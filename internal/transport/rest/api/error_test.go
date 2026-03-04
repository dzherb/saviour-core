package api_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"saviour/internal/logger"
	"saviour/internal/transport/rest/api"
)

func TestErrorHandler(t *testing.T) {
	t.Parallel()

	cases := []struct {
		err            error
		expectedStatus int
		expectedBody   string
	}{
		{
			api.NewErrorResponse(http.StatusTeapot, "TEAPOT", "I am a teapot"),
			http.StatusTeapot,
			`{"error":{"type":"TEAPOT","status_code":418,"message":"I am a teapot"}}`,
		},
		{
			api.DefaultInternalError,
			http.StatusInternalServerError,
			`{"error":{"type":"INTERNAL_ERROR","status_code":500,"message":"Something went wrong"}}`,
		},
		{
			errors.New("not an ErrorResponse type"),
			http.StatusInternalServerError,
			`{"error":{"type":"INTERNAL_ERROR","status_code":500,"message":"Something went wrong"}}`,
		},
		{
			nil,
			http.StatusInternalServerError,
			`{"error":{"type":"INTERNAL_ERROR","status_code":500,"message":"Something went wrong"}}`,
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s", c.err), func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			errHandler := api.NewErrorHandler(logger.Noop)

			errHandler.HandleResponseError(w, r, c.err)

			assert.Equal(t, c.expectedStatus, w.Code)
			assert.JSONEq(
				t,
				c.expectedBody,
				w.Body.String(),
			)
		})
	}
}
