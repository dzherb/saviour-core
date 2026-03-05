package secret_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"saviour/pkg/secret"
)

func TestSecretString(t *testing.T) {
	t.Parallel()

	s := secret.New("super_secret")

	assert.Equal(t, "super_secret", s.UnsafeString())
	assert.Equal(t, "*****", s.String())
	assert.Equal(t, "*****", s.LogValue().String())
	assert.Equal(t, "*****", fmt.Sprint(s))
	assert.Equal(t, "*****", fmt.Sprintf("%s", s))
	assert.Equal(t, "*****", fmt.Sprintf("%+v", s))
	assert.Equal(t, "*****", fmt.Sprintf("%#v", s))

	v, _ := json.Marshal(s) //nolint:staticcheck

	assert.Equal(t, "{}", string(v))
}
