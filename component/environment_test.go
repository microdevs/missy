package component_test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/microdevs/missy/component"
)

func TestOsEnvironment_SetWithoutEnv(t *testing.T) {
	// given
	testComponent := Type("test")
	// unregister on exit
	// first register, omit error, tested in other place
	env, _ := Register(testComponent)
	defer Unregister(testComponent)

	// when
	env.Set("key", "val")

	// then
	if env.Get("key") != "val" {
		t.Errorf("expected to get val, but get %s", env.Get("key"))
	}
	fmt.Println(env.Info())
}

func TestOsEnvironment_SetWithEnv(t *testing.T) {
	// given
	os.Setenv("key", "envVal")
	testComponent := Type("test")
	// unregister on exit
	// first register, omit error, tested in other place
	env, _ := Register(testComponent)
	defer Unregister(testComponent)

	// when
	env.Set("key", "val")

	// then
	if env.Get("key") != "envVal" {
		t.Errorf("expected to get envVal, but get %s", env.Get("key"))
	}

	fmt.Println(env.Info())
}
