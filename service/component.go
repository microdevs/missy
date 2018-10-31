package service

import (
	"strconv"

	"github.com/microdevs/missy/component"
	"github.com/microdevs/missy/log"
)

const Component = component.Type("service")

const (
	defaultListenHost   = "0.0.0.0"
	defaultListenPort   = 8080
	defaultJWTTokenFile = "keys/jwtRS256.rsa.pub"
	defaultLogFormat    = "json"
)

// initialize as component
func init() {

	// register component
	env, err := component.Register(Component)
	if err != nil {
		log.Panicf("cannot register component %s: %v", Component, err)
	}

	env.Set("LISTEN_HOST", defaultListenHost)
	env.Set("LISTEN_PORT", strconv.Itoa(defaultListenPort))
	env.Set("JWT_TOKEN_CA_FILE", defaultJWTTokenFile)
	env.Set("LOG_FORMAT", defaultLogFormat)

	// print info
	log.Println(env.Info())
}
