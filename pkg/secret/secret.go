package secret

import (
	"fmt"
	"log/slog"
)

const repr = "*****"

// Secret represents a string with sensitive information
// that must never be displayed as is (e.g. in panics or logs).
type Secret struct {
	v string
}

func New(secret string) Secret {
	return Secret{v: secret}
}

func (s Secret) UnsafeString() string {
	return s.v
}

var _ fmt.Stringer = new(Secret)

func (s Secret) String() string {
	return repr
}

var _ fmt.GoStringer = new(Secret)

func (s Secret) GoString() string {
	return repr
}

var _ fmt.Formatter = new(Secret)

func (s Secret) Format(f fmt.State, verb rune) {
	_, _ = f.Write([]byte(repr))
}

var _ slog.LogValuer = new(Secret)

func (s Secret) LogValue() slog.Value {
	return slog.StringValue(repr)
}
