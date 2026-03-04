package api_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"saviour/internal/testkit"
	"saviour/internal/transport/rest/api"
)

func TestWriteResponse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		obj          any
		expectedBody string
		expectErrLog bool
	}{
		{
			name: "valid obj",
			obj: map[string]string{
				"test": "ok",
			},
			expectedBody: "{\"test\":\"ok\"}\n",
		},
		{
			name:         "nil obj",
			obj:          nil,
			expectedBody: "",
		},
		{
			name: "invalid obj",
			obj: map[string]any{
				"not_serializable": make(chan struct{}),
			},
			expectedBody: "",
			expectErrLog: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			h := testkit.NewSlogHandlerMock()
			log := slog.New(h)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			api.WriteResponse(
				log,
				w, r,
				c.obj,
				http.StatusTeapot,
			)

			assert.Equal(t, http.StatusTeapot, w.Code)
			assert.Equal(t, c.expectedBody, w.Body.String())

			if c.expectErrLog {
				require.Len(
					t,
					h.Records(),
					1,
					"should have written an error log",
				)
				assert.Equal(t, slog.LevelError, h.Records()[0].Level)

				return
			}

			assert.Len(t, h.Records(), 0, "should log no errors")
		})
	}
}
