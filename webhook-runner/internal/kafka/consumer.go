package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"webhookRunner/internal/models"

	"github.com/segmentio/kafka-go"
)

type WebhookEventProcessor interface {
	ProcessAsyncDispatch(ctx context.Context, event *WebhookDispatchEvent) (*models.WebhookCall, error)
}

type Consumer struct {
	reader         *kafka.Reader
	processor      WebhookEventProcessor
	resultProducer *ResultProducer
	brokers        []string
	topic          string
	groupID        string
	enabled        bool
}

func NewConsumer(
	brokersStr, topic, groupID string,
	enabled bool,
	processor WebhookEventProcessor,
	resultProducer *ResultProducer,
) *Consumer {
	if !enabled || strings.TrimSpace(brokersStr) == "" {
		log.Println("[Kafka Consumer] Kafka consumer is disabled or brokers not specified.")
		return &Consumer{enabled: false}
	}

	brokers := strings.Split(brokersStr, ",")
	for i, b := range brokers {
		brokers[i] = strings.TrimSpace(b)
	}

	if topic == "" {
		topic = "webhook-dispatches"
	}
	if groupID == "" {
		groupID = "webhook-runner-group"
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        500 * time.Millisecond,
		CommitInterval: 1 * time.Second,
		StartOffset:    kafka.LastOffset,
	})

	log.Printf("[Kafka Consumer] Initialized Kafka Reader for topic '%s' (Group: '%s') on brokers: %v\n",
		topic, groupID, brokers)

	return &Consumer{
		reader:         reader,
		processor:      processor,
		resultProducer: resultProducer,
		brokers:        brokers,
		topic:          topic,
		groupID:        groupID,
		enabled:        true,
	}
}

func (c *Consumer) IsEnabled() bool {
	return c != nil && c.enabled && c.reader != nil
}

func (c *Consumer) Start(ctx context.Context) {
	if !c.IsEnabled() {
		return
	}

	log.Printf("[Kafka Consumer] 🚀 Starting background Kafka event loop on topic '%s'...\n", c.topic)

	for {
		select {
		case <-ctx.Done():
			log.Println("[Kafka Consumer] Context cancelled. Shutting down Kafka event loop.")
			return
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				log.Printf("[Kafka Consumer] Waiting for Kafka broker on topic '%s': %v (retrying in 5s)\n", c.topic, err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Process event
			var event WebhookDispatchEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("[Kafka Consumer] Invalid message JSON on partition %d offset %d: %v\n",
					msg.Partition, msg.Offset, err)
				_ = c.reader.CommitMessages(ctx, msg)
				continue
			}

			log.Printf("[Kafka Consumer] 📥 Received webhook dispatch event: CallID=%s, AppID=%s, Event=%s\n",
				event.CallID, event.AppID, event.EventName)

			call, procErr := c.processor.ProcessAsyncDispatch(ctx, &event)
			if procErr != nil {
				log.Printf("[Kafka Consumer] ❌ Error processing async dispatch for CallID=%s: %v\n", event.CallID, procErr)
			} else {
				log.Printf("[Kafka Consumer] ✅ Webhook dispatch completed for CallID=%s (Status=%s, HTTP=%d, Latency=%dms)\n",
					call.ID.String(), call.Status, call.ResponseStatusCode, call.ResponseLatencyMS)

				// Publish result event to results topic if configured
				if c.resultProducer != nil && c.resultProducer.IsEnabled() {
					resultEvent := &WebhookResultEvent{
						CallID:             call.ID.String(),
						AppID:              event.AppID,
						EventName:          call.EventName,
						TargetURL:          call.TargetURL,
						Status:             string(call.Status),
						ResponseStatusCode: call.ResponseStatusCode,
						ResponseBody:       call.ResponseBody,
						ResponseLatencyMS:  call.ResponseLatencyMS,
						ErrorMessage:       call.ErrorMessage,
						AttemptCount:       call.AttemptCount,
						CompletedAt:        time.Now().UTC(),
					}
					_ = c.resultProducer.PublishResult(ctx, resultEvent)
				}
			}

			// Commit offset after processing
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("[Kafka Consumer] Warning: Failed to commit offset for CallID=%s: %v\n", event.CallID, err)
			}
		}
	}
}

func (c *Consumer) Ping(ctx context.Context) error {
	if !c.IsEnabled() {
		return errors.New("kafka consumer is disabled")
	}

	if len(c.brokers) == 0 {
		return errors.New("no kafka brokers configured")
	}

	firstBroker := c.brokers[0]
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

func (c *Consumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
