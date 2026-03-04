package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/knadh/koanf/v2"

	"saviour/internal/logger"
)

type Config struct {
	Instance        string
	StartupTimeout  time.Duration
	ShutdownTimeout time.Duration
}

type App struct {
	log     *slog.Logger
	cfg     Config
	fatalCh chan error

	startQueue []Starter
	stopQueue  []Stopper
}

func New(log *slog.Logger) *App {
	return &App{
		log:     log,
		fatalCh: make(chan error),
	}
}

func (a *App) RegisterStartQueue(startQueue ...Starter) {
	a.startQueue = startQueue
}

func (a *App) RegisterStopQueue(stopQueue ...Stopper) {
	a.stopQueue = stopQueue
}

func (a *App) Run(
	ctx context.Context,
	cfg *koanf.Koanf,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()

			func() {
				defer func() {
					_ = recover() // swallow panic from stop
				}()

				_ = a.stop()
			}()

			err = fmt.Errorf("panic recovered: %v\n%s", r, stack)
		}
	}()

	a.cfg, err = appConfig(cfg)
	if err != nil {
		return fmt.Errorf("error parsing app config: %w", err)
	}

	if err := a.start(ctx, cfg); err != nil {
		_ = a.stop()

		return err
	}

	for {
		select {
		case <-ctx.Done():
			return a.stop()

		case err := <-a.fatalCh:
			a.log.ErrorContext(
				ctx,
				"received a fatal error from a component",
				logger.ErrorAttr(err),
			)
			_ = a.stop()

			return err
		}
	}
}

func (a *App) Fatal(err error) {
	select {
	case a.fatalCh <- err:
	default:
	}
}

func (a *App) start(ctx context.Context, cfg *koanf.Koanf) error {
	startCtx := contextWithApp(ctx, a)

	if a.cfg.StartupTimeout != 0 {
		timeoutCtx, cancel := context.WithTimeout(
			startCtx,
			a.cfg.StartupTimeout,
		)

		startCtx = timeoutCtx

		defer cancel()
	}

	for _, s := range a.startQueue {
		if err := s.Start(startCtx, cfg); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) stop() error {
	ctx := context.Background()

	if a.cfg.ShutdownTimeout != 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, a.cfg.ShutdownTimeout)

		ctx = timeoutCtx

		defer cancel()
	}

	var errs error

	a.log.InfoContext(ctx, "stopping the application")

	for _, s := range a.stopQueue {
		if err := s.Stop(ctx); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

type appCtxKey struct{}

func contextWithApp(ctx context.Context, a *App) context.Context {
	return context.WithValue(ctx, appCtxKey{}, a)
}

func PropagateFatalError(ctx context.Context, err error) {
	app, ok := ctx.Value(appCtxKey{}).(*App)
	if ok {
		app.Fatal(err)
	}
}

func Instance(ctx context.Context) string {
	app, ok := ctx.Value(appCtxKey{}).(*App)
	if ok {
		return app.cfg.Instance
	}

	return ""
}
