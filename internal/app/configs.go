package app

import (
	"fmt"
	"log/slog"

	"github.com/knadh/koanf/v2"

	"saviour/internal/infra/logger"
	"saviour/internal/infra/secretkey"
	"saviour/internal/infra/sqlite"
	"saviour/internal/service/auth"
	"saviour/internal/transport/rest"
	"saviour/pkg/secret"
)

var ErrConfigParsing = fmt.Errorf("error parsing config")

func asConfigParseError(err error) error {
	return fmt.Errorf("%w: %w", ErrConfigParsing, err)
}

func valueRequiredError(key string) error {
	return asConfigParseError(
		fmt.Errorf("'%s' value is required", key),
	)
}

func valueNotValidError(key string, reason string) error {
	if reason == "" {
		return asConfigParseError(
			fmt.Errorf("'%s' value is not valid", key),
		)
	}

	return asConfigParseError(
		fmt.Errorf("'%s' value is not valid: %s", key, reason),
	)
}

func appConfig(k *koanf.Koanf) (Config, error) {
	const (
		instanceKey        = "app.instance"
		startupTimeoutKey  = "app.startup_timeout"
		shutdownTimeoutKey = "app.shutdown_timeout"
	)

	cfg := Config{
		Instance:        k.String(instanceKey),
		StartupTimeout:  k.Duration(startupTimeoutKey),
		ShutdownTimeout: k.Duration(shutdownTimeoutKey),
	}

	if cfg.Instance == "" {
		return cfg, valueRequiredError(instanceKey)
	}

	return cfg, nil
}

func secretKeyConfig(k *koanf.Koanf) (secretkey.Config, error) {
	const secretKey = "secret_key"

	cfg := secretkey.Config{}

	cfg.Key = secret.New(k.String(secretKey))
	if cfg.Key.UnsafeValue() == "" {
		return cfg, valueRequiredError(secretKey)
	}

	return cfg, nil
}

func loggerConfig(k *koanf.Koanf) (logger.Config, error) {
	const lvlKey = "logger.level"

	var opt logger.Config

	var logLvl slog.Level

	err := logLvl.UnmarshalText([]byte(k.String(lvlKey)))
	if err != nil {
		return opt, valueNotValidError(lvlKey, err.Error())
	}

	opt.Level = logLvl

	return opt, nil
}

func serverConfig(k *koanf.Koanf) (rest.ServerConfig, error) {
	const (
		hostKey              = "server.host"
		portKey              = "server.port"
		readHeaderTimeoutKey = "server.read_header_timeout"
	)

	var cfg rest.ServerConfig

	cfg.Host = k.String(hostKey)
	if cfg.Host == "" {
		return cfg, valueRequiredError(hostKey)
	}

	cfg.Port = k.Int(portKey)
	if cfg.Port == 0 {
		return cfg, valueRequiredError(portKey)
	}

	cfg.ReadHeaderTimeout = k.Duration(readHeaderTimeoutKey)

	return cfg, nil
}

func apiConfig(k *koanf.Koanf) rest.APIConfig {
	const (
		responseTimeoutKey           = "server.api.response_timeout"
		debugLogProcessedRequestsKey = "server.api.debug_log_processed_requests"
	)

	return rest.APIConfig{
		ResponseTimeout:           k.Duration(responseTimeoutKey),
		DebugLogProcessedRequests: k.Bool(debugLogProcessedRequestsKey),
	}
}

const (
	dbEngineSQLite = "sqlite"
)

func sqliteConfig(k *koanf.Koanf) (sqlite.Config, error) {
	const (
		dbEngineKey = "db.engine"
		dbDSNKey    = "db.dsn"
	)

	var cfg sqlite.Config

	if k.String(dbEngineKey) != dbEngineSQLite {
		return cfg, valueNotValidError(
			dbEngineKey,
			"only "+dbEngineSQLite+" is supported",
		)
	}

	cfg.DSN = k.String(dbDSNKey)
	if cfg.DSN == "" {
		return cfg, valueRequiredError(dbDSNKey)
	}

	return cfg, nil
}

func authConfig(k *koanf.Koanf) (auth.Config, error) {
	const (
		accessTokenSecretKey  = "server.api.auth.access_token_secret"  //nolint:gosec
		refreshTokenSecretKey = "server.api.auth.refresh_token_secret" //nolint:gosec
	)

	var cfg auth.Config

	cfg.AccessTokenSecret = secret.New(k.String(accessTokenSecretKey))
	if cfg.AccessTokenSecret.UnsafeValue() == "" {
		return cfg, valueRequiredError(accessTokenSecretKey)
	}

	cfg.RefreshTokenSecret = secret.New(k.String(refreshTokenSecretKey))
	if cfg.RefreshTokenSecret.UnsafeValue() == "" {
		return cfg, valueRequiredError(refreshTokenSecretKey)
	}

	return cfg, nil
}
