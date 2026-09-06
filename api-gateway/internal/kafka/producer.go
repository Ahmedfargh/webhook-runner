package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer interface {
	PublishWebhookDispatch(ctx context.Context, event *WebhookDispatchEvent) error
	Ping(ctx context.Context) error
	Close() error
	IsEnabled() bool
}

type KafkaProducer struct {
	writer  *kafka.Writer
	brokers []string
	topic   string
	enabled bool
}

func NewKafkaProducer(brokersStr, topic string, enabled bool) *KafkaProducer {
	if !enabled || strings.TrimSpace(brokersStr) == "" {
		log.Println("[Kafka Producer] Kafka is disabled or brokers not specified.")
		return &KafkaProducer{enabled: false}
	}

	brokers := strings.Split(brokersStr, ",")
	for i, b := range brokers {
		brokers[i] = strings.TrimSpace(b)
	}

	if topic == "" {
		topic = "webhook-dispatches"
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}

	log.Printf("[Kafka Producer] Initialized Kafka Producer for topic '%s' on brokers: %v\n", topic, brokers)

	return &KafkaProducer{
		writer:  writer,
		brokers: brokers,
		topic:   topic,
		enabled: true,
	}
}

func (p *KafkaProducer) IsEnabled() bool {
	return p != nil && p.enabled && p.writer != nil
}

func (p *KafkaProducer) PublishWebhookDispatch(ctx context.Context, event *WebhookDispatchEvent) error {
	if !p.IsEnabled() {
		return errors.New("kafka producer is not enabled")
	}

	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().Unix()
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook dispatch event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.AppID),
		Value: payload,
		Time:  event.CreatedAt,
		Headers: []kafka.Header{
			{Key: "event_name", Value: []byte(event.EventName)},
			{Key: "call_id", Value: []byte(event.CallID)},
		},
	}

	writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := p.writer.WriteMessages(writeCtx, msg); err != nil {
		return fmt.Errorf("failed to publish message to kafka topic '%s': %w", p.topic, err)
	}

	return nil
}

func (p *KafkaProducer) Ping(ctx context.Context) error {
	if !p.IsEnabled() {
		return errors.New("kafka producer is disabled")
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

func (p *KafkaProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// CreateTopicIfNotExists helps ensure default topics exist on the broker
func CreateTopicIfNotExists(brokerAddr, topicName string, numPartitions int) error {
	conn, err := kafka.Dial("tcp", brokerAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topicName,
			NumPartitions:     numPartitions,
			ReplicationFactor: 1,
		},
	}

	return controllerConn.CreateTopics(topicConfigs...)
}
