package component

import (
	"sync"

	"github.com/pkg/errors"
)

// Type for a unique component
type Type string

// registeredComponents with a key->value map (can be used to set env. vars)
var registeredComponents map[Type]Environment

// once to ensure just one, synchronize call for register
var once sync.Once

var (
	ErrNotRegistered     = errors.New("component not registered")
	ErrAlreadyRegistered = errors.New("component already registered")
)

// Env returns component environment or error if not exists
func Env(component Type) (Environment, error) {
	if component, ok := registeredComponents[component]; ok {
		return component, nil
	}
	return nil, ErrNotRegistered
}

// Register registers new component and returns it's data as key->val map
func Register(component Type) (Environment, error) {
	// create registered components map just once
	once.Do(func() {
		registeredComponents = make(map[Type]Environment)
	})
	// check if we have registered this component
	if _, ok := registeredComponents[component]; ok {
		return nil, ErrAlreadyRegistered
	}
	componentData := newEnvironment(component)
	registeredComponents[component] = componentData
	return componentData, nil
}

// Unregister unregisters a component
func Unregister(component Type) {
	delete(registeredComponents, component)
}
