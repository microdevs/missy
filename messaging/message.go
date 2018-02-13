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

////go:generate mockgen -package=messaging -destination mocks/messenger.go github.com/microdevs/missy/messaging Messenger
//type Messenger interface {
//	FetchMessage(ctx context.Context) (Message, error)
//	CommitMessages(ctx context.Context, msgs ...Message) error
//	ReadMessage(ctx context.Context) (Message, error)
//	WriteMessages(ctx context.Context, msgs ...Message) error
//	io.Closer
//}
