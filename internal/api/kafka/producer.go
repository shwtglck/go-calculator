
package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"newstart/internal/model"

	kafkago "github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafkago.Writer
}

func NewProducer(broker string) *Producer {
	return &Producer{
		writer: &kafkago.Writer{
			Addr:     kafkago.TCP(broker),
			Topic:    "calculations",
			Balancer: &kafkago.LeastBytes{},
		},
	}
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

func (p *Producer) SendCalculation(ctx context.Context, calc model.Calculation) error {
	data, err := json.Marshal(calc)
	if err != nil {
		return fmt.Errorf("marshal calculation: %w", err)
	}

	err = p.writer.WriteMessages(
		ctx,
		kafkago.Message{
			Value: data,
		},
	)

	if err != nil {
		return fmt.Errorf("send kafka message: %w", err)
	}

	return nil
}