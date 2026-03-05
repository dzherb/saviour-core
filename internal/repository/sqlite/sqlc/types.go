package sqlc

import (
	"database/sql/driver"
	"fmt"
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
		return nil, nil
	}

	return tt.Unix(), nil
}

func (t UnixTime) Time() time.Time {
	return time.Time(t)
}
