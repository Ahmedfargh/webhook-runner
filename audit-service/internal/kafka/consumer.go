package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"auditService/internal/models"
	"auditService/internal/modules/audit/service"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader       *kafka.Reader
	auditService service.AuditService
	brokers      []string
	topic        string
	groupID      string
	enabled      bool
}

func NewConsumer(
	brokersStr, topic, groupID string,
	enabled bool,
	auditService service.AuditService,
) *Consumer {
	if !enabled || strings.TrimSpace(brokersStr) == "" {
		log.Println("[Audit Kafka Consumer] Kafka consumer is disabled or brokers not specified.")
		return &Consumer{enabled: false}
	}

	brokers := strings.Split(brokersStr, ",")
	for i, b := range brokers {
		brokers[i] = strings.TrimSpace(b)
	}

	if topic == "" {
		topic = "audit-events"
	}
	if groupID == "" {
		groupID = "audit-service-group"
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

	log.Printf("[Audit Kafka Consumer] Initialized on topic '%s' (Group: '%s') on brokers: %v\n",
		topic, groupID, brokers)

	return &Consumer{
		reader:       reader,
		auditService: auditService,
		brokers:      brokers,
		topic:        topic,
		groupID:      groupID,
		enabled:      true,
	}
}

func (c *Consumer) IsEnabled() bool {
	return c != nil && c.enabled && c.reader != nil
}

func (c *Consumer) Start(ctx context.Context) {
	if !c.IsEnabled() {
		return
	}

	log.Printf("[Audit Kafka Consumer] 🚀 Starting background audit event loop on topic '%s'...\n", c.topic)

	for {
		select {
		case <-ctx.Done():
			log.Println("[Audit Kafka Consumer] Context cancelled. Shutting down event loop.")
			return
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				log.Printf("[Audit Kafka Consumer] Waiting for Kafka broker on topic '%s': %v (retrying in 5s)\n", c.topic, err)
				time.Sleep(5 * time.Second)
				continue
			}

			var event AuditEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("[Audit Kafka Consumer] Invalid message JSON: %v\n", err)
				_ = c.reader.CommitMessages(ctx, msg)
				continue
			}

			var actorID uuid.UUID
			if event.ActorID != "" {
				if parsed, err := uuid.Parse(event.ActorID); err == nil {
					actorID = parsed
				}
			}

			var logID uuid.UUID
			if event.ID != "" {
				if parsed, err := uuid.Parse(event.ID); err == nil {
					logID = parsed
				} else {
					logID = uuid.New()
				}
			} else {
				logID = uuid.New()
			}

			logEntry := &models.AuditLog{
				ID:           logID,
				ActorID:      actorID,
				ActorType:    event.ActorType,
				ActorName:    event.ActorName,
				ActorEmail:   event.ActorEmail,
				ServiceName:  event.ServiceName,
				Action:       event.Action,
				Resource:     event.Resource,
				ResourceID:   event.ResourceID,
				BeforeJSON:   event.BeforeJSON,
				AfterJSON:    event.AfterJSON,
				IPAddress:    event.IPAddress,
				UserAgent:    event.UserAgent,
				Status:       event.Status,
				ErrorMessage: event.ErrorMessage,
				CreatedAt:    time.Now().UTC(),
			}

			if _, err := c.auditService.RecordLog(ctx, logEntry); err != nil {
				log.Printf("[Audit Kafka Consumer] ❌ Failed to record audit log: %v\n", err)
			} else {
				log.Printf("[Audit Kafka Consumer] 📝 Recorded audit log: Service=%s, Action=%s, Resource=%s, Actor=%s\n",
					event.ServiceName, event.Action, event.Resource, event.ActorEmail)
			}

			_ = c.reader.CommitMessages(ctx, msg)
		}
	}
}

func (c *Consumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
