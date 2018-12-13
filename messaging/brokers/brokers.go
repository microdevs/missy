package brokers

import (
	"strings"

	"github.com/microdevs/missy/log"
	"github.com/microdevs/missy/messaging/reader"
	"github.com/microdevs/missy/messaging/writer"
	"github.com/pkg/errors"
)

// Brokers are capable of creating Kafka Readers or Writers.
type Brokers struct {
	Brokers []string `env:"KAFKA_BROKERS"`
}

// NewBrokers initializes brokers.
func NewBrokers(brokersRaw string) (*Brokers, error) {
	brokers := strings.Split(brokersRaw, ",")
	if brokersRaw == "" || len(brokers) < 1 {
		return nil, errors.Errorf("no brokers provided, value='%s'", brokersRaw)
	}
	return &Brokers{
		Brokers: brokers,
	}, nil
}

// NewWriter creates new Kafka Writer which writes to provided topic.
func (b Brokers) NewWriter(topic string) writer.Writer {
	return writer.New(b.Brokers, topic)
}

// NewReader creates new Kafka Reader which reads from given topic.
// You may also provide additional variables that might be used in NewReader as well.
func (b Brokers) NewReader(groupID string, topic string, options *reader.Options, l log.FieldsLogger) (reader.Reader, error) {
	return reader.New(b.Brokers, groupID, topic, options, l)
}
