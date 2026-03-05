package logger_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"saviour/internal/logger"
	"saviour/internal/testkit/testmock"
)

func TestTraceCtx(t *testing.T) {
	t.Parallel()

	assert.Equal(t, uuid.Nil, logger.TraceIDFromCtx(context.Background()))

	traceID := logger.NewTraceID()
	ctxWithTrace := logger.ContextWithTraceID(context.Background(), traceID)

	assert.Equal(t, traceID, logger.TraceIDFromCtx(ctxWithTrace))

	wrappedCtx := context.WithValue(
		ctxWithTrace,
		"test_key", //nolint:staticcheck
		"some_val",
	)

	assert.Equal(t, traceID, logger.TraceIDFromCtx(wrappedCtx))

	newTraceID := logger.NewTraceID()
	wrappedCtx = logger.ContextWithTraceID(wrappedCtx, newTraceID)

	assert.Equal(t, newTraceID, logger.TraceIDFromCtx(wrappedCtx))
}

func TestTraceContextHandler_AddsTraceID(t *testing.T) {
	t.Parallel()

	base := testmock.NewSlogHandlerMock()
	h := logger.NewTraceContextHandler(base)

	log := slog.New(h)

	traceID := uuid.New()
	ctx := logger.ContextWithTraceID(context.Background(), traceID)

	log.InfoContext(ctx, "hello")

	require.Len(t, base.Records(), 1)

	var got string

	base.Records()[0].Attrs(func(a slog.Attr) bool {
		if a.Key == logger.TraceIDLogKey {
			if got != "" {
				require.FailNow(t, "should have only one trace id attr")
			}

			got = a.Value.String()
		}

		return true
	})

	require.Equal(t, traceID.String(), got)
}

func TestTraceContextHandler_NoTraceID(t *testing.T) {
	t.Parallel()

	base := testmock.NewSlogHandlerMock()
	h := logger.NewTraceContextHandler(base)

	log := slog.New(h)

	log.InfoContext(context.Background(), "hello")

	require.Len(t, base.Records(), 1)

	found := false

	base.Records()[0].Attrs(func(a slog.Attr) bool {
		if a.Key == logger.TraceIDLogKey {
			found = true
		}

		return true
	})

	require.False(t, found)
}

func TestTraceContextHandler_WithGroupPanics(t *testing.T) {
	t.Parallel()

	h := logger.NewTraceContextHandler(slog.DiscardHandler)

	assert.Panics(t, func() {
		slog.New(h).WithGroup("group")
	})
}

func TestTraceContextHandler_WithAttrsPreservesTraceID(t *testing.T) {
	t.Parallel()

	base := testmock.NewSlogHandlerMock()
	h := logger.NewTraceContextHandler(base)

	log := slog.New(h).
		With(slog.String("foo", "bar")).
		With(slog.String("baz", "qux"))

	traceID := uuid.New()
	ctx := logger.ContextWithTraceID(context.Background(), traceID)

	log.InfoContext(ctx, "hello")

	require.Len(t, base.Records(), 1)

	attrs := map[string]string{}

	base.Records()[0].Attrs(func(a slog.Attr) bool {
		if _, ok := attrs[a.Key]; ok {
			require.FailNow(t, "should not have duplicate attrs")
		}

		attrs[a.Key] = a.Value.String()

		return true
	})

	require.Equal(t, "bar", attrs["foo"])
	require.Equal(t, "qux", attrs["baz"])
	require.Equal(t, traceID.String(), attrs[logger.TraceIDLogKey])
}
