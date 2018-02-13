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

type Reader interface {
	Read(msgFunc ReadMessageFunc) (io.Closer, error)
}

//go:generate mockgen -package=messaging -destination broker_reader_mock.go -source reader.go BrokerReader
type BrokerReader interface {
	FetchMessage(ctx context.Context) (Message, error)
	CommitMessages(ctx context.Context, msgs ...Message) error
	ReadMessage(ctx context.Context) (Message, error)
	io.Closer
}

type missyReader struct {
	brokers      []string
	groupID      string
	topic        string
	brokerReader BrokerReader
	readFunc     *ReadMessageFunc
}

type readMessenger struct {
	*kafka.Reader
}

//FetchMessage(ctx context.Context) (Message, error)
//CommitMessages(ctx context.Context, msgs ...Message) error
//ReadMessage(ctx context.Context) (Message, error)
//WriteMessages(ctx context.Context, msgs ...Message) error

func (rm *readMessenger) FetchMessage(ctx context.Context) (Message, error) {
	m, err := rm.Reader.FetchMessage(ctx)

	if err != nil {
		return Message{}, err
	}

	return Message{Topic: m.Topic, Key: m.Key, Value: m.Value, Time: m.Time, Partition: m.Partition, Offset: m.Offset}, nil
}

func (rm *readMessenger) ReadMessage(ctx context.Context) (Message, error) {
	m, err := rm.Reader.ReadMessage(ctx)

	if err != nil {
		return Message{}, err
	}

	return Message{Topic: m.Topic, Key: m.Key, Value: m.Value, Time: m.Time, Partition: m.Partition, Offset: m.Offset}, nil
}

func (rm *readMessenger) CommitMessages(ctx context.Context, msgs ...Message) error {

	kafkaMessages := make([]kafka.Message, len(msgs))

	for _, m := range msgs {
		kafkaMsg := kafka.Message{Topic: m.Topic, Key: m.Key, Value: m.Value, Time: m.Time, Partition: m.Partition, Offset: m.Offset}
		kafkaMessages = append(kafkaMessages, kafkaMsg)
	}

	return rm.Reader.CommitMessages(ctx, kafkaMessages...)
}

func (rm *readMessenger) Close() error {
	return rm.Reader.Close()
}

// NewReader based on brokers hosts, consumerGroup and topic. You need to close it after use. (Close())
// we are leaving using the missy config for now, because we don't know how we want to configure this yet.
func NewReader(brokers []string, groupID string, topic string) Reader {

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB do we want it from config?
		MaxBytes: 10e6, // 10MB do we want it frozm config?
	})

	return &missyReader{brokers: brokers, groupID: groupID, topic: topic, brokerReader: &readMessenger{kafkaReader}}
}

// Read start reading goroutine that calls msgFunc on new message, you need to close it after use
func (mr *missyReader) Read(msgFunc ReadMessageFunc) (io.Closer, error) {
	// we've got a read function on this reader, return error
	if mr.readFunc != nil {
		return mr.brokerReader, errors.New("this reader is currently reading from underlying broker")
	}

	// set current read func
	mr.readFunc = &msgFunc

	// start reading goroutine
	go func() {
		for {
			log.Infoln("# messaging # listening for new message")
			ctx := context.Background()

			m, err := mr.brokerReader.FetchMessage(ctx)
			if err != nil {
				break
			}

			log.Infof("# messaging # new message: [%v] %v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
			if err := msgFunc(m); err != nil {
				log.Errorf("# messaging # cannot commit a message: &v", err)
				continue
			}

			// commit message if no error
			if err := mr.brokerReader.CommitMessages(ctx, m); err != nil {
				log.Errorf("cannot commit message [%v] %v/%v: %s = %s; with error: %v", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value), err)
			}
		}
	}()

	return mr.brokerReader, nil
}

func (mr *missyReader) Close() error {
	return mr.brokerReader.Close()
}
