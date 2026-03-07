package secret_test

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"

	"saviour/internal/service/secret"
)

func TestEncryptDecrypt(t *testing.T) {
	t.Parallel()

	keySum := sha256.Sum256([]byte("secret_key"))
	secretKey := keySum[:]

	val, err := secret.Encrypt(secretKey, []byte("my_secret_value"))
	require.NoError(t, err)

	res, err := secret.Decrypt(secretKey, val)
	require.NoError(t, err)

	require.Equal(t, []byte("my_secret_value"), res)
}
