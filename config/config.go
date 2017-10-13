package config

import (
	"syscall"
	"fmt"
	"os"
	"github.com/microdevs/missy/log"
)

type Parameter struct {
	EnvName string
	Value string
	Mandatory bool
	Usage string
}

var parameters = map[string]*Parameter{}
var failedParameters = []*Parameter{}

func RegisterParameter(envName string, defaultValue string, internalName string, mandatory bool, usage string) {
	p := &Parameter{EnvName: envName, Value: defaultValue, Mandatory: mandatory, Usage: usage}
	parameters[internalName] = p
}

func Parse() {

	// loop through registered parameters and try to find them in env
	for _, parameter := range parameters {
		envValue, found := syscall.Getenv(parameter.EnvName)
		// if mandatory but not found add them to error list
		if found == false && parameter.Mandatory == true {
			failedParameters = append(failedParameters, parameter)
			continue
		}

		// Notify ops when non-mandatory parameter is using default value
		if found == false && parameter.Mandatory == false {
			log.Debugf("Using default value \"%s\" for variable %s - %s", parameter.Value, parameter.EnvName, parameter.Usage)
			continue
		}
		parameter.Value = envValue
	}

	// if parameters are missing, print errors and exit
	if len(failedParameters) > 0 {
		msg := "Mandatory config values are missing,\nplease set the following environment variable(s):\n\n"
		for _, fp := range failedParameters {
			msg = msg + fp.EnvName + " - " + fp.Usage + "\n"
		}
		fmt.Print(msg)
		os.Exit(1)
	}
}

func Get(name string) string { // todo needs error handling? or second return param like "found" true/false?
	p := parameters[name]
	return p.Value
}

