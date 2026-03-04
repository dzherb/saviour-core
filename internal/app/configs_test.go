package app_test

import (
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"saviour/internal/app"
	"saviour/internal/infra/config"
)

type configParser func(*koanf.Koanf) error

func wrapParser[T any](fn func(*koanf.Koanf) (T, error)) configParser {
	return func(k *koanf.Koanf) error {
		_, err := fn(k)
		return err
	}
}

// test anyway, make sure there's no panics.
func wrapParserNoErr[T any](fn func(*koanf.Koanf) T) configParser {
	return func(k *koanf.Koanf) error {
		_ = fn(k)
		return nil
	}
}

func TestConfigParsers_NoErrorsOnValidInput(t *testing.T) {
	t.Parallel()

	cfgParsers := []configParser{
		wrapParser(app.AppConfig),
		wrapParser(app.LoggerOptions),
		wrapParser(app.ServerConfig),
		wrapParserNoErr(app.APIConfig),
	}

	fileSource := config.FileSource("../../configs/config.dev.json")
	devCfg := koanf.New(config.KeysDelimiter)

	require.NoError(t, devCfg.Load(fileSource.Provider, fileSource.Parser))

	validConfigs := []*koanf.Koanf{
		devCfg,
	}

	for _, cfgParser := range cfgParsers {
		for _, cfg := range validConfigs {
			assert.NotPanics(t, func() {
				assert.NoError(t, cfgParser(cfg))
			})
		}
	}
}
