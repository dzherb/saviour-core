package logger

import (
	"context"
	"log"
	"log/slog"
	"strings"
)

func SlogToStdLog(
	logger *slog.Logger,
	level slog.Level,
) *log.Logger {
	return log.New(
		&slogWriter{
			logger: logger,
			level:  level,
		},
		"",
		0,
	)
}

type slogWriter struct {
	logger *slog.Logger
	level  slog.Level
}

func (w *slogWriter) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(p), "\n")

	w.logger.Log(context.Background(), w.level, msg) //nolint:sloglint

	return len(p), nil
}
