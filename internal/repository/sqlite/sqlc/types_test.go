package sqlc_test

import (
	"net/netip"
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

func TestIP_Scan_IPv4(t *testing.T) {
	t.Parallel()

	addr := netip.MustParseAddr("192.168.1.1")

	var ip sqlc.IP

	err := ip.Scan(addr.AsSlice())

	require.NoError(t, err)
	require.Equal(t, addr, netip.Addr(ip))
}

func TestIP_Scan_IPv6(t *testing.T) {
	t.Parallel()

	addr := netip.MustParseAddr("2001:db8::1")

	var ip sqlc.IP

	err := ip.Scan(addr.AsSlice())

	require.NoError(t, err)
	require.Equal(t, addr, netip.Addr(ip))
}

func TestIP_Scan_InvalidType(t *testing.T) {
	t.Parallel()

	var ip sqlc.IP

	err := ip.Scan("not bytes")

	require.Error(t, err)
}

func TestIP_Scan_InvalidBytes(t *testing.T) {
	t.Parallel()

	var ip sqlc.IP

	err := ip.Scan([]byte{1, 2, 3})

	require.Error(t, err)
}

func TestIP_Value(t *testing.T) {
	t.Parallel()

	addr := netip.MustParseAddr("10.0.0.1")

	ip := sqlc.IP(addr)

	val, err := ip.Value()

	require.NoError(t, err)

	b, ok := val.([]byte)
	require.True(t, ok)

	require.Equal(t, addr.AsSlice(), b)
}

func TestIP_RoundTrip(t *testing.T) {
	t.Parallel()

	addr := netip.MustParseAddr("2001:db8::1")

	ip := sqlc.IP(addr)

	val, err := ip.Value()
	require.NoError(t, err)

	var ip2 sqlc.IP

	err = ip2.Scan(val)

	require.NoError(t, err)
	require.Equal(t, addr, netip.Addr(ip2))
}
