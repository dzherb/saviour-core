package app_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"saviour/internal/app"
)

func TestDependency_SetAndGet(t *testing.T) {
	t.Parallel()

	c := app.NewContainer()

	dep := app.DefineDependency[int]("answer")
	dep.Set(c, 42)

	v, ok := dep.Get(c)

	require.True(t, ok)
	require.Equal(t, 42, v)
}

func TestDependency_Get_NotFound(t *testing.T) {
	t.Parallel()

	c := app.NewContainer()

	dep := app.DefineDependency[string]("missing")

	v, ok := dep.Get(c)

	require.False(t, ok)
	require.Equal(t, "", v) // zero value
}

func TestDependency_MustGet(t *testing.T) {
	t.Parallel()

	c := app.NewContainer()

	dep := app.DefineDependency[string]("name")
	dep.Set(c, "service-a")

	require.NotPanics(t, func() {
		require.Equal(t, "service-a", dep.MustGet(c))
	})
}

func TestDependency_MustGet_PanicsIfMissing(t *testing.T) {
	t.Parallel()

	c := app.NewContainer()

	dep := app.DefineDependency[int]("missing")

	require.Panics(t, func() {
		_ = dep.MustGet(c)
	})
}

func TestDependency_Get_PanicsOnWrongType(t *testing.T) {
	t.Parallel()

	c := app.NewContainer()

	// manually poison the container
	(*sync.Map)(c).Store(app.DependencyKey("value"), "not-an-int")

	dep := app.DefineDependency[int]("value")

	require.Panics(t, func() {
		_, _ = dep.Get(c)
	})
}

func TestDependency_DifferentKeysAreIsolated(t *testing.T) {
	t.Parallel()

	c := app.NewContainer()

	depInt := app.DefineDependency[int]("value")
	depStr := app.DefineDependency[string]("value_str")

	depInt.Set(c, 10)
	depStr.Set(c, "hello")

	v1, ok1 := depInt.Get(c)
	v2, ok2 := depStr.Get(c)

	require.True(t, ok1)
	require.True(t, ok2)

	require.Equal(t, 10, v1)
	require.Equal(t, "hello", v2)
}
