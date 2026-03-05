package sqlc

import (
	"database/sql/driver"
	"fmt"
	"net/netip"
	"time"
)

type UnixTime time.Time

func (t *UnixTime) Scan(v any) error {
	if v == nil {
		*t = UnixTime(time.Time{})
		return nil
	}

	ts, ok := v.(int64)
	if !ok {
		return fmt.Errorf("cannot scan %T into UnixTime", v)
	}

	*t = UnixTime(time.Unix(ts, 0).UTC())

	return nil
}

func (t UnixTime) Value() (driver.Value, error) {
	tt := time.Time(t)

	if tt.IsZero() {
		return nil, nil //nolint:nilnil
	}

	return tt.Unix(), nil
}

func (t UnixTime) Time() time.Time {
	return time.Time(t)
}

type IP netip.Addr

func (ip *IP) Scan(v any) error {
	b, ok := v.([]byte)
	if !ok {
		return fmt.Errorf("invalid type %T", v)
	}

	addr, ok := netip.AddrFromSlice(b)
	if !ok {
		return fmt.Errorf("invalid ip")
	}

	*ip = IP(addr)

	return nil
}

func (ip IP) Value() (driver.Value, error) {
	addr := netip.Addr(ip)
	b := addr.AsSlice()

	return b, nil
}

func (ip IP) Addr() netip.Addr {
	return netip.Addr(ip)
}
