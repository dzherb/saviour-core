package config_test

import (
	"testing"

	"github.com/knadh/koanf/maps"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"saviour/internal/infra/config"
)

func TestSourcesPriority(t *testing.T) {
	t.Parallel()

	testCfg := map[string]any{
		"logger.level": "error",
		"test_key_1":   "1",
		"test_key_2":   "2",
	}

	s1 := &config.Source{
		Name:     "test_source1",
		Provider: confmap.Provider(testCfg, config.KeysDelimiter),
	}

	s2 := &config.Source{
		Name: "test_source1",
		Provider: confmap.Provider(
			map[string]any{
				"logger.level": "debug",
				"test_key_2":   "new",
			},
			config.KeysDelimiter,
		),
	}

	s3 := &config.Source{
		Name: "test_source1",
		Provider: confmap.Provider(
			map[string]any{
				"logger.level": "warn",
			},
			config.KeysDelimiter,
		),
	}

	cfg, err := config.LoadFromSources(s1, s2, s3)

	require.NoError(t, err)

	assert.Equal(
		t,
		testCfg["test_key_1"],
		cfg.Get("test_key_1"),
	)
	assert.Equal(
		t,
		"new",
		cfg.Get("test_key_2"),
	)
	assert.Equal(t, "warn", cfg.Get("logger.level"))
}

func TestEnvSource(t *testing.T) { //nolint:paralleltest
	kv := map[string]string{
		config.EnvPrefix + "SIMPLE_KEY":        "4",
		config.EnvPrefix + "PARENT__CHILD_KEY": "8",
		config.EnvPrefix + "SLICE":             "15 16 23",
		"NOT_PREFIXED":                         "42",
	}

	for k, v := range kv {
		t.Setenv(k, v)
	}

	res, err := config.EnvSource().Provider.Read()
	require.NoError(t, err)

	res, _ = maps.Flatten(res, nil, config.KeysDelimiter)

	assert.Equal(t, "4", res["simple_key"])
	assert.Equal(t, "8", res["parent.child_key"])
	assert.Equal(t, []string{"15", "16", "23"}, res["slice"])
	assert.Nil(t, res["not_prefixed"])
}

func TestBuildSources(t *testing.T) {
	t.Parallel()

	cfgPaths := []string{
		"a/b/c.json",
		"", // should be omitted
		"../d/e/f.json",
	}

	sources := config.BuildSources(cfgPaths)

	expectedSourceNames := []string{
		"defaults",
		"file",
		"file",
		"env",
	}

	for i, source := range sources {
		assert.Equal(t, expectedSourceNames[i], source.Name)
	}
}
