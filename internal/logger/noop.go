package logger

import (
	"log/slog"
)

var Noop = slog.New(slog.DiscardHandler)
