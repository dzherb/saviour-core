package sqliterepo_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"saviour/internal/testkit"
)

type RepositoryTestSuite struct {
	testkit.InMemoryMigratedSQLiteSuite
}

func TestRepositories(t *testing.T) {
	testkit.SkipIntegrationTestInShortMode(t)
	t.Parallel()

	suite.Run(t, new(RepositoryTestSuite))
}
