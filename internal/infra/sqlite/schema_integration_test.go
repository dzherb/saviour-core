package sqlite_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"saviour/internal/testkit"
)

type SchemaTestSuite struct {
	testkit.InMemoryMigratedSQLiteSuite
}

func (suite *SchemaTestSuite) TestUpdatedAtTrigger() {
	row := suite.DB.QueryRow(
		`INSERT INTO users (uuid, username, password_hash)
		VALUES (?, ?, ?)
		RETURNING uuid, created_at, updated_at`,
		uuid.Must(uuid.NewV7()), "test_user", "password",
	)

	type userRow struct {
		uuid      uuid.UUID
		createdAt int64
		updatedAt int64
	}

	usr := userRow{}
	err := row.Scan(&usr.uuid, &usr.createdAt, &usr.updatedAt)

	suite.Require().NoError(err)
	suite.Require().NotEmpty(usr.updatedAt)
	suite.Require().Equal(usr.updatedAt, usr.createdAt)

	// we store time with seconds precision,
	// so sleep to catch an updated_at change...
	time.Sleep(1100 * time.Millisecond)

	_, err = suite.DB.Exec(
		"UPDATE users SET username = ? WHERE uuid = ?",
		"new_username", usr.uuid,
	)
	suite.Require().NoError(err)

	row = suite.DB.QueryRow(
		"SELECT uuid, updated_at FROM users WHERE uuid = ?",
		usr.uuid,
	)

	updatedUsr := userRow{}

	err = row.Scan(&updatedUsr.uuid, &updatedUsr.updatedAt)

	suite.Require().NoError(err)
	suite.Require().Greater(updatedUsr.updatedAt, usr.updatedAt)
}

func TestSchema(t *testing.T) {
	testkit.SkipIntegrationTestInShortMode(t)
	t.Parallel()

	suite.Run(t, new(SchemaTestSuite))
}
