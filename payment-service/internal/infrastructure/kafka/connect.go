package kafka

import (
	"fmt"
	"net"
	"time"

	"github.com/IBM/sarama"
	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/config"
)

func NewProducer(cfg *config.Config) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Idempotent = true
	config.Producer.Timeout = 10 * time.Second
	config.Net.MaxOpenRequests = 1

	producer, err := sarama.NewSyncProducer(
		[]string{
			net.JoinHostPort(
				cfg.Secrets.Kafka.Host,
				fmt.Sprint(cfg.Secrets.Kafka.Port),
			),
		},
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return producer, nil
}

func Close(producer sarama.SyncProducer) error {
	if err := producer.Close(); err != nil {
		return fmt.Errorf("failed to close producer: %w", err)
	}

	return nil
}
