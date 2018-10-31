package messaging

import (
	"strconv"

	"github.com/microdevs/missy/component"
	"github.com/microdevs/missy/log"
)

const defaultMaxRetries = 3
const defaultRetriesIntervalMS = 5000

const Component = component.Type("messaging")

// initialize as component
func init() {
	// register component
	env, err := component.Register(Component)
	if err != nil {
		log.Panicf("cannot register component %s: %v", Component, err)
	}

	env.Set("KAFKA_RETRIES_MAX_NUMBER", strconv.Itoa(defaultMaxRetries))
	env.Set("KAFKA_RETRIES_INTERVAL_MS", strconv.Itoa(defaultRetriesIntervalMS))

	// print info
	log.Println(env.Info())
}
