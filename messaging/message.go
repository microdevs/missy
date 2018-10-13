package messaging

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"hash"
	"time"
)

// Message is a Kafka message
type Message struct {
	Topic     string
	Key       []byte
	Value     []byte
	Time      time.Time
	Partition int
	Offset    int64
}

// Hash returns bytes array of a hash of a Message using provided hash mechanism
func (m Message) Hash(hash hash.Hash) ([]byte, error) {
	var binBuffer bytes.Buffer
	enc := gob.NewEncoder(&binBuffer)

	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	if _, err := hash.Write(binBuffer.Bytes()); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

// HashString returns string representation of a hash of a Message using provided hash mechanism
func (m Message) HashString(hash hash.Hash) (string, error) {
	hashBytes, err := m.Hash(hash)
	if err != nil {
		return "", err
	}

	// return hash encoded to string
	return hex.EncodeToString(hashBytes), nil
}

// Sha256 returns bytes array of a hash of a Message using Sha256 hash mechanism
func (m Message) Sha256() ([]byte, error) {
	return m.Hash(sha256.New())
}

// Sha256String returns string representation of a hash of a Message Sha256 hash mechanism
func (m Message) Sha256String() (string, error) {
	return m.HashString(sha256.New())
}
