package kafka

import (
	"fmt"
	"net"
	"order_service/internal/config"
	"strconv"

	"github.com/segmentio/kafka-go"
)

func NewReader(conf config.KafkaConfig) *kafka.Reader {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{conf.Broker},
		GroupID:  conf.GroupID,
		Topic:    conf.Topic,
		MaxBytes: 10e6,
	})

	return reader
}

func CreateTopicIfNotExists(conf config.KafkaConfig) error {
	conn, err := kafka.Dial("tcp", conf.Broker)
	if err != nil {
		return fmt.Errorf("cannot connect to kafka Broker: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("cannot create a kafka controller: %w", err)
	}

	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("cannot connect to kafka Controller: %w", err)
	}

	defer controllerConn.Close()

	topicConfigs := kafka.TopicConfig{
		Topic:             conf.Topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}
	err = controllerConn.CreateTopics(topicConfigs)
	if err != nil {
		return fmt.Errorf("cannot create kafka topic: %w", err)
	}
	return nil
}

// func NewWriter(conf config.KafkaConfig) *kafka.Writer {
// 	w := kafka.NewWriter(kafka.WriterConfig{
// 		Brokers: []string{"localhost:9092"},
// 		Topic:   conf.Topic,
// 	})
// 	return w
// }
