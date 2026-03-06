package handler_test

import (
	"net/http"
	"net/netip"
	"testing"

	"saviour/internal/transport/rest/handler"
)

func TestClientIP(t *testing.T) {
	t.Parallel()

	trusted := []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/8"),
	}

	tests := []struct {
		name     string
		req      *http.Request
		expected netip.Addr
	}{
		{
			name: "direct client, no proxy",
			req: &http.Request{
				RemoteAddr: "1.2.3.4:1234",
				Header:     http.Header{},
			},
			expected: netip.MustParseAddr("1.2.3.4"),
		},
		{
			name: "trusted proxy with X-Forwarded-For",
			req: &http.Request{
				RemoteAddr: "10.1.2.3:5678",
				Header: http.Header{
					"X-Forwarded-For": []string{"8.8.8.8, 9.9.9.9"},
				},
			},
			expected: netip.MustParseAddr("8.8.8.8"),
		},
		{
			name: "trusted proxy with X-Real-IP",
			req: &http.Request{
				RemoteAddr: "10.1.2.3:5678",
				Header: http.Header{
					"X-Real-Ip": []string{"7.7.7.7"},
				},
			},
			expected: netip.MustParseAddr("7.7.7.7"),
		},
		{
			name: "trusted proxy with empty XFF uses X-Real-IP",
			req: &http.Request{
				RemoteAddr: "10.1.2.3:5678",
				Header: http.Header{
					"X-Forwarded-For": []string{""},
					"X-Real-Ip":       []string{"7.7.7.7"},
				},
			},
			expected: netip.MustParseAddr("7.7.7.7"),
		},
		{
			name: "untrusted proxy ignores headers",
			req: &http.Request{
				RemoteAddr: "5.5.5.5:1111",
				Header: http.Header{
					"X-Forwarded-For": []string{"8.8.8.8"},
				},
			},
			expected: netip.MustParseAddr("5.5.5.5"),
		},
		{
			name: "invalid forwarded ip falls back",
			req: &http.Request{
				RemoteAddr: "10.1.2.3:5678",
				Header: http.Header{
					"X-Forwarded-For": []string{"not-an-ip"},
				},
			},
			expected: netip.MustParseAddr("10.1.2.3"),
		},
		{
			name:     "nil request",
			req:      nil,
			expected: handler.FallbackIP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ip := handler.ClientIP(tt.req, trusted)
			if ip != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, ip)
			}
		})
	}
}
