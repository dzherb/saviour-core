package logger_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"saviour/internal/logger"
	"saviour/internal/testkit/testmock"
)

func TestSlogToStdLog(t *testing.T) {
	t.Parallel()

	h := testmock.NewSlogHandlerMock()
	log := slog.New(h)

	std := logger.SlogToStdLog(log, slog.LevelWarn)
	std.Print("boom")

	require.Len(t, h.Records(), 1)
	require.Equal(t, slog.LevelWarn, h.Records()[0].Level)
	require.Equal(t, "boom", h.Records()[0].Message)
}
