package messaging

import (
	"context"
	"io"

	"github.com/segmentio/kafka-go"
)

// Writer is used to write messages to underlying broker
type Writer interface {
	Write(key []byte, value []byte) error
	io.Closer
}

// BrokerWriter interface used for underlying broker implementation
//go:generate mockgen -package=messaging -destination broker_writer_mock.go -source writer.go BrokerWriter
type BrokerWriter interface {
	WriteMessages(ctx context.Context, msgs ...Message) error
	io.Closer
}

// missyWriter used as a default missy Writer implementation
type missyWriter struct {
	brokers      []string
	groupID      string
	topic        string
	brokerWriter BrokerWriter
}

// writeBroker us as a wrapper for kafka.Writer implementation to fulfill BrokerWriter interface
type writeBroker struct {
	*kafka.Writer
}

// WriteMessages used to write messages to the broker
func (wb *writeBroker) WriteMessages(ctx context.Context, msgs ...Message) error {

	kafkaMessages := make([]kafka.Message, len(msgs))

	for _, m := range msgs {
		kMessage := kafka.Message{Key: m.Key, Value: m.Value}
		kafkaMessages = append(kafkaMessages, kMessage)
	}

	return wb.Writer.WriteMessages(ctx, kafkaMessages...)
}

// Close used for closing underlying broker
func (wb *writeBroker) Close() error {
	return wb.Writer.Close()
}

// NewWriter based on brokers hosts, consumerGroup and topic. You need to close it after use. (Close())
// we are leaving using the missy config for now, because we don't know how we want to configure this yet.
func NewWriter(brokers []string, groupID string, topic string) Writer {

	// kafka writer
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})

	return &missyWriter{brokers: brokers, groupID: groupID, topic: topic, brokerWriter: &writeBroker{w}}
}

// Write new message
func (mw *missyWriter) Write(key []byte, value []byte) error {
	msg := Message{
		Key:   key,
		Value: value,
	}
	return mw.brokerWriter.WriteMessages(context.Background(), msg)
}

// Close writer after use
func (mw *missyWriter) Close() error {
	return mw.brokerWriter.Close()
}
