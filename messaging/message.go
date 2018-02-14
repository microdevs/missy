package messaging

import (
	"time"
)

type Message struct {
	Topic     string
	Key       []byte
	Value     []byte
	Time      time.Time
	Partition int
	Offset    int64
}
