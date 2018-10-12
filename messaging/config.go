package messaging

import "github.com/microdevs/missy/service"

func InitConfig() {
	service.Config().RegisterOptionalParameter("KAFKA_RETRIES_MAX_NUMBER", "3", "kafka.retries.max.number", "The number of times a kafka reader will retry")
	service.Config().RegisterOptionalParameter("KAFKA_RETRIES_INTERVAL_MS", "5000", "kafka.retries.interval.ms", "The time between retries in a kafka reader in ms")
}
