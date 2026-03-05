package testmock

import (
	"context"
	"log/slog"
	"sync"
)

type handlerState struct {
	mu      sync.Mutex
	records []slog.Record
}

type SlogHandlerMock struct {
	state  *handlerState
	attrs  []slog.Attr
	groups []string
}

func NewSlogHandlerMock() *SlogHandlerMock {
	return &SlogHandlerMock{
		state: &handlerState{},
	}
}

func (h *SlogHandlerMock) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *SlogHandlerMock) Handle(_ context.Context, r slog.Record) error {
	rc := slog.Record{
		Time:    r.Time,
		Level:   r.Level,
		Message: r.Message,
		PC:      r.PC,
	}

	var all []slog.Attr

	all = append(all, h.attrs...)

	r.Attrs(func(a slog.Attr) bool {
		all = append(all, a)
		return true
	})

	if len(h.groups) > 0 {
		grouped := all
		for i := len(h.groups) - 1; i >= 0; i-- {
			grouped = []slog.Attr{
				{
					Key:   h.groups[i],
					Value: slog.GroupValue(grouped...),
				},
			}
		}

		all = grouped
	}

	rc.AddAttrs(all...)

	h.state.mu.Lock()
	h.state.records = append(h.state.records, rc)
	h.state.mu.Unlock()

	return nil
}

func (h *SlogHandlerMock) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := h.clone()
	h2.attrs = append(h2.attrs, attrs...)

	return h2
}

func (h *SlogHandlerMock) WithGroup(name string) slog.Handler {
	h2 := h.clone()
	h2.groups = append(h2.groups, name)

	return h2
}

func (h *SlogHandlerMock) Records() []slog.Record {
	h.state.mu.Lock()
	defer h.state.mu.Unlock()

	out := make([]slog.Record, len(h.state.records))
	copy(out, h.state.records)

	return out
}

func (h *SlogHandlerMock) clone() *SlogHandlerMock {
	return &SlogHandlerMock{
		state:  h.state,
		attrs:  append([]slog.Attr(nil), h.attrs...),
		groups: append([]string(nil), h.groups...),
	}
}
