package secretkey

import (
	"crypto/sha256"

	"saviour/pkg/secret"
)

type Config struct {
	Key secret.Secret[string]
}

func New(cfg Config) secret.Secret[[]byte] {
	sum := sha256.Sum256([]byte(cfg.Key.UnsafeValue()))
	return secret.New(sum[:]) // 32 bytes -> AES-256
}
