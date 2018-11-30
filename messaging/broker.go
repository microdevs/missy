package messaging

import (
	"fmt"
	"strings"
)

// Brokers are capable of creating Kafka Readers or Writers.
type Brokers struct {
	brokers []string
}

// NewBrokers initializes brokers.
func NewBrokers(brokersRaw string) (*Brokers, error) {
	brokers := strings.Split(brokersRaw, ",")
	if brokersRaw == "" || len(brokers) < 1 {
		return nil, fmt.Errorf("no brokers provided, value='%s'", brokersRaw)
	}
	return &Brokers{
		brokers: brokers,
	}, nil
}

// NewWriter creates new Kafka Writer which writes to provided topic.
func (b *Brokers) NewWriter(topic string) Writer {
	return NewWriter(b.brokers, topic)
}

// NewReader creates new Kafka Reader which reads from given topic.
// You may also provide additional variables that might be used in NewReader as well.
func (b *Brokers) NewReader(groupID string, topic string, options *ReaderOptions) (Reader, error) {
	return NewReader(b.brokers, groupID, topic, options)
}
