package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"newstart/internal/model"
)

const (
	defaultBroker = "localhost:9092"
	topic         = "calculations"
	consumerGroup = "storage-service"
)

// Repository описывает сохранение вычислений в хранилище.
type Repository interface {
	SaveCalculation(ctx context.Context, userID int, a, b float64, operator string, result float64) (model.Calculation, error)
}

// Consumer читает события из Kafka и сохраняет их через repository.
type Consumer struct {
	broker string
	reader *kafkago.Reader
	repo   Repository
}

// NewConsumer создаёт Kafka consumer для topic calculations.
func NewConsumer(broker string, repo Repository) *Consumer {
	if broker == "" {
		broker = defaultBroker
	}

	dialer := &kafkago.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		// Docker Kafka публикует в metadata внутренний адрес kafka:29092.
		// Storage-сервис запускается на хосте, поэтому перенаправляем на localhost:9092.
		DialFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			if host == "kafka" {
				address = net.JoinHostPort("localhost", "9092")
				log.Printf("kafka consumer: remap broker address kafka:* -> %s", address)
			}

			var d net.Dialer
			return d.DialContext(ctx, network, address)
		},
	}

	log.Printf(
		"kafka consumer: config broker=%s topic=%s group=%s startOffset=LastOffset minBytes=1 maxWait=1s",
		broker,
		topic,
		consumerGroup,
	)

	return &Consumer{
		broker: broker,
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:        []string{broker},
			Topic:          topic,
			GroupID:        consumerGroup,
			StartOffset:    kafkago.LastOffset,
			MinBytes:       1,
			MaxBytes:       10e6,
			MaxWait:        1 * time.Second,
			CommitInterval: 0,
			Dialer:         dialer,
		}),
		repo: repo,
	}
}

// Run непрерывно читает сообщения из Kafka и сохраняет их в базу.
func (c *Consumer) Run(ctx context.Context) error {
	for {
		log.Printf(
			"kafka consumer: waiting for message (broker=%s, topic=%s, group=%s)...",
			c.broker,
			topic,
			consumerGroup,
		)

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("kafka consumer: FetchMessage error: %v", err)
			return fmt.Errorf("read kafka message: %w", err)
		}

		log.Printf(
			"kafka consumer: message received partition=%d offset=%d size=%d bytes payload=%s",
			msg.Partition,
			msg.Offset,
			len(msg.Value),
			string(msg.Value),
		)

		if err := c.handleMessage(ctx, msg.Value); err != nil {
			log.Printf("kafka consumer: handle message error: %v", err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("kafka consumer: commit error: %v", err)
			return fmt.Errorf("commit kafka message: %w", err)
		}

		log.Printf(
			"kafka consumer: message committed partition=%d offset=%d",
			msg.Partition,
			msg.Offset,
		)
	}
}

func (c *Consumer) handleMessage(ctx context.Context, payload []byte) error {
	var calc model.Calculation
	if err := json.Unmarshal(payload, &calc); err != nil {
		log.Printf("kafka consumer: json unmarshal error: %v payload=%s", err, string(payload))
		return fmt.Errorf("parse calculation json: %w", err)
	}

	log.Printf(
		"kafka consumer: parsed calculation a=%v b=%v operator=%q result=%v",
		calc.OperandA,
		calc.OperandB,
		calc.Operator,
		calc.Result,
	)

	saved, err := c.repo.SaveCalculation(ctx, calc.UserID, calc.OperandA, calc.OperandB, calc.Operator, calc.Result)
	if err != nil {
		log.Printf("kafka consumer: save to database error: %v", err)
		return fmt.Errorf("save calculation: %w", err)
	}

	log.Printf(
		"kafka consumer: saved to database id=%d created_at=%s",
		saved.ID,
		saved.CreatedAt,
	)

	return nil
}

// Close закрывает соединение с Kafka.
func (c *Consumer) Close() error {
	return c.reader.Close()
}
