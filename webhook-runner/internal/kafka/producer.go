package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type ResultProducer struct {
	writer  *kafka.Writer
	brokers []string
	topic   string
	enabled bool
}

func NewResultProducer(brokersStr, topic string, enabled bool) *ResultProducer {
	if !enabled || strings.TrimSpace(brokersStr) == "" {
		log.Println("[Kafka Result Producer] Kafka is disabled or brokers not specified.")
		return &ResultProducer{enabled: false}
	}

	brokers := strings.Split(brokersStr, ",")
	for i, b := range brokers {
		brokers[i] = strings.TrimSpace(b)
	}

	if topic == "" {
		topic = "webhook-results"
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        true,
	}

	log.Printf("[Kafka Result Producer] Initialized for topic '%s' on brokers: %v\n", topic, brokers)

	return &ResultProducer{
		writer:  writer,
		brokers: brokers,
		topic:   topic,
		enabled: true,
	}
}

func (p *ResultProducer) IsEnabled() bool {
	return p != nil && p.enabled && p.writer != nil
}

func (p *ResultProducer) PublishResult(ctx context.Context, result *WebhookResultEvent) error {
	if !p.IsEnabled() {
		return nil // gracefully skip if disabled
	}

	if result.CompletedAt.IsZero() {
		result.CompletedAt = time.Now().UTC()
	}

	payload, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook result event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(result.AppID),
		Value: payload,
		Time:  result.CompletedAt,
		Headers: []kafka.Header{
			{Key: "status", Value: []byte(result.Status)},
			{Key: "call_id", Value: []byte(result.CallID)},
			{Key: "event_name", Value: []byte(result.EventName)},
		},
	}

	writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := p.writer.WriteMessages(writeCtx, msg); err != nil {
		return fmt.Errorf("failed to publish result to topic '%s': %w", p.topic, err)
	}

	return nil
}

func (p *ResultProducer) Ping(ctx context.Context) error {
	if !p.IsEnabled() {
		return errors.New("kafka result producer is disabled")
	}

	if len(p.brokers) == 0 {
		return errors.New("no kafka brokers configured")
	}

	firstBroker := p.brokers[0]
	dialCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	conn, err := kafka.DialContext(dialCtx, "tcp", firstBroker)
	if err != nil {
		return fmt.Errorf("failed to connect to kafka broker %s: %w", firstBroker, err)
	}
	defer conn.Close()

	brokers, err := conn.Brokers()
	if err != nil {
		return fmt.Errorf("failed to fetch kafka broker metadata: %w", err)
	}

	if len(brokers) == 0 {
		return errors.New("zero kafka brokers discovered in cluster")
	}

	return nil
}

func (p *ResultProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
