package messaging_test

import (
	"crypto/sha1"
	"testing"
	"time"

	. "github.com/microdevs/missy/messaging"
)

func TestMessage_Hash(t *testing.T) {
	message := Message{
		Topic:     "topicName",
		Key:       []byte("key"),
		Value:     []byte("value"),
		Time:      time.Now(),
		Partition: 0,
		Offset:    12,
	}

	hashBytes, err := message.Hash(sha1.New())

	if err != nil {
		t.Errorf("error during message hashing!: %v", err)
	}

	if len(hashBytes) == 0 {
		t.Error("hash bytes len is 0!")
	}
}

func TestMessage_HashString(t *testing.T) {
	message := Message{
		Topic:     "topicName",
		Key:       []byte("key"),
		Value:     []byte("value"),
		Time:      time.Now(),
		Partition: 0,
		Offset:    12,
	}

	hashString, err := message.HashString(sha1.New())

	if err != nil {
		t.Errorf("error during message hashing!: %v", err)
	}

	if len(hashString) == 0 {
		t.Error("hash string len is 0!")
	}
}

func TestMessage_Sha256String(t *testing.T) {
	message := Message{
		Topic:     "topicName",
		Key:       []byte("key"),
		Value:     []byte("value"),
		Time:      time.Now(),
		Partition: 0,
		Offset:    12,
	}

	hashString, err := message.Sha256String()

	if err != nil {
		t.Errorf("error during message hashing!: %v", err)
	}

	if len(hashString) == 0 {
		t.Error("hash string len is 0!")
	}
}
