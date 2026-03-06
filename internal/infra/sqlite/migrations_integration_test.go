package sqlite_test

import (
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"

	"saviour/internal/testkit"
)

type MigrationsTestSuite struct {
	testkit.InMemorySQLiteSuite
}

func (suite *MigrationsTestSuite) TestStairway() {
	goose.SetBaseFS(testkit.SQLiteMigrationsFS(suite.T()))

	_ = goose.SetDialect("sqlite3")

	const migrationsDir = "."

	migrations, err := goose.CollectMigrations(
		migrationsDir,
		0,
		goose.MaxVersion,
	)
	suite.Require().NoError(err)

	for _, m := range migrations {
		err := goose.UpTo(suite.DB, migrationsDir, m.Version)
		suite.Require().NoError(err)

		// down by one
		err = goose.Down(suite.DB, migrationsDir)
		suite.Require().NoError(err)

		err = goose.UpTo(suite.DB, migrationsDir, m.Version)
		suite.Require().NoError(err)

		err = goose.Reset(suite.DB, migrationsDir)
		suite.Require().NoError(err)
	}
}

func TestMigrations(t *testing.T) {
	testkit.SkipIntegrationTestInShortMode(t)
	t.Parallel()

	suite.Run(t, new(MigrationsTestSuite))
}
