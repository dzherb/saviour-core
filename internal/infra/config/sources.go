package config

import (
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Source struct {
	Name     string
	Provider koanf.Provider
	Parser   koanf.Parser
}

const KeysDelimiter = "."

var DefaultsSource = func() *Source {
	return &Source{
		Name:     "defaults",
		Provider: confmap.Provider(defaults, KeysDelimiter),
	}
}

func FileSource(path string) *Source {
	provider := file.Provider(path)

	return &Source{
		Name:     "file",
		Provider: provider,
		Parser:   json.Parser(),
	}
}

const EnvPrefix = "SAVIOUR__"

var EnvSource = func() *Source {
	return &Source{
		Name: "env",
		// transform env keys from SAVIOUR__PARENT__CHILD_KEY to parent.child_key
		Provider: env.Provider(
			KeysDelimiter,
			env.Opt{
				Prefix: EnvPrefix,
				TransformFunc: func(k, v string) (string, any) {
					k = strings.ReplaceAll(strings.ToLower(
						strings.TrimPrefix(k, EnvPrefix),
					), "__", KeysDelimiter)

					if strings.Contains(v, " ") {
						return k, strings.Split(v, " ")
					}

					return k, v
				},
			},
		),
	}
}
