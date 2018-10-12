package service

import (
	"fmt"
	"github.com/microdevs/missy/log"
	"os"
	"sync"
)

var config *Configuration
var once sync.Once

// Config returns the singleton  instance of the service configuration
func Config() *Configuration {
	once.Do(func() {
		config = &Configuration{}
	})
	return config
}

// RegisterParameter registers a configuration parameter in the central service config
func (c *Configuration) RegisterParameter(envName string, defaultValue string, internalName string, mandatory bool, usage string) {
	c.Environment = append(
		c.Environment,
		EnvParameter{EnvName: envName, InternalName: internalName, DefaultValue: defaultValue, Mandatory: mandatory, Usage: usage},
	)
}

// RegisterMandatoryParameter is a shorthand to register a mandatory configuration parameter
func (c *Configuration) RegisterMandatoryParameter(envName string, internalName string, usage string) {
	c.RegisterParameter(envName, "", internalName, true, usage)
}

// RegisterOptionalParameter is a shorthand to register a optional configuration parameter with default value
func (c *Configuration) RegisterOptionalParameter(envName string, defaultValue string, internalName string, usage string) {
	c.RegisterParameter(envName, defaultValue, internalName, false, usage)
}

// Parse parses all configured environment variables according to configuration parameters and makes them accessible through
// internal names. It also checks if values have been set and if not sets default values. If parameter is not set but
// mandatory this function will collect all missing parameters in a list and exits the program with a usage message.
// This function can be called multiple times.
func (c *Configuration) Parse() {

	var failedParameters []EnvParameter
	// loop through registered parameters and try to find them in env
	for k, parameter := range c.Environment {

		if parameter.Parsed {
			continue
		}

		envValue, found := os.LookupEnv(parameter.EnvName)
		// if mandatory but not found add them to error list
		if found == false && parameter.Mandatory == true {
			failedParameters = append(failedParameters, parameter)
			continue
		}

		// Use default value if environment parameter is not set
		if found == false && parameter.Mandatory == false {
			log.Debugf("Using default value \"%s\" for variable %s - %s", parameter.DefaultValue, parameter.EnvName, parameter.Usage)
			c.Environment[k].Value = parameter.DefaultValue
			c.Environment[k].Parsed = true
			continue
		}

		// if environment parameter is set use the actual env value
		c.Environment[k].Value = envValue
		c.Environment[k].Parsed = true
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

// Get returns a value for a config parameter
func (c *Configuration) Get(internalName string) string {
	// loop through all environment parameters and look for the internal name
	// todo: enhance speed with an index if needed
	for _, ep := range c.Environment {
		if internalName == ep.EnvName {
			log.Warnf("You're trying to get a config value by the environment variable name: %s", ep.EnvName)
			log.Warnf("Please use the internal name instead: %s", ep.InternalName)
			return ""
		}

		if ep.InternalName == internalName {
			return ep.Value
		}
	}
	log.Warnf("You're trying to get unknown config value: %s", internalName)
	return ""
}
