package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"saviour/internal/infra/config"
)

func Run(cfgPaths []string) error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	di := NewContainer()

	cfg, err := config.Load(cfgPaths)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	logComponent := NewLog(di)
	// Logger is used internally by the app runner,
	// so initiate it manually first...
	_ = logComponent.Start(context.TODO(), cfg)

	instanceComponent := NewInstance(di)
	dbComponent := NewDB(di)
	repositoryComponent := NewRepository(di)
	authServiceComponent := NewAuthService(di)
	userServiceComponent := NewUserService(di)
	serverComponent := NewServer(di)

	app := New(LogDependency.MustGet(di))

	app.RegisterStartQueue(
		instanceComponent,
		dbComponent,
		repositoryComponent,
		authServiceComponent,
		userServiceComponent,
		serverComponent,
	)

	// Gracefully stop the app by stopping the specified components one by one.
	// shutdown_timeout can be set to force stop the app after N second.
	app.RegisterStopQueue(
		serverComponent,
		dbComponent,
	)

	return app.Run(ctx, cfg)
}
