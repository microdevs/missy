package config

import (
	"syscall"
	"fmt"
	"os"
	"github.com/microdevs/missy/log"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"sync"
)

const MissyConfigFile = ".missy.yml"

var config *Config
var once sync.Once

// return a signleton instance of service config
func GetInstance() *Config {

	once.Do(func() {
		// load config file todo: enable json
		cb, fileErr := readDefaultFile()
		if fileErr != nil {
			log.Fatalf("Cannot read config file %s with error: \"%s\"", MissyConfigFile, fileErr)
		}
		parseErr := parseConfigYAML(cb)
		if parseErr != nil {
			log.Fatalf("Cannot parse config file %s with error: \"%s\"", MissyConfigFile, parseErr)
		}

		config.ParseEnv()
	})

	return config
}

func Get(name string) string {
	return GetInstance().Get(name)
}

// parse yaml data to config struct
func parseConfigYAML(cb []byte) error {
	return yaml.Unmarshal(cb, &config)
}

// read default config file
func readDefaultFile() ([]byte, error) {
	return ioutil.ReadFile(MissyConfigFile)
}

// Parses all configured environment variables according to configuration to the internal names. Checks if values have
// been set and if not sets default values. If parameter is not set but mandatory this function will collect all missing
// parameters in a list and exits the program with a usage message.
func (c *Config) ParseEnv() {

	var failedParameters []EnvParameter
	// loop through registered parameters and try to find them in env
	for k, parameter := range config.Environment {
		envValue, found := syscall.Getenv(parameter.EnvName)
		// if mandatory but not found add them to error list
		if found == false && parameter.Mandatory == true {
			failedParameters = append(failedParameters, parameter)
			continue
		}

		// Use default value if environment parameter is not set
		if found == false && parameter.Mandatory == false {
			log.Debugf("Using default value \"%s\" for variable %s - %s", parameter.DefaultValue, parameter.EnvName, parameter.Usage)
			config.Environment[k].Value = parameter.DefaultValue
			continue
		}

		// if environment parameter is set use the actual env value
		config.Environment[k].Value = envValue
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

// returns a value for a config parameter
func (c *Config) Get(internalName string) string {
	// loop through all environment parameters and look for the internal name
	// todo: enhance speed with an index if needed
	for _, ep := range c.Environment {
		if ep.InternalName == internalName {
			return ep.Value
		}
	}
	return ""
}
