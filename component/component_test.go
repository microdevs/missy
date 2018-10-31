package component_test

import (
	"testing"

	. "github.com/microdevs/missy/component"
)

func TestRegister_Success(t *testing.T) {
	// given
	testComponent := Type("test")
	// unregister on exit
	defer Unregister(testComponent)

	// when
	env, err := Register(testComponent)

	// then
	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}

	if env == nil {
		t.Errorf("expected to get environment but got nil")
	}
}

func TestRegister_Error(t *testing.T) {
	// given
	testComponent := Type("test")
	// unregister on exit
	defer Unregister(testComponent)
	// first register
	Register(testComponent)

	// when
	env, err := Register(testComponent)

	// then
	if err != ErrAlreadyRegistered {
		t.Errorf("expected %v error but got: %v", ErrAlreadyRegistered, err)
	}

	if env != nil {
		t.Errorf("expected to get nil environment but got %v", env)
	}
}

func TestEnv_Success(t *testing.T) {
	// given
	testComponent := Type("test")
	// unregister on exit
	// first register
	Register(testComponent)
	defer Unregister(testComponent)

	// when
	env, err := Env(testComponent)
	if err != nil {
		t.Errorf("expected no error, but got %v", err)
	}

	if env == nil {
		t.Error("expected to get environment, but got nil")
	}
}

func TestEnv_Error(t *testing.T) {
	// given
	testComponent := Type("test")

	// when
	env, err := Env(testComponent)
	if err != ErrNotRegistered {
		t.Errorf("expected to get %v error, but got %v", ErrNotRegistered, err)
	}

	if env != nil {
		t.Errorf("expected to get nil environment, but got %v", env)
	}
}
