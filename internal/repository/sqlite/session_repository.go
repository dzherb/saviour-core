package sqliterepo

import (
	"context"

	"github.com/google/uuid"

	"saviour/internal/model"
	"saviour/internal/repository"
	"saviour/internal/repository/sqlite/sqlc"
	sqlitequery "saviour/internal/repository/sqlite/sqlc/gen"
)

type SessionsRepository struct {
	baseRepository
}

func NewSessionsRepository(db repository.DBTX) *SessionsRepository {
	return &SessionsRepository{
		newRepository(db),
	}
}

func (r *SessionsRepository) CreateSession(
	ctx context.Context,
	params repository.CreateSessionParams,
) (*model.Session, error) {
	if params.UUID == uuid.Nil {
		params.UUID = uuid.Must(uuid.NewV7())
	}

	session, err := r.txAwareQueries(ctx).
		CreateSession(
			ctx,
			sqlitequery.CreateSessionParams{
				UUID:             sqlc.UUID(params.UUID),
				UserUUID:         sqlc.UUID(params.UserUUID),
				RefreshTokenHash: params.RefreshTokenHash,
				UserAgent:        params.UserAgent,
				IP:               sqlc.IP(params.IP),
			},
		)
	if err != nil {
		return nil, err
	}

	return mapSession(&session), nil
}

func (r *SessionsRepository) RefreshActiveSession(
	ctx context.Context,
	params repository.RefreshActiveSessionParams,
) error {
	rowsAffected, err := r.txAwareQueries(ctx).
		RefreshActiveSession(
			ctx,
			sqlitequery.RefreshActiveSessionParams{
				UUID:                sqlc.UUID(params.UUID),
				RefreshTokenHash:    params.RefreshTokenHash,
				NewRefreshTokenHash: params.NewRefreshTokenHash,
				IP:                  sqlc.IP(params.IP),
				RefreshTTLInSec:     int64(params.RefreshTTL.Seconds()),
			},
		)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrSessionNotFound
	}

	return nil
}

func (r *SessionsRepository) RevokeSession(
	ctx context.Context,
	params repository.RevokeSessionParams,
) error {
	rowsAffected, err := r.txAwareQueries(ctx).
		RevokeSession(
			ctx,
			sqlitequery.RevokeSessionParams{
				UUID:     sqlc.UUID(params.UUID),
				UserUUID: sqlc.UUID(params.UserUUID),
			},
		)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrSessionNotFound
	}

	return nil
}

func mapSession(session *sqlitequery.Session) *model.Session {
	return &model.Session{
		UUID:          uuid.UUID(session.UUID),
		CreatedAt:     session.CreatedAt.Time(),
		UserUUID:      uuid.UUID(session.UserUUID),
		LastRefreshAt: session.LastRefreshAt.Time(),
		UserAgent:     session.UserAgent,
		IP:            session.IPLast.Addr(),
	}
}
