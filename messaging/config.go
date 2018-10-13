package messaging

import (
	"github.com/microdevs/missy/service"
	"strconv"
)

const defaultKafkaMaxRetries = 3
const defaultKafkaRetriesIntervalMS = 5000

// InitConfig registeres
func InitConfig() {
	service.Config().RegisterOptionalParameter("KAFKA_RETRIES_MAX_NUMBER", strconv.Itoa(defaultKafkaMaxRetries), "kafka.retries.max.number", "The number of times a kafka reader will retry")
	service.Config().RegisterOptionalParameter("KAFKA_RETRIES_INTERVAL_MS", strconv.Itoa(defaultKafkaRetriesIntervalMS), "kafka.retries.interval.ms", "The time between retries in a kafka reader in ms")
	service.Config().Parse()
}
