package messaging

import (
	"context"
	"errors"
	"io"

	"github.com/microdevs/missy/log"
	"github.com/segmentio/kafka-go"
)

// ReadMessageFunc is a message reading callback function, on error message will not be committed to underlying
type ReadMessageFunc func(msg Message) error

// Reader is used to read messages giving callback function
type Reader interface {
	Read(msgFunc ReadMessageFunc) error
	io.Closer
}

// BrokerReader interface used for underlying broker implementation
//go:generate mockgen -package=messaging -destination broker_reader_mock.go -source reader.go BrokerReader
type BrokerReader interface {
	FetchMessage(ctx context.Context) (Message, error)
	CommitMessages(ctx context.Context, msgs ...Message) error
	ReadMessage(ctx context.Context) (Message, error)
	io.Closer
}

// missyReader used as a default missy Reader implementation
type missyReader struct {
	brokers      []string
	groupID      string
	topic        string
	brokerReader BrokerReader
	readFunc     *ReadMessageFunc
}

// readBroker us as a wrapper for kafka.Reader implementation to fulfill BrokerReader interface
type readBroker struct {
	*kafka.Reader
}

// FetchMessages used to fetch messages from the broker
func (rm *readBroker) FetchMessage(ctx context.Context) (Message, error) {
	m, err := rm.Reader.FetchMessage(ctx)

	if err != nil {
		return Message{}, err
	}

	return Message{Topic: m.Topic, Key: m.Key, Value: m.Value, Time: m.Time, Partition: m.Partition, Offset: m.Offset}, nil
}

// ReadMessage used to read and auto commit messages from the broker (currently not used in missy)
func (rm *readBroker) ReadMessage(ctx context.Context) (Message, error) {
	m, err := rm.Reader.ReadMessage(ctx)

	if err != nil {
		return Message{}, err
	}

	return Message{Topic: m.Topic, Key: m.Key, Value: m.Value, Time: m.Time, Partition: m.Partition, Offset: m.Offset}, nil
}

// CommitMessages used to commit red messages for the broker
func (rm *readBroker) CommitMessages(ctx context.Context, msgs ...Message) error {

	kafkaMessages := make([]kafka.Message, len(msgs))

	for i, m := range msgs {
		kafkaMsg := kafka.Message{Topic: m.Topic, Key: m.Key, Value: m.Value, Time: m.Time, Partition: m.Partition, Offset: m.Offset}
		kafkaMessages[i] = kafkaMsg
	}

	return rm.Reader.CommitMessages(ctx, kafkaMessages...)
}

// Close used to close underlying connection with broker
func (rm *readBroker) Close() error {
	return rm.Reader.Close()
}

// NewReader based on brokers hosts, consumerGroup and topic. You need to close it after use. (Close())
// we are leaving using the missy config for now, because we don't know how we want to configure this yet.
func NewReader(brokers []string, groupID string, topic string) Reader {

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		CommitInterval: 0,    // 0 indicates that commits should be done synchronically
		MinBytes:       10e3, // 10KB do we want it from config?
		MaxBytes:       10e6, // 10MB do we want it from config?
	})

	return &missyReader{brokers: brokers, groupID: groupID, topic: topic, brokerReader: &readBroker{kafkaReader}}
}

// Read start reading goroutine that calls msgFunc on new message, you need to close it after use
func (mr *missyReader) Read(msgFunc ReadMessageFunc) error {
	// we've got a read function on this reader, return error
	if mr.readFunc != nil {
		return errors.New("this reader is currently reading from underlying broker")
	}

	// set current read func
	mr.readFunc = &msgFunc

	// start reading goroutine
	for {
		ctx := context.Background()

		m, err := mr.brokerReader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		log.Infof("# messaging # new message: [topic] %v; [part] %v; [offset] %v; %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
		if err := msgFunc(m); err != nil {
			log.Errorf("# messaging # cannot commit a message: %v", err)
			continue
		}

		// commit message if no error
		if err := mr.brokerReader.CommitMessages(ctx, m); err != nil {
			// should we do something else to just logging not committed message?
			log.Errorf("cannot commit message [%s] %v/%v: %s = %s; with error: %v", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value), err)
		}
	}

	return nil
}

// Close used to close underlying connection with broker
func (mr *missyReader) Close() error {
	return mr.brokerReader.Close()
}
