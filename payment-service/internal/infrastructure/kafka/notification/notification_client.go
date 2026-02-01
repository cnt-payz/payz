package notificationkafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/cnt-payz/payz/payment-service/internal/domain/notification"
	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/config"
)

type NotificationClient struct {
	cfg      *config.Config
	producer sarama.SyncProducer
}

func New(cfg *config.Config, producer sarama.SyncProducer) *NotificationClient {
	return &NotificationClient{
		cfg:      cfg,
		producer: producer,
	}
}

func (nc *NotificationClient) SendMsg(ctx context.Context, msg *notification.Message) error {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	producerMsg := sarama.ProducerMessage{
		Topic: nc.cfg.Secrets.Kafka.Topic,
		Key:   sarama.StringEncoder(msg.TransactionID.String()),
		Value: sarama.ByteEncoder(bytes),
	}

	_, _, err = nc.producer.SendMessage(&producerMsg)
	if err != nil {
		return fmt.Errorf("failed to send msg: %w", err)
	}

	return nil
}
