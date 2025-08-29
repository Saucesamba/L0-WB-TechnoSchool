package kafka

import (
	"L0WB/internal/config"
	"L0WB/internal/db"
	"L0WB/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
)

type Consumer struct {
	reader *kafka.Reader
	db     db.Database
}

// создает нового консьюмера
func NewConsumer(cfg config.KafkaConfig, db db.Database) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{cfg.BrokerAddress},
		GroupID:        cfg.GroupID,
		Topic:          cfg.Topic,
		CommitInterval: 0,
		MaxAttempts:    3,
	})
	return &Consumer{reader: reader, db: db}, nil
}

// читает одно сообщение из Kafka
func (c *Consumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return kafka.Message{}, fmt.Errorf("failed to read message: %w", err)
	}
	return msg, nil
}

// закрывает Kafka-консьюмера
func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("failed to close reader: %w", err)
	}
	return nil
}

func (c *Consumer) ProcessMessage(ctx context.Context, msg kafka.Message) error {
	var order models.Order
	err := json.Unmarshal(msg.Value, &order)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	log.Printf("Received order: %s", order.OrderUID)

	err = c.db.CreateOrder(ctx, &order)
	if err != nil {
		return fmt.Errorf("failed to create order in database: %w", err)
	}

	log.Printf("Order %s processed successfully", order.OrderUID)
	return nil
}
