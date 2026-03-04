package app

import (
	"fmt"
	"sync"
)

// Container contains initialized runtime dependencies.
//
// It's supposed to simplify the app startup process
// and should not be directly accessed outside a component's Start method.
type Container sync.Map

func NewContainer() *Container {
	return &Container{}
}

type DependencyKey string

type Dependency[T any] struct {
	key DependencyKey
}

func DefineDependency[T any](key DependencyKey) *Dependency[T] {
	return &Dependency[T]{
		key: key,
	}
}

func (d *Dependency[T]) Set(di *Container, value T) {
	(*sync.Map)(di).Store(d.key, value)
}

func (d *Dependency[T]) Get(di *Container) (T, bool) {
	v, found := (*sync.Map)(di).Load(d.key)
	if found {
		res, ok := v.(T)
		if !ok {
			panic(
				fmt.Sprintf(
					"di: unexpected type %T under the key %s",
					v,
					d.key,
				),
			)
		}

		return res, found
	}

	var zero T

	return zero, false
}

// MustGet retrieves a dependency from a container or panics.
// It's only meant to be used inside a component's Start method.
func (d *Dependency[T]) MustGet(di *Container) T {
	v, ok := d.Get(di)

	if !ok {
		panic(
			fmt.Sprintf(
				"di: key '%s' not found "+
					"(most likely the dependency is accessed before its initialization)",
				d.key,
			),
		)
	}

	return v
}
