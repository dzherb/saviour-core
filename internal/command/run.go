package command

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"saviour/internal/app"
	"saviour/internal/infra/config"
)

func Run(ctx context.Context, cfgPaths []string) error {
	ctx, stop := signal.NotifyContext(
		ctx,
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	di := app.NewContainer()

	cfg, err := config.Load(cfgPaths)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	logComponent := app.NewLog(di)
	// Logger is used internally by the app runner,
	// so initiate it manually first...
	_ = logComponent.Start(context.TODO(), cfg)

	instanceMetaComponent := app.NewInstanceMeta(di)
	dbComponent := app.NewDB(di)
	repositoryComponent := app.NewRepository(di)
	authServiceComponent := app.NewAuthService(di)
	aclServiceComponent := app.NewACLService(di)
	userServiceComponent := app.NewUserService(di)
	workspaceServiceComponent := app.NewWorkspaceService(di)
	secretServiceComponent := app.NewSecretService(di)
	rootHandlerComponent := app.NewRootHandler(di)
	serverComponent := app.NewServer(di)

	appRunner := app.New(app.LogDependency.MustGet(di))

	appRunner.RegisterStartQueue(
		instanceMetaComponent,
		dbComponent,
		repositoryComponent,
		authServiceComponent,
		aclServiceComponent,
		userServiceComponent,
		workspaceServiceComponent,
		secretServiceComponent,
		rootHandlerComponent,
		serverComponent,
	)

	// Gracefully stop the app by stopping the specified components one by one.
	// shutdown_timeout can be set to force stop the app after N second.
	appRunner.RegisterStopQueue(
		serverComponent,
		dbComponent,
	)

	return appRunner.Run(ctx, cfg)
}
