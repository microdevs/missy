package messaging

import (
	"strconv"
	"time"

	"github.com/microdevs/missy/service"
)

const defaultKafkaMaxRetries = 3
const defaultKafkaRetriesInterval = time.Second * 5
const defaultKafkaRetentionTime = time.Minute * 60 * 24 * 30

const (
	kafkaRetriesMaxNumber = "kafka.retries.max.number"
	kafkaRetriesInterval  = "kafka.retries.interval"
	kafkaRetentionTime    = "kafka.retention.time"
)

func init() {
	InitConfig()
}

func InitConfig() {
	cfg := service.Config()
	cfg.RegisterOptionalParameter("KAFKA_RETRIES_MAX_NUMBER", strconv.Itoa(defaultKafkaMaxRetries), kafkaRetriesMaxNumber, "The number of times a kafka reader will retry")
	cfg.RegisterOptionalParameter("KAFKA_RETRIES_INTERVAL", defaultKafkaRetriesInterval.String(), kafkaRetriesInterval, "The time between retries in a kafka reader")
	cfg.RegisterOptionalParameter("KAFKA_RETENTION_TIME", defaultKafkaRetentionTime.String(), kafkaRetentionTime, "Consumer retention duration on kafka broker, defaults to "+defaultKafkaRetentionTime.String())
	cfg.Parse()
}
