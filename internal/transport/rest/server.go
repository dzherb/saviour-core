package rest

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/justinas/alice"

	"saviour/internal/logger"
	"saviour/internal/transport/rest/api"
	"saviour/internal/transport/rest/middleware"
)

type ServerConfig struct {
	Host              string
	Port              int
	ReadHeaderTimeout time.Duration
}

type APIConfig struct {
	ResponseTimeout           time.Duration
	DebugLogProcessedRequests bool
}

const apiPrefix = "/api"

func RootHandler(
	log *slog.Logger,
	apiImpl api.HandlerInterface,
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
	}

	apiMiddlewares := []alice.Constructor{
		middleware.UseIf(
			cfg.DebugLogProcessedRequests,
			middleware.DebugLogProcessedRequests(log),
		),
		middleware.Timeout(log, cfg.ResponseTimeout),
		middleware.OAPIValidator(log, spec),
	}

	apiHandler = alice.New(apiMiddlewares...).Then(apiHandler)

	mux := http.NewServeMux()

	mux.Handle(
		apiPrefix+"/",
		http.StripPrefix(apiPrefix, apiHandler),
	)

	return alice.New(globalMiddlewares...).Then(mux)
}

type Server struct {
	log *slog.Logger
	srv *http.Server
}

func NewServer(
	log *slog.Logger,
	h http.Handler,
	cfg ServerConfig,
) *Server {
	srv := &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Handler:           h,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}

	srv.ErrorLog = logger.SlogToStdLog(log, slog.LevelError)

	m := &Server{
		log: log,
		srv: srv,
	}

	return m
}

func (s *Server) Serve() error {
	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return err
	}

	if err := s.srv.Serve(ln); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

const shutdownTimeout = 30 * time.Second

func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	return s.srv.Shutdown(ctx)
}

func (s *Server) Addr() string {
	return s.srv.Addr
}
