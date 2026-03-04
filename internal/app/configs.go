package app

import (
	"fmt"
	"log/slog"

	"github.com/knadh/koanf/v2"

	"saviour/internal/infra/logger"
	"saviour/internal/transport/rest"
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
