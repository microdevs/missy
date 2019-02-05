package reader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/microdevs/missy/log"
	"github.com/microdevs/missy/messaging"
	"github.com/microdevs/missy/messaging/writer"
	kafka "github.com/segmentio/kafka-go"
)

const (
	// DefaultRetries defines the default number of reading message retries.
	DefaultRetries = 3
	// DefaultTimeInterval defines the interval between trying to consume the message once again in case of the error.
	DefaultTimeInterval = time.Second * 5
)

// MessageFunc is a message reading callback function, on error message will not be committed to underlying.
type MessageFunc func(msg messaging.Message) error

// Reader is used to read messages giving callback function.
type Reader interface {
	StartReading(ctx context.Context, msgFunc MessageFunc) error
	io.Closer
}

// Options defines variables that may change the behaviour of Kafka Reader.
type Options struct {
	// Retries defines how many times Kafka Reader should try to consume the message (if the consumer process fails).
	Retries int
	// IntervalTime defines how long reader should wait until reading the message again.
	// The process of reading message includes consuming it by provided ReadFunc.
	IntervalTime time.Duration
	// CommitInterval indicates the interval at which offsets are committed to the broker.
	// If 0, commits will be handled synchronously.
	CommitInterval time.Duration

	DLQEnabled bool
	// DLQTopic defines the topic of Kafka DLQ.
	// If the value is empty, but DLQ is enabled it will be set to the normal topic + '.dlq'.
	DLQTopic string
}

// New based on brokers hosts, consumerGroup and topic. You need to close it after use.
func New(brokers []string, groupID string, topic string, options *Options, l log.FieldsLogger) (reader Reader, err error) {
	l = l.WithField("topic", topic)
	retries := DefaultRetries
	if options != nil && options.Retries > 0 {
		retries = options.Retries
	}

	intervalTime := DefaultTimeInterval
	if options != nil && options.IntervalTime > 0 {
		intervalTime = options.IntervalTime
	}

	dlqTopic := ""
	if options != nil && options.DLQEnabled {
		if options.DLQTopic != "" {
			dlqTopic = options.DLQTopic
		} else {
			dlqTopic = topic + ".dlq"
		}
	}
	var dlqWriter writer.Writer
	if dlqTopic != "" {
		dlqWriter = writer.New(brokers, dlqTopic)
	}

	commitInterval := time.Duration(0)
	if options != nil && options.CommitInterval != 0 {
		commitInterval = options.CommitInterval
	}

	// WTF, the kafka.NewReader constructor panics if provided with wrong arguments. It's fucked up.
	// The library should never panic, like seriously. What's wrong with them.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		CommitInterval: commitInterval,
		MinBytes:       10e3, // 10KB do we want it from config?
		MaxBytes:       10e6, // 10MB do we want it from config?
	})
	reader = &missyReader{
		brokers:         brokers,
		brokerReader:    &readBroker{kafkaReader},
		topic:           topic,
		groupID:         groupID,
		maxRetries:      retries,
		retriesInterval: intervalTime,
		dlqWriter:       dlqWriter,
		l:               l,
	}
	return
}

type missyReader struct {
	brokers []string
	groupID string
	topic   string

	brokerReader brokerReader
	readFunc     *MessageFunc

	dlqWriter       writer.Writer
	maxRetries      int
	retriesInterval time.Duration

	isReading bool
	mutex     sync.Mutex

	l log.FieldsLogger
}

// StartReading start reading goroutine that calls msgFunc on new message, you need to close it after use
func (mr *missyReader) StartReading(ctx context.Context, msgFunc MessageFunc) error {
	mr.mutex.Lock()
	if mr.isReading {
		mr.mutex.Unlock()
		return errors.New("reader is already reading kafka messages (don't call StartReading twice)")
	}
	mr.isReading = true
	mr.mutex.Unlock()

	go func() {
		defer func() {
			mr.mutex.Lock()
			mr.isReading = false
			mr.mutex.Unlock()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				m, err := mr.brokerReader.FetchMessage(ctx)
				if err != nil {
					mr.l.Errorf("failed to fetch a message err: %v", err)
					break
				}

				mr.l.Debugf("new message (part=%d,offset=%d): %s = %s",
					m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
				if err := mr.processMessage(msgFunc, m, 0); err != nil {
					mr.l.Errorf("processing message err: %s", err)

					if mr.dlqWriter != nil {
						if err = mr.dlqWriter.Write(ctx, m.Key, m.Value); err != nil {
							mr.l.Errorf("sending message to DLQ err: %v", err)
						}
					}
					mr.commit(ctx, m)
					continue
				}
				mr.commit(ctx, m)
			}
		}
	}()
	return nil
}

func (mr *missyReader) commit(ctx context.Context, m messaging.Message) {
	if err := mr.brokerReader.CommitMessages(ctx, m); err != nil {
		mr.l.Errorf("(partition=%d,offset=%d) commiting message (key='%s',value='%s') err: %s",
			m.Partition, m.Offset, string(m.Key), string(m.Value), err)
	}
}

// processMessage tries to consume provided message with msgFunc.
// In case of error it will retry calling msgFunc mr.maxRetries times. Interval between retries is defined by
// mr.retriesInterval. You may set these parameters via Options passed to constructor.
func (mr *missyReader) processMessage(msgFunc MessageFunc, message messaging.Message, retryNumber int) error {
	if retryNumber > mr.maxRetries {
		return errors.New("reached maximum number of retries")
	}
	retryNumber++
	if err := msgFunc(message); err != nil {
		mr.l.Errorf("processing message retry number %d failed, err: %s", retryNumber, err)
		time.Sleep(mr.retriesInterval)
		return mr.processMessage(msgFunc, message, retryNumber)
	}
	return nil
}

// Close used to close underlying connection with broker
func (mr *missyReader) Close() error {
	return mr.brokerReader.Close()
}

// brokerReader interface used for underlying broker implementation.
//go:generate mockgen -package=messaging -destination broker_reader_mock.go -source reader.go BrokerReader
type brokerReader interface {
	FetchMessage(ctx context.Context) (messaging.Message, error)
	ReadMessage(ctx context.Context) (messaging.Message, error)
	CommitMessages(ctx context.Context, msgs ...messaging.Message) error
	io.Closer
}

// readBroker is a simple struct to cast kafka.Message to Message.
type readBroker struct {
	*kafka.Reader
}

// FetchMessages used to fetch messages from the broker
func (rm *readBroker) FetchMessage(ctx context.Context) (messaging.Message, error) {
	m, err := rm.Reader.FetchMessage(ctx)
	if err != nil {
		return messaging.Message{}, err
	}
	return messaging.Message(m), nil
}

// ReadMessage used to read and auto commit messages from the broker.
func (rm *readBroker) ReadMessage(ctx context.Context) (messaging.Message, error) {
	m, err := rm.Reader.ReadMessage(ctx)
	if err != nil {
		return messaging.Message{}, err
	}
	return messaging.Message(m), err
}

// CommitMessages used to commit red messages for the broker.
func (rm *readBroker) CommitMessages(ctx context.Context, msgs ...messaging.Message) error {
	messages := make([]kafka.Message, len(msgs))
	for i, m := range msgs {
		messages[i] = kafka.Message(m)
	}
	return rm.Reader.CommitMessages(ctx, messages...)
}

// Close used to close underlying connection with broker.
func (rm *readBroker) Close() error {
	return rm.Reader.Close()
}
