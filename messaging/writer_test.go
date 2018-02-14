package messaging

import (
	"context"
	"reflect"
	"testing"

	"github.com/bouk/monkey"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

func TestNewWriter(t *testing.T) {
	r := NewWriter([]string{"localhost:9091"}, "", "test")

	writerType := reflect.TypeOf((*Writer)(nil)).Elem()

	if !reflect.TypeOf(r).Implements(writerType) {
		t.Error("messaging.NewWriter does not implement messaging.Writer interface")
	}

}

func TestMissyWriter_WriteSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	brokerWriterMock := NewMockBrokerWriter(mockCtrl)
	msg := &Message{Key: []byte("key"), Value: []byte("value")}

	brokerWriterMock.EXPECT().WriteMessages(gomock.Any(), *msg).Return(nil)
	brokerWriterMock.EXPECT().Close().Return(nil)

	writer := missyWriter{brokerWriter: brokerWriterMock}

	key := []byte("key")
	value := []byte("value")

	if err := writer.Write(key, value); err != nil {
		t.Error("there was an unexpected error during Write message")
	}

	defer writer.Close()

}

func TestMissyWriter_WriteError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	brokerWriterMock := NewMockBrokerWriter(mockCtrl)
	msg := &Message{Key: []byte("key"), Value: []byte("value")}

	brokerWriterMock.EXPECT().WriteMessages(gomock.Any(), *msg).Return(errors.New("error"))
	brokerWriterMock.EXPECT().Close().Return(nil)

	writer := missyWriter{brokerWriter: brokerWriterMock}

	key := []byte("key")
	value := []byte("value")

	if err := writer.Write(key, value); err == nil {
		t.Error("error was expected")
	}

	defer writer.Close()

}

func TestMissyWriter_Close(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	brokerWriterMock := NewMockBrokerWriter(mockCtrl)
	brokerWriterMock.EXPECT().Close().Return(nil)

	writer := missyWriter{brokerWriter: brokerWriterMock}

	err := writer.Close()

	if err != nil {
		t.Errorf("there is an error during Close call")
	}

}

func TestWriteBroker_WriteMessages(t *testing.T) {
	kw := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9999"},
		Topic:   "test",
	})

	exec := false
	// using monkey patching to patch underlying function call (https://github.com/bouk/monkey)
	monkey.PatchInstanceMethod(reflect.TypeOf(kw), "WriteMessages", func(_ *kafka.Writer, ctx context.Context, messages ...kafka.Message) error {
		exec = true
		return nil
	})

	defer monkey.Unpatch(kw.WriteMessages)

	wb := writeBroker{kw}

	if err := wb.WriteMessages(context.Background(), Message{}); err != nil {
		t.Error("there is an unexpected error during WriteMessage call")
	}

	if !exec {
		t.Error("function patching was not called!")
	}
}

func TestWriteBroker_WriteMessages_Error(t *testing.T) {
	kw := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9999"},
		Topic:   "test",
	})

	exec := false
	// using monkey patching to patch underlying function call (https://github.com/bouk/monkey)
	monkey.PatchInstanceMethod(reflect.TypeOf(kw), "WriteMessages", func(_ *kafka.Writer, ctx context.Context, messages ...kafka.Message) error {
		exec = true
		return errors.New("error")
	})

	defer monkey.Unpatch(kw.WriteMessages)

	wb := writeBroker{kw}

	if err := wb.WriteMessages(context.Background(), Message{}); err == nil {
		t.Error("there should be an error during WriteMessage call")
	}

	if !exec {
		t.Error("function patching was not called!")
	}
}

func TestWriteBroker_Close(t *testing.T) {
	kw := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9999"},
		Topic:   "test",
	})

	exec := false
	// using monkey patching to patch underlying function call (https://github.com/bouk/monkey)
	monkey.PatchInstanceMethod(reflect.TypeOf(kw), "Close", func(_ *kafka.Writer) error {
		exec = true
		return nil
	})

	defer monkey.Unpatch(kw.Close)

	wb := writeBroker{kw}

	if err := wb.Close(); err != nil {
		t.Error("there is an unexpected error during WriteMessage call")
	}

	if !exec {
		t.Error("function patching was not called!")
	}

}
