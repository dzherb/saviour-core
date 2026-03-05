package sqliterepo

import (
	"context"

	"saviour/internal/repository"
	sqlitequery "saviour/internal/repository/sqlite/sqlc/gen"
)

type DBTX = sqlitequery.DBTX

type baseRepository struct {
	db DBTX
}

func (r *baseRepository) txAwareQueries(
	ctx context.Context,
) *sqlitequery.Queries {
	if tx, ok := repository.TxFromContext(ctx); ok {
		return r.queries().WithTx(tx)
	}

	return r.queries()
}

func (r *baseRepository) queries() *sqlitequery.Queries {
	return sqlitequery.New(r.db)
}

func newRepository(db DBTX) baseRepository {
	return baseRepository{
		db: db,
	}
}
