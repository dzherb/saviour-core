package logger

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type traceCtxKeyType string

const traceCtxKey traceCtxKeyType = "trace"

func NewTraceID() uuid.UUID {
	id, _ := uuid.NewV7()
	return id
}

func ContextWithTraceID(
	ctx context.Context,
	traceID uuid.UUID,
) context.Context {
	if traceID == uuid.Nil {
		return ctx
	}

	return context.WithValue(ctx, traceCtxKey, traceID)
}

func TraceIDFromCtx(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(traceCtxKey).(uuid.UUID); ok {
		return id
	}

	return uuid.Nil
}

const TraceIDLogKey = "trace_id"

// TraceContextHandler injects a trace id attribute into log records
// using the value extracted from the context.
type TraceContextHandler struct {
	next slog.Handler
}

func NewTraceContextHandler(next slog.Handler) *TraceContextHandler {
	return &TraceContextHandler{next: next}
}

func (h *TraceContextHandler) Enabled(
	ctx context.Context,
	level slog.Level,
) bool {
	return h.next.Enabled(ctx, level)
}

func (h *TraceContextHandler) Handle(ctx context.Context, r slog.Record) error {
	traceID := TraceIDFromCtx(ctx)
	if traceID != uuid.Nil {
		r.AddAttrs(
			slog.String(TraceIDLogKey, traceID.String()),
		)
	}

	return h.next.Handle(ctx, r)
}

func (h *TraceContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceContextHandler{
		next: h.next.WithAttrs(attrs),
	}
}

func (h *TraceContextHandler) WithGroup(_ string) slog.Handler {
	// Wrapping a grouped handler would place trace_id inside the group.
	// It is not a desirable behavior, trace_id must stay top-level.
	panic(
		"TraceContextHandler must not be used with WithGroup; use slog.Group instead",
	)
}
