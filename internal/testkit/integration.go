package testkit

import (
	"testing"
)

func SkipIntegrationTestInShortMode(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
}
