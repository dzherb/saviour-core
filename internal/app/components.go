package app

import (
	"context"
	"database/sql"
	"log/slog"
	"net/netip"
	"os"

	"github.com/knadh/koanf/v2"

	"saviour/internal/infra/sqlite"
	"saviour/internal/logger"
	"saviour/internal/repository"
	sqliterepo "saviour/internal/repository/sqlite"
	"saviour/internal/service/auth"
	"saviour/internal/transport/rest"
	"saviour/internal/transport/rest/handler"
)

type Starter interface {
	Start(ctx context.Context, cfg *koanf.Koanf) error
}

type Stopper interface {
	Stop(ctx context.Context) error
}

var (
	InstanceDependency = DefineDependency[string]("instance")
	LogDependency      = DefineDependency[*slog.Logger]("log")
	DBDependency       = DefineDependency[*sql.DB]("db")

	UserRepositoryDependency = DefineDependency[repository.UserRepository](
		"repository.user",
	)
	SessionRepositoryDependency = DefineDependency[repository.SessionRepository](
		"repository.session",
	)

	AuthServiceDependency = DefineDependency[*auth.Auth]("auth")
)

type InstanceComponent struct {
	di *Container
}

func NewInstance(di *Container) *InstanceComponent {
	return &InstanceComponent{
		di: di,
	}
}

func (c *InstanceComponent) Start(ctx context.Context, _ *koanf.Koanf) error {
	InstanceDependency.Set(c.di, Instance(ctx))

	return nil
}

type LogComponent struct {
	di *Container
}

func NewLog(di *Container) *LogComponent {
	return &LogComponent{
		di: di,
	}
}

func (c *LogComponent) Start(_ context.Context, cfg *koanf.Koanf) error {
	opt, err := loggerConfig(cfg)
	if err != nil {
		return err
	}

	h := slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     opt.Level,
		},
	)

	log := slog.New(logger.NewTraceContextHandler(h))

	LogDependency.Set(c.di, log)

	return nil
}

type ServerComponent struct {
	di *Container

	log    *slog.Logger
	server *rest.Server
}

func NewServer(di *Container) *ServerComponent {
	return &ServerComponent{di: di}
}

func (c *ServerComponent) Start(ctx context.Context, cfg *koanf.Koanf) error {
	serverCfg, err := serverConfig(cfg)
	if err != nil {
		return err
	}

	c.log = LogDependency.MustGet(c.di)
	instance := InstanceDependency.MustGet(c.di)
	authService := AuthServiceDependency.MustGet(c.di)

	h := rest.RootHandler(
		c.log,
		handler.NewAPIHandler(
			handler.NewPingHandler(instance),
			handler.NewAuthHandler(
				c.log,
				authService,
				[]netip.Prefix{},
			),
		),
		authService,
		apiConfig(cfg),
	)

	server := rest.NewServer(c.log, h, serverCfg)

	c.server = server

	go func() {
		c.log.Info(
			"starting http server",
			slog.String("address", c.server.Addr()),
		)

		err := c.server.Serve()
		if err != nil {
			PropagateFatalError(ctx, err)
		}
	}()

	return nil
}

func (c *ServerComponent) Stop(ctx context.Context) error {
	if c.server != nil {
		c.log.DebugContext(ctx, "shutting down http server")

		return c.server.Shutdown(ctx)
	}

	return nil
}

type DBComponent struct {
	di *Container

	db *sql.DB
}

func NewDB(di *Container) *DBComponent {
	return &DBComponent{
		di: di,
	}
}

func (c *DBComponent) Start(_ context.Context, cfg *koanf.Koanf) error {
	sqliteCfg, err := sqliteConfig(cfg)
	if err != nil {
		return err
	}

	db, err := sqlite.NewDB(sqliteCfg)
	if err != nil {
		return err
	}

	c.db = db

	DBDependency.Set(c.di, db)

	return nil
}

func (c *DBComponent) Stop(_ context.Context) error {
	if c.db != nil {
		return c.db.Close()
	}

	return nil
}

type RepositoryComponent struct {
	di *Container
}

func NewRepository(di *Container) *RepositoryComponent {
	return &RepositoryComponent{di: di}
}

func (c *RepositoryComponent) Start(_ context.Context, cfg *koanf.Koanf) error {
	db := DBDependency.MustGet(c.di)

	UserRepositoryDependency.Set(
		c.di,
		sqliterepo.NewUserRepository(db),
	)
	SessionRepositoryDependency.Set(
		c.di,
		sqliterepo.NewSessionsRepository(db),
	)

	return nil
}

type AuthServiceComponent struct {
	di *Container
}

func NewAuthService(di *Container) *AuthServiceComponent {
	return &AuthServiceComponent{di: di}
}

func (c *AuthServiceComponent) Start(
	_ context.Context,
	cfg *koanf.Koanf,
) error {
	authCfg, err := authConfig(cfg)
	if err != nil {
		return err
	}

	log := LogDependency.MustGet(c.di)
	userRepo := UserRepositoryDependency.MustGet(c.di)
	sessionRepo := SessionRepositoryDependency.MustGet(c.di)

	authService := auth.New(log, userRepo, sessionRepo, authCfg)

	AuthServiceDependency.Set(c.di, authService)

	return nil
}
