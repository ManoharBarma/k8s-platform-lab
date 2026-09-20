package kafka

import "github.com/segmentio/kafka-go"

// headerCarrier adapts Kafka message headers to the OpenTelemetry
// TextMapCarrier interface so trace context propagates from producers to
// consumers.
type headerCarrier struct {
	headers []kafka.Header
}

func (c *headerCarrier) Get(key string) string {
	for _, h := range c.headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c *headerCarrier) Set(key, value string) {
	for i := range c.headers {
		if c.headers[i].Key == key {
			c.headers[i].Value = []byte(value)
			return
		}
	}
	c.headers = append(c.headers, kafka.Header{Key: key, Value: []byte(value)})
}

func (c *headerCarrier) Keys() []string {
	keys := make([]string, len(c.headers))
	for i, h := range c.headers {
		keys[i] = h.Key
	}
	return keys
}
