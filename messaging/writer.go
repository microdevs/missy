package messaging

import (
	"context"
	"io"

	"github.com/segmentio/kafka-go"
)

type Writer interface {
	Write(key []byte, value []byte) error
	io.Closer
}

type missyWriter struct {
	brokers []string
	groupID string
	topic   string
	*kafka.Writer
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

	return &missyWriter{brokers: brokers, groupID: groupID, topic: topic, Writer: w}
}

// Read start reading goroutine that calls msgFunc on new message, you need to close it after use
func (mw *missyWriter) Write(key []byte, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
	}
	return mw.Writer.WriteMessages(context.Background(), msg)
}

func (mw *missyWriter) Close() error {
	return mw.Writer.Close()
}
