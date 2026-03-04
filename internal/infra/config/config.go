package config

import (
	"fmt"

	"github.com/knadh/koanf/v2"
)

func Load(cfgPaths []string) (*koanf.Koanf, error) {
	if len(cfgPaths) == 0 {
		return nil, fmt.Errorf("no config file provided")
	}

	return loadFromSources(buildSources(cfgPaths)...)
}

func loadFromSources(sources ...*Source) (*koanf.Koanf, error) {
	var k = koanf.New(KeysDelimiter)

	for _, s := range sources {
		if err := k.Load(s.Provider, s.Parser); err != nil {
			return nil, fmt.Errorf(
				"error loading config from source %s: %w",
				s.Name,
				err,
			)
		}
	}

	return k, nil
}

func buildSources(cfgPaths []string) []*Source {
	var fileSources []*Source

	for _, cfgPath := range cfgPaths {
		if cfgPath != "" {
			fileSources = append(fileSources, FileSource(cfgPath))
		}
	}

	var sources []*Source

	// order matters, keys can be overridden
	sources = append(sources, DefaultsSource())
	sources = append(sources, fileSources...)
	sources = append(sources, EnvSource())

	return sources
}
