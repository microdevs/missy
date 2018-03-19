package messaging

import (
	"context"
	"errors"
	"io"

	"strconv"
	"time"

	"github.com/microdevs/missy/config"
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
	brokers         []string
	groupID         string
	topic           string
	brokerReader    BrokerReader
	readFunc        *ReadMessageFunc
	dlqWriter       Writer
	maxRetries      int
	retriesInterval time.Duration
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

	retries, intervalTime := fetchRetriesAndInterval()

	log.Infof("Configured num of maxRetries: %v with interval %v", retries, intervalTime)

	return &missyReader{brokers: brokers,
		groupID:         groupID,
		topic:           topic,
		brokerReader:    &readBroker{kafkaReader},
		maxRetries:      retries,
		retriesInterval: intervalTime,
	}
}

func NewReaderWithDLQ(brokers []string, groupID string, topic string, dlqTopic string) Reader {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		CommitInterval: 0,    // 0 indicates that commits should be done synchronically
		MinBytes:       10e3, // 10KB do we want it from config?
		MaxBytes:       10e6, // 10MB do we want it from config?
	})

	retries, intervalTime := fetchRetriesAndInterval()
	log.Infof("Configured num of maxRetries: %v with interval %v", retries, intervalTime)

	if dlqTopic == "" {
		dlqTopic = topic + ".dlq"
		log.Debugf("Setting default dlq topic name because none was passed")
	}

	log.Infof("Configured %s as topic name for dead letter queue for topic %s", dlqTopic, topic)

	return &missyReader{brokers: brokers,
		groupID:         groupID,
		topic:           topic,
		brokerReader:    &readBroker{kafkaReader},
		dlqWriter:       NewWriter(brokers, dlqTopic),
		maxRetries:      retries,
		retriesInterval: intervalTime,
	}
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
	go func() {
		for {
			ctx := context.Background()

			m, err := mr.brokerReader.FetchMessage(ctx)
			if err != nil {
				break
			}

			log.Infof("# messaging # new message: [topic] %v; [part] %v; [offset] %v; %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
			if err := mr.processMessage(msgFunc, m, 0); err != nil {
				log.Errorf("# messaging # %v, sending message to dead letter queue", err)

				if mr.dlqWriter != nil {
					if err = mr.dlqWriter.Write(m.Key, m.Value); err != nil {
						log.Errorf("Sending message to dead letter queue failed because: %v", err)
					}
				}
				mr.commit(ctx, m)
				continue
			}

			// commit message if no error
			mr.commit(ctx, m)
		}
	}()

	return nil
}
func (mr *missyReader) commit(ctx context.Context, m Message) {
	if err := mr.brokerReader.CommitMessages(ctx, m); err != nil {
		// should we do something else to just logging not committed message?
		log.Errorf("cannot commit message [%s] %v/%v: %s = %s; with error: %v", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value), err)
	}
}

//will try to process same message for configured number of times
func (mr *missyReader) processMessage(msgFunc ReadMessageFunc, message Message, retryNumber int) error {

	if retryNumber > mr.maxRetries {
		return errors.New("reached maximum number of retries")
	}
	if err := msgFunc(message); err != nil {
		log.Errorf("# messaging # retry number %v failed, trying again", retryNumber)
		time.Sleep(mr.retriesInterval)
		return mr.processMessage(msgFunc, message, retryNumber+1)
	}
	return nil
}

// Close used to close underlying connection with broker
func (mr *missyReader) Close() error {
	return mr.brokerReader.Close()
}

func fetchRetriesAndInterval() (int, time.Duration) {
	retries, err := strconv.Atoi(config.Get("kafka.retries.max.number"))
	if retries <= 0 || err != nil {
		log.Debug("Setting retries number to 3, as kafka.retries.max.number was not set or wrong")
		retries = 3
	}
	var intervalTime time.Duration
	interval, err := strconv.Atoi(config.Get("kafka.retries.interval.ms"))
	if interval <= 0 || err != nil {
		log.Debug("Setting retries interval to 5000 ms, as kafka.retries.interval.ms was not set or wrong")
		intervalTime = 5000 * time.Millisecond
	}
	return retries, intervalTime
}
