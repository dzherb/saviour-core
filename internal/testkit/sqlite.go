package testkit

import (
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	_ "modernc.org/sqlite"
)

func SQLiteMigrationsFS(t *testing.T) fs.FS {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to determine the current file path")
	}

	return os.DirFS(filepath.Join(
		filepath.Dir(currentFile),
		"../infra/sqlite/migrations",
	))
}

type InMemorySQLiteSuite struct {
	suite.Suite

	DB *sql.DB
}

var _ suite.SetupTestSuite = new(InMemorySQLiteSuite)
var _ suite.TearDownTestSuite = new(InMemorySQLiteSuite)

func (suite *InMemorySQLiteSuite) SetupTest() {
	db, err := sql.Open(
		"sqlite",
		"file::memory:?_pragma=foreign_keys(1)",
	)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	suite.DB = db
}

func (suite *InMemorySQLiteSuite) TearDownTest() {
	if suite.DB != nil {
		_ = suite.DB.Close()
	}

	suite.DB = nil
}

type InMemoryMigratedSQLiteSuite struct {
	InMemorySQLiteSuite
}

var _ suite.SetupTestSuite = new(InMemorySQLiteSuite)
var _ suite.TearDownTestSuite = new(InMemorySQLiteSuite)

func (suite *InMemoryMigratedSQLiteSuite) SetupTest() {
	suite.InMemorySQLiteSuite.SetupTest()

	goose.SetBaseFS(SQLiteMigrationsFS(suite.T()))

	_ = goose.SetDialect("sqlite3")

	err := goose.Up(suite.DB, ".")
	if err != nil {
		panic("failed to migrate: " + err.Error())
	}
}

func (suite *InMemoryMigratedSQLiteSuite) TearDownTest() {
	suite.InMemorySQLiteSuite.TearDownTest()
}
