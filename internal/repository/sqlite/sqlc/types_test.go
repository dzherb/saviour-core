package sqlc_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"saviour/internal/repository/sqlite/sqlc"
)

func TestUnixTimeScanInt64(t *testing.T) {
	t.Parallel()

	var ut sqlc.UnixTime

	err := ut.Scan(int64(1700000000))
	require.NoError(t, err)

	require.Equal(
		t,
		time.Unix(1700000000, 0).UTC(),
		ut.Time(),
	)
}

func TestUnixTimeScanNil(t *testing.T) {
	t.Parallel()

	var ut sqlc.UnixTime

	err := ut.Scan(nil)
	require.NoError(t, err)

	require.True(t, ut.Time().IsZero())
}

func TestUnixTimeScanInvalidType(t *testing.T) {
	t.Parallel()

	var ut sqlc.UnixTime

	err := ut.Scan("invalid")

	require.Error(t, err)
}

func TestUnixTimeValue(t *testing.T) {
	t.Parallel()

	tm := time.Unix(1700000000, 0).UTC()
	ut := sqlc.UnixTime(tm)

	v, err := ut.Value()

	require.NoError(t, err)
	require.Equal(t, int64(1700000000), v)
}

func TestUnixTimeValueZero(t *testing.T) {
	t.Parallel()

	var ut sqlc.UnixTime

	v, err := ut.Value()

	require.NoError(t, err)
	require.Nil(t, v)
}

func TestUnixTimeTimeMethod(t *testing.T) {
	t.Parallel()

	tm := time.Unix(1700000000, 0).UTC()
	ut := sqlc.UnixTime(tm)

	require.Equal(t, tm, ut.Time())
}
