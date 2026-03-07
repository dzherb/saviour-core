package secret

import (
	"fmt"
	"log/slog"
)

const repr = "*****"

// Secret represents a value with sensitive information
// that must never be displayed as is (e.g. in panics or logs).
type Secret[T any] struct {
	v T
}

func New[T any](secret T) Secret[T] {
	return Secret[T]{v: secret}
}

func (s Secret[T]) UnsafeValue() T {
	return s.v
}

var _ fmt.Stringer = new(Secret[struct{}])

func (s Secret[T]) String() string {
	return repr
}

var _ fmt.GoStringer = new(Secret[struct{}])

func (s Secret[T]) GoString() string {
	return repr
}

var _ fmt.Formatter = new(Secret[struct{}])

func (s Secret[T]) Format(f fmt.State, _ rune) {
	_, _ = f.Write([]byte(repr))
}

var _ slog.LogValuer = new(Secret[struct{}])

func (s Secret[T]) LogValue() slog.Value {
	return slog.StringValue(repr)
}
