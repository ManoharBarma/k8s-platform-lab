// Package kafka provides thin producer/consumer wrappers over segmentio/kafka-go
// with OpenTelemetry trace-context propagation baked into message headers.
package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

// Producer publishes messages to Kafka with trace context injected into headers.
type Producer struct {
	writer *kafka.Writer
}

// NewProducer builds a Producer that hashes keys across partitions and waits
// for all in-sync replicas to acknowledge writes (durability).
func NewProducer(brokers []string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireAll,
			WriteTimeout: 10 * time.Second,
			Async:        false,
		},
	}
}

// Publish writes a single message, propagating the active trace context.
func (p *Producer) Publish(ctx context.Context, topic, key string, value []byte) error {
	carrier := &headerCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic:   topic,
		Key:     []byte(key),
		Value:   value,
		Headers: carrier.headers,
	})
}

// Close flushes and closes the underlying writer.
func (p *Producer) Close() error { return p.writer.Close() }
