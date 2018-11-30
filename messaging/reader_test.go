package messaging

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/microdevs/missy/service"

	"reflect"

	"github.com/bouk/monkey"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

func expected(expected string, value string) string {
	return fmt.Sprintf("expecting %s to be %s", expected, value)
}

func TestNewReader(t *testing.T) {
	monkeyPatchConfig()
	r, _ := NewReader([]string{"localhost:9091"}, "", "test", nil)
	readerType := reflect.TypeOf((*Reader)(nil)).Elem()
	if !reflect.TypeOf(r).Implements(readerType) {
		t.Error("messaging.NewReader does not implement messaging.Reader interface")
	}
}

func TestReader_ReadSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	brokerReaderMock := NewMockBrokerReader(mockCtrl)
	msg := &Message{Topic: "test", Key: []byte("key"), Value: []byte("value"), Partition: 0, Offset: 0}
	brokerReaderMock.EXPECT().FetchMessage(gomock.Any()).AnyTimes().Return(*msg, nil)
	brokerReaderMock.EXPECT().CommitMessages(gomock.Any(), *msg).AnyTimes().Return(nil)
	brokerReaderMock.EXPECT().Close().Return(nil)

	reader := missyReader{brokerReader: brokerReaderMock}

	readFunc := func(msg Message) error {
		if msg.Topic != "test" {
			t.Error(expected(msg.Topic, "test"))
		}
		if string(msg.Key) != "key" {
			t.Error(expected(string(msg.Key), "key"))
		}
		if string(msg.Value) != "value" {
			t.Error(expected(string(msg.Value), "value"))
		}
		if msg.Partition != 0 {
			t.Error(expected(string(msg.Partition), string(0)))
		}
		if msg.Offset != 0 {
			t.Error(expected(string(msg.Offset), string(0)))
		}
		return nil
	}

	err := reader.StartReading(context.Background(), readFunc)
	if err != nil {
		t.Errorf("error during read function unexpected!")
	}

	err = reader.StartReading(context.Background(), readFunc)
	if err == nil {
		t.Errorf("error during read function expected, because readFunc is set!")
	}
}

func TestReader_ReadErrorOnFetch(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	brokerReaderMock := NewMockBrokerReader(mockCtrl)
	msg := &Message{Topic: "test", Key: []byte("key"), Value: []byte("value"), Partition: 0, Offset: 0}
	brokerReaderMock.EXPECT().FetchMessage(gomock.Any()).AnyTimes().Return(*msg, errors.New("ferch error"))
	brokerReaderMock.EXPECT().Close().Return(nil)

	reader := missyReader{brokerReader: brokerReaderMock, maxRetries: 1}

	readFunc := func(msg Message) error {
		return nil
	}

	err := reader.StartReading(context.Background(), readFunc)
	if err != nil {
		t.Errorf("error during read function unexpected!")
	}

	// todo how to test internal goroutines loops better?
	time.Sleep(time.Millisecond)
}

func TestMissyReader_ReadErrorOnCommit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	brokerReaderMock := NewMockBrokerReader(mockCtrl)
	msg := &Message{Topic: "test", Key: []byte("key"), Value: []byte("value"), Partition: 0, Offset: 0}
	brokerReaderMock.EXPECT().FetchMessage(gomock.Any()).AnyTimes().Return(*msg, nil)
	brokerReaderMock.EXPECT().CommitMessages(gomock.Any(), *msg).AnyTimes().Return(errors.New("error"))

	reader := missyReader{brokerReader: brokerReaderMock}

	readFunc := func(msg Message) error {
		if msg.Topic != "test" {
			t.Error(expected(msg.Topic, "test"))
		}
		if string(msg.Key) != "key" {
			t.Error(expected(string(msg.Key), "key"))
		}
		if string(msg.Value) != "value" {
			t.Error(expected(string(msg.Value), "value"))
		}
		if msg.Partition != 0 {
			t.Error(expected(string(msg.Partition), string(0)))
		}
		if msg.Offset != 0 {
			t.Error(expected(string(msg.Offset), string(0)))
		}
		return nil
	}

	err := reader.StartReading(context.Background(), readFunc)
	if err != nil {
		t.Errorf("error during read function unexpected!")
	}

	// todo how to test internal goroutines loops better?
	time.Sleep(time.Millisecond)
}

func TestMissyReader_ReadErrorOnReadFuncShouldWriteToDLQAndCommitMsg(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	brokerReaderMock := NewMockBrokerReader(mockCtrl)
	key := []byte("key")
	value := []byte("value")
	msg := &Message{Topic: "test", Key: key, Value: value, Partition: 0, Offset: 0}
	brokerReaderMock.EXPECT().FetchMessage(gomock.Any()).AnyTimes().Return(*msg, nil)
	brokerReaderMock.EXPECT().Close().Return(nil)
	dlqWriterMock := NewMockWriter(mockCtrl)
	dlqWriterMock.EXPECT().Write(key, value).MinTimes(1).Return(nil)
	brokerReaderMock.EXPECT().CommitMessages(gomock.Any(), gomock.Any()).AnyTimes()
	reader := missyReader{brokerReader: brokerReaderMock, maxRetries: 1, dlqWriter: dlqWriterMock}

	readFunc := func(msg Message) error {
		return errors.New("error")
	}

	err := reader.StartReading(context.Background(), readFunc)
	if err != nil {
		t.Errorf("error during read function unexpected!")
	}
	// todo how to test internal goroutines loops better?
	time.Sleep(time.Millisecond)
}

func TestMissyReader_Close(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	brokerReaderMock := NewMockBrokerReader(mockCtrl)
	brokerReaderMock.EXPECT().Close().Return(nil)
	reader := missyReader{brokerReader: brokerReaderMock}
	err := reader.Close()
	if err != nil {
		t.Errorf("there is an error during Close call")
	}
}

func TestReadBroker_FetchMessage(t *testing.T) {
	monkeyPatchConfig()

	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9999"},
		GroupID: "gr1",
		Topic:   "test",
	})

	exec := false
	// using monkey patching to patch underlying function call (https://github.com/bouk/monkey)
	monkey.PatchInstanceMethod(reflect.TypeOf(kr), "FetchMessage", func(_ *kafka.Reader, ctx context.Context) (kafka.Message, error) {
		exec = true
		return kafka.Message{Topic: "test", Key: []byte("key"), Value: []byte("value"), Partition: 0, Offset: 0}, nil
	})
	defer monkey.Unpatch(kr.FetchMessage)

	rb := readBroker{kr}
	msg, err := rb.FetchMessage(context.Background())
	if !exec {
		t.Error("function patching was not called!")
	}
	if err != nil {
		t.Error("there is an error during ReadMessage call")
	}
	if msg.Topic != "test" {
		t.Error(expected(msg.Topic, "test"))
	}
	if string(msg.Key) != "key" {
		t.Error(expected(string(msg.Key), "key"))
	}
	if string(msg.Value) != "value" {
		t.Error(expected(string(msg.Value), "value"))
	}
	if msg.Partition != 0 {
		t.Error(expected(string(msg.Partition), string(0)))
	}
	if msg.Offset != 0 {
		t.Error(expected(string(msg.Offset), string(0)))
	}
}

func monkeyPatchConfig() {
	monkey.Patch(service.Config().Get, func(name string) string {
		return ""
	})
}

func TestReadBroker_FetchMessage_Error(t *testing.T) {
	monkeyPatchConfig()

	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9999"},
		GroupID: "gr1",
		Topic:   "test",
	})

	exec := false
	// using monkey patching to patch underlying function call (https://github.com/bouk/monkey)
	monkey.PatchInstanceMethod(reflect.TypeOf(kr), "FetchMessage", func(_ *kafka.Reader, ctx context.Context) (kafka.Message, error) {
		exec = true
		return kafka.Message{}, errors.New("fetch error")
	})
	defer monkey.Unpatch(kr.FetchMessage)

	rb := readBroker{kr}
	_, err := rb.FetchMessage(context.Background())
	if !exec {
		t.Error("function patching was not called!")
	}
	if err == nil {
		t.Error("there should be an error during FetchMessage call")
	}
}

func TestReadBroker_ReadMessage(t *testing.T) {
	monkeyPatchConfig()

	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9999"},
		GroupID: "gr1",
		Topic:   "test",
	})

	exec := false
	// using monkey patching to patch underlying function call (https://github.com/bouk/monkey)
	monkey.PatchInstanceMethod(reflect.TypeOf(kr), "ReadMessage", func(_ *kafka.Reader, ctx context.Context) (kafka.Message, error) {
		exec = true
		return kafka.Message{Topic: "test", Key: []byte("key"), Value: []byte("value"), Partition: 0, Offset: 0}, nil
	})
	defer monkey.Unpatch(kr.ReadMessage)

	rb := readBroker{kr}
	msg, err := rb.ReadMessage(context.Background())
	if !exec {
		t.Error("function patching was not called!")
	}
	if err != nil {
		t.Error("there is an error during ReadMessage call")
	}
	if msg.Topic != "test" {
		t.Error(expected(msg.Topic, "test"))
	}
	if string(msg.Key) != "key" {
		t.Error(expected(string(msg.Key), "key"))
	}
	if string(msg.Value) != "value" {
		t.Error(expected(string(msg.Value), "value"))
	}
	if msg.Partition != 0 {
		t.Error(expected(string(msg.Partition), string(0)))
	}
	if msg.Offset != 0 {
		t.Error(expected(string(msg.Offset), string(0)))
	}
}

func TestReadBroker_ReadMessage_Error(t *testing.T) {
	monkeyPatchConfig()
	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9999"},
		GroupID: "gr1",
		Topic:   "test",
	})

	exec := false
	// using monkey patching to patch underlying function call (https://github.com/bouk/monkey)
	monkey.PatchInstanceMethod(reflect.TypeOf(kr), "ReadMessage", func(_ *kafka.Reader, ctx context.Context) (kafka.Message, error) {
		exec = true
		return kafka.Message{}, errors.New("read error")
	})
	defer monkey.Unpatch(kr.ReadMessage)

	rb := readBroker{kr}
	_, err := rb.ReadMessage(context.Background())
	if !exec {
		t.Error("function patching was not called!")
	}
	if err == nil {
		t.Error("there should be an error during ReadMessage call")
	}
}

func TestReadBroker_CommitMessages(t *testing.T) {
	monkeyPatchConfig()
	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9999"},
		GroupID: "gr1",
		Topic:   "test",
	})

	messages := make([]Message, 1)
	messages[0] = Message{}

	exec := false
	// using monkey patching to patch underlying function call (https://github.com/bouk/monkey)
	monkey.PatchInstanceMethod(reflect.TypeOf(kr), "CommitMessages", func(_ *kafka.Reader, ctx context.Context, msgs ...kafka.Message) error {
		if len(messages) != len(msgs) {
			t.Errorf("not equal messages len: %s", expected(string(len(messages)), string(len(msgs))))
		}
		exec = true
		return nil
	})
	defer monkey.Unpatch(kr.CommitMessages)

	rb := readBroker{kr}
	err := rb.CommitMessages(context.Background(), messages...)
	if !exec {
		t.Error("function patching was not called!")
	}
	if err != nil {
		t.Error("error was not expected in CommitMessage call")
	}
}

func TestReadBroker_CommitMessages_Error(t *testing.T) {
	monkeyPatchConfig()
	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9999"},
		GroupID: "gr1",
		Topic:   "test",
	})

	messages := make([]Message, 1)
	messages[0] = Message{}
	exec := false
	// using monkey patching to patch underlying function call (https://github.com/bouk/monkey)
	monkey.PatchInstanceMethod(reflect.TypeOf(kr), "CommitMessages", func(_ *kafka.Reader, ctx context.Context, msgs ...kafka.Message) error {
		if len(messages) != len(msgs) {
			t.Errorf("not equal messages len: %s", expected(string(len(messages)), string(len(msgs))))
		}
		exec = true
		return errors.New("commit error")
	})
	defer monkey.Unpatch(kr.CommitMessages)

	rb := readBroker{kr}
	err := rb.CommitMessages(context.Background(), messages...)
	if !exec {
		t.Error("function patching was not called!")
	}
	if err == nil {
		t.Error("error was expected in CommitMessage call")
	}
}

func TestReadBroker_Close(t *testing.T) {
	monkeyPatchConfig()
	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9999"},
		GroupID: "gr1",
		Topic:   "test",
	})

	// using monkey patching to patch underlying function call (https://github.com/bouk/monkey)
	exec := false
	monkey.PatchInstanceMethod(reflect.TypeOf(kr), "Close", func(_ *kafka.Reader) error {
		exec = true
		return nil
	})
	defer monkey.Unpatch(kr.Close)

	rb := readBroker{kr}
	err := rb.Close()
	if !exec {
		t.Error("function patching was not called!")
	}
	if err != nil {
		t.Error("error was not expected in Close call")
	}
}

func TestReadBroker_Close_Error(t *testing.T) {
	monkeyPatchConfig()

	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9999"},
		GroupID: "gr1",
		Topic:   "test",
	})

	exec := false
	// using monkey patching to patch underlying function call (https://github.com/bouk/monkey)
	monkey.PatchInstanceMethod(reflect.TypeOf(kr), "Close", func(_ *kafka.Reader) error {
		exec = true
		return errors.New("close error")
	})

	defer monkey.Unpatch(kr.Close)

	rb := readBroker{kr}

	err := rb.Close()
	if !exec {
		t.Error("function patching was not called!")
	}
	if err == nil {
		t.Error("there should be error in Close call")
	}
}

func TestReadBroker_FailingMessageShouldBeSendToDLQ(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	brokerReaderMock := NewMockBrokerReader(mockCtrl)
	dlqWriterMock := NewMockWriter(mockCtrl)
	key := []byte("key")
	value := []byte("value")
	msg := &Message{Topic: "test", Key: key, Value: value, Partition: 0, Offset: 0}
	brokerReaderMock.EXPECT().CommitMessages(gomock.Any(), *msg).AnyTimes().Return(nil)
	brokerReaderMock.EXPECT().FetchMessage(gomock.Any()).AnyTimes().Return(*msg, nil)
	reader := &missyReader{brokerReader: brokerReaderMock, dlqWriter: dlqWriterMock, retriesInterval: 1 * time.Second, maxRetries: 1}
	readFunc := func(msg Message) error {
		return errors.New("error")
	}

	//main check, we just want to know that dlqWriterMock was called
	dlqWriterMock.EXPECT().Write(key, value).Return(nil)

	err := reader.StartReading(context.Background(), readFunc)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(3 * time.Second)
	mockCtrl.Finish()
}
