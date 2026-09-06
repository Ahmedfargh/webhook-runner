package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"requestTrackerService/internal/models"
	"requestTrackerService/internal/repository"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type TraceEvent struct {
	ID           string  `json:"id"`
	TraceID      string  `json:"trace_id"`
	RequestID    string  `json:"request_id"`
	ActorType    string  `json:"actor_type"`
	ActorID      string  `json:"actor_id"`
	ActorName    string  `json:"actor_name"`
	ActorEmail   string  `json:"actor_email"`
	ActorRole    string  `json:"actor_role"`
	ServiceName  string  `json:"service_name"`
	Method       string  `json:"method"`
	Path         string  `json:"path"`
	Route        string  `json:"route"`
	QueryParams  string  `json:"query_params"`
	ClientIP     string  `json:"client_ip"`
	UserAgent    string  `json:"user_agent"`
	StatusCode   int     `json:"status_code"`
	LifetimeMs   float64 `json:"lifetime_ms"`
	RequestBody  string  `json:"request_body"`
	ResponseBody string  `json:"response_body"`
	ErrorMessage string  `json:"error_message"`
	SpansJSON    string  `json:"spans_json"`
	ReceivedAt   string  `json:"received_at"`
	CompletedAt  string  `json:"completed_at"`
}

type Consumer struct {
	reader    *kafka.Reader
	traceRepo repository.TraceRepository
	brokers   []string
	topic     string
	groupID   string
	enabled   bool
}

func NewConsumer(
	brokersStr, topic, groupID string,
	enabled bool,
	traceRepo repository.TraceRepository,
) *Consumer {
	if !enabled || strings.TrimSpace(brokersStr) == "" {
		log.Println("[Request Tracker Kafka Consumer] Consumer is disabled or brokers not specified.")
		return &Consumer{enabled: false}
	}

	brokers := strings.Split(brokersStr, ",")
	for i, b := range brokers {
		brokers[i] = strings.TrimSpace(b)
	}

	if topic == "" {
		topic = "http-request-traces"
	}
	if groupID == "" {
		groupID = "request-tracker-group"
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        200 * time.Millisecond,
		CommitInterval: 500 * time.Millisecond,
		StartOffset:    kafka.LastOffset,
	})

	log.Printf("[Request Tracker Kafka Consumer] Initialized on topic '%s' (Group: '%s') on brokers: %v\n",
		topic, groupID, brokers)

	return &Consumer{
		reader:    reader,
		traceRepo: traceRepo,
		brokers:   brokers,
		topic:     topic,
		groupID:   groupID,
		enabled:   true,
	}
}

func (c *Consumer) IsEnabled() bool {
	return c != nil && c.enabled && c.reader != nil
}

func (c *Consumer) Start(ctx context.Context) {
	if !c.IsEnabled() {
		return
	}

	log.Printf("[Request Tracker Kafka Consumer] 🚀 Starting background trace event loop on topic '%s'...\n", c.topic)

	batch := make([]*models.RequestTrace, 0, 50)
	messagesToCommit := make([]kafka.Message, 0, 50)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := c.traceRepo.BatchCreate(ctx, batch); err != nil {
			log.Printf("[Request Tracker Kafka Consumer] ❌ Error batch-inserting %d traces: %v\n", len(batch), err)
		} else {
			log.Printf("[Request Tracker Kafka Consumer] 💾 Batch stored %d request traces successfully\n", len(batch))
		}
		if len(messagesToCommit) > 0 {
			_ = c.reader.CommitMessages(ctx, messagesToCommit...)
			messagesToCommit = messagesToCommit[:0]
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("[Request Tracker Kafka Consumer] Context cancelled. Flushing remaining traces...")
			flush()
			return

		case <-ticker.C:
			flush()

		default:
			readCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
			msg, err := c.reader.FetchMessage(readCtx)
			cancel()

			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					continue
				}
				if errors.Is(err, context.Canceled) {
					flush()
					return
				}
				log.Printf("[Request Tracker Kafka Consumer] Waiting for Kafka broker on '%s': %v (retrying)\n", c.topic, err)
				time.Sleep(2 * time.Second)
				continue
			}

			var event TraceEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("[Request Tracker Kafka Consumer] Invalid message JSON: %v\n", err)
				_ = c.reader.CommitMessages(ctx, msg)
				continue
			}

			var traceID uuid.UUID
			if event.ID != "" {
				if parsed, err := uuid.Parse(event.ID); err == nil {
					traceID = parsed
				} else {
					traceID = uuid.New()
				}
			} else {
				traceID = uuid.New()
			}

			receivedAt := time.Now().UTC()
			if event.ReceivedAt != "" {
				if parsed, err := time.Parse(time.RFC3339Nano, event.ReceivedAt); err == nil {
					receivedAt = parsed
				} else if parsed, err := time.Parse(time.RFC3339, event.ReceivedAt); err == nil {
					receivedAt = parsed
				}
			}

			completedAt := time.Now().UTC()
			if event.CompletedAt != "" {
				if parsed, err := time.Parse(time.RFC3339Nano, event.CompletedAt); err == nil {
					completedAt = parsed
				} else if parsed, err := time.Parse(time.RFC3339, event.CompletedAt); err == nil {
					completedAt = parsed
				}
			}

			record := &models.RequestTrace{
				ID:           traceID,
				TraceID:      event.TraceID,
				RequestID:    event.RequestID,
				ActorType:    event.ActorType,
				ActorID:      event.ActorID,
				ActorName:    event.ActorName,
				ActorEmail:   event.ActorEmail,
				ActorRole:    event.ActorRole,
				ServiceName:  event.ServiceName,
				Method:       event.Method,
				Path:         event.Path,
				Route:        event.Route,
				QueryParams:  event.QueryParams,
				ClientIP:     event.ClientIP,
				UserAgent:    event.UserAgent,
				StatusCode:   event.StatusCode,
				LifetimeMs:   event.LifetimeMs,
				RequestBody:  event.RequestBody,
				ResponseBody: event.ResponseBody,
				ErrorMessage: event.ErrorMessage,
				SpansJSON:    event.SpansJSON,
				ReceivedAt:   receivedAt,
				CompletedAt:  completedAt,
				CreatedAt:    time.Now().UTC(),
			}

			batch = append(batch, record)
			messagesToCommit = append(messagesToCommit, msg)

			if len(batch) >= 50 {
				flush()
			}
		}
	}
}

func (c *Consumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
