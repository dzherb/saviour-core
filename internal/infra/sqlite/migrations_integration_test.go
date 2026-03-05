package sqlite_test

import (
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"

	"saviour/internal/infra/sqlite"
	"saviour/internal/testkit"
)

func TestStairway(t *testing.T) {
	testkit.SkipIntegrationTestInShortMode(t)
	t.Parallel()

	goose.SetBaseFS(testkit.SQLiteMigrationsFS(t))

	_ = goose.SetDialect("sqlite3")

	const migrationsDir = "."

	migrations, err := goose.CollectMigrations(
		migrationsDir,
		0,
		goose.MaxVersion,
	)
	require.NoError(t, err)

	db, err := sqlite.NewDB(
		sqlite.Config{
			DSN: "file::memory:?_pragma=foreign_keys(1)",
		},
	)
	require.NoError(t, err)

	defer func() {
		_ = db.Close()
	}()

	for _, m := range migrations {
		err := goose.UpTo(db, migrationsDir, m.Version)
		require.NoError(t, err)

		// down by one
		err = goose.Down(db, migrationsDir)
		require.NoError(t, err)

		err = goose.UpTo(db, migrationsDir, m.Version)
		require.NoError(t, err)

		err = goose.Reset(db, migrationsDir)
		require.NoError(t, err)
	}
}
