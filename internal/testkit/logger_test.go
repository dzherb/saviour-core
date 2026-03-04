package testkit_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"saviour/internal/testkit"
)

func TestMockHandler_CapturesRecord(t *testing.T) {
	t.Parallel()

	h := testkit.NewSlogHandlerMock()
	log := slog.New(h)

	log.Info("hello", slog.String("k", "v"))

	recs := h.Records()
	require.Len(t, recs, 1)

	r := recs[0]
	require.Equal(t, "hello", r.Message)
	require.Equal(t, slog.LevelInfo, r.Level)

	var attrs []slog.Attr

	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	require.Len(t, attrs, 1)
	require.Equal(t, "k", attrs[0].Key)
	require.Equal(t, "v", attrs[0].Value.String())
}

func TestMockHandler_WithAttrs(t *testing.T) {
	t.Parallel()

	h := testkit.NewSlogHandlerMock()
	log := slog.New(h)

	base := log
	with := log.With("a", "1")

	base.Info("base")
	with.Info("with")

	recs := h.Records()
	require.Len(t, recs, 2)

	var baseAttrs []slog.Attr

	recs[0].Attrs(func(a slog.Attr) bool {
		baseAttrs = append(baseAttrs, a)
		return true
	})
	require.Empty(t, baseAttrs)

	var withAttrs []slog.Attr

	recs[1].Attrs(func(a slog.Attr) bool {
		withAttrs = append(withAttrs, a)
		return true
	})
	require.Len(t, withAttrs, 1)
	require.Equal(t, "a", withAttrs[0].Key)
	require.Equal(t, "1", withAttrs[0].Value.String())
}

func TestMockHandler_WithGroup(t *testing.T) {
	t.Parallel()

	h := testkit.NewSlogHandlerMock()
	log := slog.New(h)

	log.
		WithGroup("grp").
		With("a", "1").
		Info("msg")

	recs := h.Records()
	require.Len(t, recs, 1)

	var attrs []slog.Attr

	recs[0].Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	require.Len(t, attrs, 1)
	require.Equal(t, "grp", attrs[0].Key)

	group := attrs[0].Value.Group()
	require.Len(t, group, 1)
	require.Equal(t, "a", group[0].Key)
	require.Equal(t, "1", group[0].Value.String())
}

func TestMockHandler_WithAndInlineAttrs(t *testing.T) {
	t.Parallel()

	h := testkit.NewSlogHandlerMock()
	log := slog.New(h)

	log.
		With("a", "1").
		Info("msg", slog.String("b", "2"))

	recs := h.Records()
	require.Len(t, recs, 1)

	var m = map[string]string{}

	recs[0].Attrs(func(a slog.Attr) bool {
		m[a.Key] = a.Value.String()
		return true
	})

	require.Equal(t, map[string]string{
		"a": "1",
		"b": "2",
	}, m)
}

func TestMockHandler_Concurrent(t *testing.T) {
	t.Parallel()

	h := testkit.NewSlogHandlerMock()
	log := slog.New(h)

	done := make(chan struct{})

	for i := range 100 {
		go func(i int) {
			log.Info("msg", slog.Int("i", i))

			done <- struct{}{}
		}(i)
	}

	for range 100 {
		<-done
	}

	require.Equal(t, 100, len(h.Records()))
}
