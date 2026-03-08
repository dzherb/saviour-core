package app

import (
	"context"
	"database/sql"
	"log/slog"
	"net/netip"
	"os"

	"github.com/knadh/koanf/v2"

	"saviour/internal/infra/secretkey"
	"saviour/internal/infra/sqlite"
	"saviour/internal/logger"
	"saviour/internal/repository"
	sqliterepo "saviour/internal/repository/sqlite"
	"saviour/internal/service/acl"
	"saviour/internal/service/auth"
	secretservice "saviour/internal/service/secret"
	"saviour/internal/service/user"
	"saviour/internal/service/workspace"
	"saviour/internal/transport/rest"
	"saviour/internal/transport/rest/handler"
	"saviour/pkg/secret"
)

type Starter interface {
	Start(ctx context.Context, cfg *koanf.Koanf) error
}

type StarterFunc func(ctx context.Context, cfg *koanf.Koanf) error

func (f StarterFunc) Start(ctx context.Context, cfg *koanf.Koanf) error {
	return f(ctx, cfg)
}

type Stopper interface {
	Stop(ctx context.Context) error
}

var (
	InstanceDependency  = DefineDependency[string]("instance")
	SecretKeyDependency = DefineDependency[secret.Secret[[]byte]]("secret_key")
	LogDependency       = DefineDependency[*slog.Logger]("log")
	DBDependency        = DefineDependency[*sql.DB]("db")

	UserRepositoryDependency = DefineDependency[repository.UserRepository](
		"repository.user",
	)
	SessionRepositoryDependency = DefineDependency[repository.SessionRepository](
		"repository.session",
	)
	WorkspaceRepositoryDependency = DefineDependency[repository.WorkspaceRepository]( //nolint:lll
		"repository.workspace",
	)
	SecretRepositoryDependency = DefineDependency[repository.SecretRepository](
		"repository.secret",
	)

	AuthServiceDependency = DefineDependency[*auth.ServiceImpl](
		"service.auth",
	)
	ACLServiceDependency = DefineDependency[*acl.ServiceImpl](
		"service.acl",
	)
	UserServiceDependency = DefineDependency[*user.ServiceImpl](
		"service.user",
	)
	WorkspaceServiceDependency = DefineDependency[*workspace.ServiceImpl](
		"service.workspace",
	)
	SecretServiceDependency = DefineDependency[*secretservice.ServiceImpl](
		"service.secret",
	)
)

type InstanceMetaComponent struct {
	di *Container
}

func NewInstanceMeta(di *Container) *InstanceMetaComponent {
	return &InstanceMetaComponent{
		di: di,
	}
}

func (c *InstanceMetaComponent) Start(
	ctx context.Context,
	cfg *koanf.Koanf,
) error {
	secretKeyCfg, err := secretKeyConfig(cfg)
	if err != nil {
		return err
	}

	InstanceDependency.Set(c.di, InstanceFromCtx(ctx))
	SecretKeyDependency.Set(c.di, secretkey.New(secretKeyCfg))

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
	aclService := ACLServiceDependency.MustGet(c.di)
	userService := UserServiceDependency.MustGet(c.di)
	workspaceService := WorkspaceServiceDependency.MustGet(c.di)
	secretService := SecretServiceDependency.MustGet(c.di)

	h := rest.RootHandler(
		c.log,
		handler.NewAPIHandler(
			handler.NewPingHandler(instance),
			handler.NewAuthHandler(
				c.log,
				authService,
				[]netip.Prefix{},
			),
			handler.NewUserHandler(
				c.log,
				userService,
			),
			handler.NewWorkspaceHandler(
				c.log,
				workspaceService,
			),
			handler.NewSecretHandler(
				c.log,
				aclService,
				secretService,
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
	WorkspaceRepositoryDependency.Set(
		c.di,
		sqliterepo.NewWorkspaceRepository(db),
	)
	SecretRepositoryDependency.Set(
		c.di,
		sqliterepo.NewSecretRepository(db),
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

type ACLServiceComponent struct {
	di *Container
}

func NewACLService(di *Container) *ACLServiceComponent {
	return &ACLServiceComponent{di: di}
}

func (c *ACLServiceComponent) Start(_ context.Context, _ *koanf.Koanf) error {
	aclService := acl.NewService(
		WorkspaceRepositoryDependency.MustGet(c.di),
	)

	ACLServiceDependency.Set(c.di, aclService)

	return nil
}

type UserServiceComponent struct {
	di *Container
}

func NewUserService(di *Container) *UserServiceComponent {
	return &UserServiceComponent{di: di}
}

func (c *UserServiceComponent) Start(
	_ context.Context,
	_ *koanf.Koanf,
) error {
	log := LogDependency.MustGet(c.di)
	userRepo := UserRepositoryDependency.MustGet(c.di)

	userService := user.NewService(log, userRepo)

	UserServiceDependency.Set(c.di, userService)

	return nil
}

type WorkspaceServiceComponent struct {
	di *Container
}

func NewWorkspaceService(di *Container) *WorkspaceServiceComponent {
	return &WorkspaceServiceComponent{di: di}
}

func (c *WorkspaceServiceComponent) Start(
	_ context.Context,
	_ *koanf.Koanf,
) error {
	log := LogDependency.MustGet(c.di)
	workspaceRepo := WorkspaceRepositoryDependency.MustGet(c.di)

	workspaceService := workspace.NewService(log, workspaceRepo)

	WorkspaceServiceDependency.Set(c.di, workspaceService)

	return nil
}

type SecretServiceComponent struct {
	di *Container
}

func NewSecretService(di *Container) *SecretServiceComponent {
	return &SecretServiceComponent{di: di}
}

func (c *SecretServiceComponent) Start(
	_ context.Context,
	_ *koanf.Koanf,
) error {
	secretService := secretservice.NewService(
		SecretKeyDependency.MustGet(c.di),
		SecretRepositoryDependency.MustGet(c.di),
	)

	SecretServiceDependency.Set(c.di, secretService)

	return nil
}
