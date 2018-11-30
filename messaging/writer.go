package messaging

import (
	"context"
	"io"
	"time"

	"github.com/segmentio/kafka-go"
)

// Writer is used to write messages to underlying broker.
type Writer interface {
	Write(ctx context.Context, key []byte, value []byte) error
	io.Closer
}

// missyWriter used as a default missy Writer implementation
type missyWriter struct {
	brokers      []string
	topic        string
	brokerWriter brokerWriter
}

// NewWriter based on brokers hosts, consumerGroup and topic. You need to close it after use. (Close())
// we are leaving using the missy config for now, because we don't know how we want to configure this yet.
func NewWriter(brokers []string, topic string) Writer {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	return &missyWriter{
		brokers:      brokers,
		topic:        topic,
		brokerWriter: &writeBroker{w},
	}
}

// Write new message to Kafka Writer.
func (mw *missyWriter) Write(ctx context.Context, key []byte, value []byte) error {
	msg := Message{
		Key:   key,
		Value: value,
		Time:  time.Now().UTC(),
	}
	return mw.brokerWriter.WriteMessages(ctx, msg)
}

// Close writer after use
func (mw *missyWriter) Close() error {
	return mw.brokerWriter.Close()
}

// brokerWriter interface used for underlying broker implementation.
//go:generate mockgen -package=messaging -destination broker_writer_mock.go -source writer.go BrokerWriter
type brokerWriter interface {
	WriteMessages(ctx context.Context, msgs ...Message) error
	io.Closer
}

// writeBroker us as a wrapper for kafka.Writer implementation to fulfill brokerWriter interface
type writeBroker struct {
	*kafka.Writer
}

// WriteMessages used to write messages to the broker
func (wb *writeBroker) WriteMessages(ctx context.Context, msgs ...Message) error {
	messages := make([]kafka.Message, len(msgs))
	for i, m := range msgs {
		messages[i] = kafka.Message(m)
	}
	return wb.Writer.WriteMessages(ctx, messages...)
}

// Close used for closing underlying broker
func (wb *writeBroker) Close() error {
	return wb.Writer.Close()
}
