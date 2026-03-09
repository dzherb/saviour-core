package command

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/knadh/koanf/v2"

	"saviour/internal/app"
	"saviour/internal/infra/config"
	"saviour/internal/service/user"
)

type CreateAdminParams struct {
	Username string
	Password string
}

func CreateAdmin(
	ctx context.Context,
	cfgPaths []string,
	params CreateAdminParams,
) error {
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
	_ = logComponent.Start(context.TODO(), cfg)

	instanceMetaComponent := app.NewInstanceMeta(di)
	dbComponent := app.NewDB(di)
	repositoryComponent := app.NewRepository(di)
	userServiceComponent := app.NewUserService(di)

	appRunner := app.New(app.LogDependency.MustGet(di))

	appRunner.RegisterStartQueue(
		instanceMetaComponent,
		dbComponent,
		repositoryComponent,
		userServiceComponent,
		app.StarterFunc(func(ctx context.Context, _ *koanf.Koanf) error {
			defer stop()

			return createAdminRunner(
				ctx,
				app.LogDependency.MustGet(di),
				app.UserServiceDependency.MustGet(di),
				params,
			)
		}),
	)

	appRunner.RegisterStopQueue(
		dbComponent,
	)

	return appRunner.Run(ctx, cfg)
}

func createAdminRunner(
	ctx context.Context,
	log *slog.Logger,
	userService user.Service,
	params CreateAdminParams,
) error {
	_, err := userService.CreateUser(
		ctx,
		user.CreateUserParams{
			Username: params.Username,
			Password: params.Password,
			IsAdmin:  true,
		},
	)
	if err != nil {
		return err
	}

	log.InfoContext(
		ctx,
		"admin user successfully created",
		"username", params.Username,
	)

	return nil
}
