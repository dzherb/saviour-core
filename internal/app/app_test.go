package app_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"saviour/internal/app"
	"saviour/internal/infra/config"
	"saviour/internal/logger"
)

type mockComponent struct {
	name string

	startFn  func(context.Context, *koanf.Koanf) error
	stopFn   func(context.Context) error
	reloadFn func(context.Context, *koanf.Koanf) error
}

func (m *mockComponent) Start(ctx context.Context, k *koanf.Koanf) error {
	if m.startFn != nil {
		return m.startFn(ctx, k)
	}

	return nil
}

func (m *mockComponent) Stop(ctx context.Context) error {
	if m.stopFn != nil {
		return m.stopFn(ctx)
	}

	return nil
}

func (m *mockComponent) Reload(ctx context.Context, k *koanf.Koanf) error {
	if m.reloadFn != nil {
		return m.reloadFn(ctx, k)
	}

	return nil
}

func validCfg(extra ...map[string]any) *koanf.Koanf {
	valid := map[string]any{
		"app.instance": "test",
	}

	for _, m := range extra {
		for k, v := range m {
			valid[k] = v
		}
	}

	k := koanf.New(config.KeysDelimiter)

	err := k.Load(
		confmap.Provider(
			valid,
			config.KeysDelimiter,
		),
		nil,
	)
	if err != nil {
		panic(err)
	}

	return k
}

func TestApp_StartOrder(t *testing.T) {
	t.Parallel()

	var order []string

	c1 := &mockComponent{
		name: "c1",
		startFn: func(ctx context.Context, _ *koanf.Koanf) error {
			order = append(order, "c1")
			return nil
		},
	}

	c2 := &mockComponent{
		name: "c2",
		startFn: func(ctx context.Context, _ *koanf.Koanf) error {
			order = append(order, "c2")
			return nil
		},
	}

	app := app.New(logger.Noop)
	app.RegisterStartQueue(c1, c2)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assert.NoError(t, app.Run(ctx, validCfg()))
	assert.Equal(t, []string{"c1", "c2"}, order)
}

func TestApp_StopOrder(t *testing.T) {
	t.Parallel()

	var order []string

	c1 := &mockComponent{
		stopFn: func(ctx context.Context) error {
			order = append(order, "c1")
			return nil
		},
	}

	c2 := &mockComponent{
		stopFn: func(ctx context.Context) error {
			order = append(order, "c2")
			return nil
		},
	}

	app := app.New(logger.Noop)
	app.RegisterStopQueue(c2, c1)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	assert.NoError(t, app.Run(ctx, validCfg()))
	assert.Equal(t, []string{"c2", "c1"}, order)
}

func TestApp_PanicRecovered(t *testing.T) {
	t.Parallel()

	stopped := false

	c := &mockComponent{
		startFn: func(ctx context.Context, _ *koanf.Koanf) error {
			panic("boom")
		},
		stopFn: func(ctx context.Context) error {
			stopped = true
			return nil
		},
	}

	app := app.New(logger.Noop)
	app.RegisterStartQueue(c)
	app.RegisterStopQueue(c)

	err := app.Run(context.Background(), validCfg())

	require.Error(t, err)
	require.Contains(t, err.Error(), "boom")
	require.True(t, stopped)
}

func TestApp_StartupTimeout(t *testing.T) {
	t.Parallel()

	var (
		started bool
		stopped bool
	)

	blocking := &mockComponent{
		startFn: func(ctx context.Context, _ *koanf.Koanf) error {
			started = true

			<-ctx.Done()

			return ctx.Err()
		},
		stopFn: func(ctx context.Context) error {
			stopped = true
			return nil
		},
	}

	app := app.New(logger.Noop)

	app.RegisterStartQueue(blocking)
	app.RegisterStopQueue(blocking)

	start := time.Now()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	cfg := validCfg(map[string]any{
		"app.startup_timeout": "50ms",
	})

	assert.ErrorIs(t, app.Run(ctx, cfg), context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 200*time.Millisecond)
	assert.True(t, started)
	assert.True(t, stopped)
}

func TestApp_ShutdownTimeout(t *testing.T) {
	t.Parallel()

	c := &mockComponent{
		stopFn: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	app := app.New(logger.Noop)
	app.RegisterStopQueue(c)

	cfg := validCfg(map[string]any{
		"app.shutdown_timeout": "50ms",
	})

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	start := time.Now()

	assert.ErrorIs(t, app.Run(ctx, cfg), context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 100*time.Millisecond)
}

func TestApp_FatalErrorStopsApplication(t *testing.T) {
	t.Parallel()

	var stopped bool

	fatalErr := fmt.Errorf("server crashed")

	component := &mockComponent{
		startFn: func(ctx context.Context, cfg *koanf.Koanf) error {
			go func() {
				time.Sleep(10 * time.Millisecond)
				app.PropagateFatalError(ctx, fatalErr)
			}()

			return nil
		},
		stopFn: func(ctx context.Context) error {
			stopped = true
			return nil
		},
	}

	application := app.New(logger.Noop)

	application.RegisterStartQueue(component)
	application.RegisterStopQueue(component)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	err := application.Run(ctx, validCfg())

	assert.Error(t, err)
	assert.ErrorIs(t, err, fatalErr)
	assert.True(t, stopped)
}
