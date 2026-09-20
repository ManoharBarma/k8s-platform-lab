package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

// Handler processes a single consumed message. Returning an error leaves the
// message uncommitted so it is redelivered.
type Handler func(ctx context.Context, msg kafka.Message) error

// Consumer reads from a topic as part of a consumer group.
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer builds a Consumer for the given group and topic.
func NewConsumer(brokers []string, groupID, topic string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   topic,
		}),
	}
}

// Message is a transport-agnostic view of a consumed message, so consuming
// services don't need to import the underlying Kafka library.
type Message struct {
	Topic string
	Key   string
	Value []byte
}

// SimpleHandler processes a Message. Returning an error leaves it uncommitted.
type SimpleHandler func(ctx context.Context, m Message) error

// Run consumes with a SimpleHandler, hiding the underlying kafka.Message type.
func (c *Consumer) Run(ctx context.Context, h SimpleHandler) error {
	return c.Consume(ctx, func(ctx context.Context, msg kafka.Message) error {
		return h(ctx, Message{Topic: msg.Topic, Key: string(msg.Key), Value: msg.Value})
	})
}

// Consume loops until ctx is cancelled, extracting trace context from each
// message and invoking handler. Messages are committed only on success.
func (c *Consumer) Consume(ctx context.Context, handler Handler) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		carrier := &headerCarrier{headers: msg.Headers}
		msgCtx := otel.GetTextMapPropagator().Extract(ctx, carrier)

		if err := handler(msgCtx, msg); err != nil {
			// Leave uncommitted for redelivery.
			continue
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
	}
}

// Close closes the underlying reader.
func (c *Consumer) Close() error { return c.reader.Close() }
