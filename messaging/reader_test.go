package messaging

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestNewReader(t *testing.T) {
	r := NewReader([]string{"localhost:9091"}, "", "test")

	readerType := reflect.TypeOf((*Reader)(nil)).Elem()

	if !reflect.TypeOf(r).Implements(readerType) {
		t.Error("messaging.NewReader does not implement messaging.Reader interface")
	}

}

func TestReaderRead(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	brokerReaderMock := NewMockBrokerReader(mockCtrl)
	msg := &Message{Topic: "test", Key: []byte("key"), Value: []byte("value"), Partition: 0, Offset: 0}
	brokerReaderMock.EXPECT().FetchMessage(gomock.Any()).Return(*msg, nil)
	brokerReaderMock.EXPECT().CommitMessages(gomock.Any(), *msg).Return(nil)
	brokerReaderMock.EXPECT().Close().Return(nil)

	reader := missyReader{brokerReader: brokerReaderMock}

	readFunc := func(msg Message) error {
		if msg.Topic != "test" {
			t.Error(expected(msg.Topic, "test"))
		}
		return nil
	}

	closer, err := reader.Read(readFunc)

	defer closer.Close()

	if err != nil {
		t.Errorf("error during read function unexpected!")
	}

	_, err = reader.Read(readFunc)

	if err == nil {
		t.Errorf("error during read function expected, bacause readFunc is set!")
	}

	time.Sleep(time.Nanosecond)
}

func expected(expected string, value string) string {
	return fmt.Sprintf("expecting %s to be %s", expected, value)
}
