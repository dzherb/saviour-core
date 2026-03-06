package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"saviour/internal/logger"
	"saviour/internal/transport/rest/api"
	"saviour/internal/transport/rest/middleware"
)

func TestOAPIValidator_ValidationPasses(t *testing.T) {
	t.Parallel()

	spec, err := api.GetSwagger()
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodGet,
		"/ping",
		nil,
	)

	h := middleware.OAPIValidator(logger.Noop, spec)(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("all ok"))
			},
		),
	)

	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "all ok")
}

func TestOAPIValidator_ValidationFails(t *testing.T) {
	t.Parallel()

	spec, err := api.GetSwagger()
	require.NoError(t, err)

	cases := []struct {
		path              string
		method            string
		body              []byte
		expectedStatus    int
		expectedToContain string
	}{
		{
			path:              "/nonexistent",
			method:            http.MethodGet,
			expectedStatus:    http.StatusNotFound,
			expectedToContain: string(api.ResourceDoesNotExistErrorType),
		},
		{
			path:              "/ping",
			method:            http.MethodPost,
			expectedStatus:    http.StatusMethodNotAllowed,
			expectedToContain: string(api.MethodNotAllowedErrorType),
		},
		{
			path:              "/auth/sessions",
			method:            http.MethodPost,
			body:              []byte(`{"wrong_field": true}`),
			expectedStatus:    http.StatusBadRequest,
			expectedToContain: string(api.ValidationFailedErrorType),
		},
	}

	for _, c := range cases {
		t.Run(c.method+" "+c.path, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(
				c.method,
				c.path,
				bytes.NewReader(c.body),
			)

			h := middleware.OAPIValidator(logger.Noop, spec)(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						_, _ = w.Write([]byte("all ok"))
					},
				),
			)

			h.ServeHTTP(w, r)

			assert.Equal(t, c.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), c.expectedToContain)
			assert.NotContains(t, w.Body.String(), "all ok")
		})
	}
}

func TestUrlRegexp(t *testing.T) {
	t.Parallel()

	urlRegexp := regexp.MustCompile(middleware.URLRegexp)

	tests := []struct {
		name  string
		url   string
		valid bool
	}{
		{"http simple", "http://example.com", true},
		{"https simple", "https://example.com", true},
		{"with www", "https://www.example.com", true},
		{"with path", "https://example.com/path/to/resource", true},
		{"with query", "https://example.com/search?q=test&x=1", true},
		{"with fragment", "https://example.com/path#section", true},
		{"subdomain", "https://api.dev.example.io", true},
		{"port and path", "http://example.com:8080/api/v1", true},
		{
			"complex url",
			"https://user:pass@example-domain.com/path?x=1&y=2#top",
			true,
		},

		{"no scheme", "example.com", false},
		{"ftp scheme", "ftp://example.com", false},
		{"empty", "", false},
		{"spaces", "https://example.com/with space", false},
		{"invalid domain", "https://example", false},
		{"double scheme", "https://https://example.com", false},
		{"leading text", "text https://example.com", false},
		{"trailing text", "https://example.com text", false},
		{"invalid tld", "https://example.abcdefg", false}, // > 6 chars
		{"only scheme", "https://", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := urlRegexp.MatchString(tt.url)
			assert.Equal(t, tt.valid, result)
		})
	}
}
