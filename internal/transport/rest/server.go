package rest

import (
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/justinas/alice"

	"saviour/internal/logger"
	"saviour/internal/service/auth"
	"saviour/internal/transport/rest/api"
	"saviour/internal/transport/rest/middleware"
)

type ServerConfig struct {
	Host              string
	Port              int
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration
}

type APIConfig struct {
	ResponseTimeout           time.Duration
	DebugLogProcessedRequests bool
}

const apiPrefix = "/api"

func RootHandler(
	log *slog.Logger,
	apiImpl api.HandlerInterface,
	tokenValidator auth.TokenValidator,
	cfg APIConfig,
) http.Handler {
	spec, err := api.GetSwagger()
	if err != nil {
		panic("parsing swagger spec: " + err.Error())
	}

	apiHandler := api.NewHTTPHandler(log, apiImpl)

	globalMiddlewares := []alice.Constructor{
		middleware.WithTraceIDCtx(),
		middleware.Recover(log),
		middleware.CORS(),
	}

	apiMiddlewares := []alice.Constructor{
		middleware.UseIf(
			cfg.DebugLogProcessedRequests,
			middleware.DebugLogProcessedRequests(log),
		),
		middleware.Timeout(log, cfg.ResponseTimeout),
		middleware.OAPIValidator(log, spec),
		middleware.Authenticator(log, tokenValidator),
	}

	apiHandler = alice.New(apiMiddlewares...).Then(apiHandler)

	mux := http.NewServeMux()

	mux.Handle(
		apiPrefix+"/",
		http.StripPrefix(apiPrefix, apiHandler),
	)

	return alice.New(globalMiddlewares...).Then(mux)
}

func NewServer(
	log *slog.Logger,
	h http.Handler,
	cfg ServerConfig,
) *http.Server {
	srv := &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Handler:           h,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ErrorLog:          logger.SlogToStdLog(log, slog.LevelError),
	}

	return srv
}
