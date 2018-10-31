package component

import (
	"fmt"
	"os"
	"strings"

	"github.com/microdevs/missy/log"
)

// Environment is used to set/get key->val data
type Environment interface {
	Set(key, value string)
	Get(key string) string
	Info() string
}

// osEnvironment that uses os environment variables
type osEnvironment struct {
	data map[string]string
	info strings.Builder
}

// ensure osEnvironment is an Environment
var _ Environment = &osEnvironment{}

func newEnvironment(name Type) Environment {
	env := &osEnvironment{
		data: make(map[string]string),
		info: strings.Builder{},
	}
	// prepare info header
	env.info.WriteString("#################################################\n")
	env.info.WriteString(fmt.Sprintf("Component (%s) environment\n", name))
	env.info.WriteString("#################################################\n")
	env.info.WriteString("Environment variables:\n")
	env.info.WriteString("#################################################\n")
	return env
}

// Set environment variable, checks if os has one and uses it if true
func (e *osEnvironment) Set(key, value string) {
	toSet := value
	defer func() {
		e.data[key] = toSet
		msg := fmt.Sprintf("[%s] default[%s] current[%s]\n", key, value, toSet)
		e.info.WriteString(fmt.Sprintf("# %s #", msg))
		log.Infof("setting variable %s", msg)
	}()
	// check if we have os env already
	if val, ok := os.LookupEnv(key); ok {
		// we will set current as os val
		toSet = val
		log.Infoln(fmt.Sprintf("[%s] environment variable is found and will be used as [%s]", key, val))
	}
}

// Get env variable
func (e *osEnvironment) Get(key string) string {
	return e.data[key]
}

// Info gets Environment information
func (e *osEnvironment) Info() string {
	return e.info.String()
}
