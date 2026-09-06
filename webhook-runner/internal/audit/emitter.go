package audit

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Event struct {
	ID           string `json:"id"`
	ActorID      string `json:"actor_id"`
	ActorType    string `json:"actor_type"` // USER, ADMIN, SYSTEM, SERVICE
	ActorName    string `json:"actor_name"`
	ActorEmail   string `json:"actor_email"`
	ServiceName  string `json:"service_name"`
	Action       string `json:"action"`   // CREATE, UPDATE, DELETE, ROTATE_SECRET, RETRY, DISPATCH
	Resource     string `json:"resource"` // APP, WEBHOOK_CALL
	ResourceID   string `json:"resource_id"`
	BeforeJSON   string `json:"before_json"`
	AfterJSON    string `json:"after_json"`
	IPAddress    string `json:"ip_address"`
	UserAgent    string `json:"user_agent"`
	Status       string `json:"status"` // SUCCESS, FAILED
	ErrorMessage string `json:"error_message"`
	Timestamp    int64  `json:"timestamp"`
}

type Emitter interface {
	Emit(ctx context.Context, event Event)
	Close() error
}

type KafkaEmitter struct {
	writer      *kafka.Writer
	serviceName string
	enabled     bool
}

func NewEmitter(brokersStr, topic, serviceName string, enabled bool) *KafkaEmitter {
	if !enabled || strings.TrimSpace(brokersStr) == "" {
		return &KafkaEmitter{serviceName: serviceName, enabled: false}
	}

	brokers := strings.Split(brokersStr, ",")
	for i, b := range brokers {
		brokers[i] = strings.TrimSpace(b)
	}

	if topic == "" {
		topic = "audit-events"
	}
	if serviceName == "" {
		serviceName = "webhook-runner"
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        true,
	}

	return &KafkaEmitter{
		writer:      writer,
		serviceName: serviceName,
		enabled:     true,
	}
}

func (e *KafkaEmitter) Emit(ctx context.Context, event Event) {
	if !e.enabled || e.writer == nil {
		return
	}

	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.ServiceName == "" {
		event.ServiceName = e.serviceName
	}
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().Unix()
	}
	if event.Status == "" {
		event.Status = "SUCCESS"
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("[Audit Emitter] Error marshaling audit event: %v\n", err)
		return
	}

	msg := kafka.Message{
		Key:   []byte(event.Resource),
		Value: payload,
		Time:  time.Now().UTC(),
		Headers: []kafka.Header{
			{Key: "service_name", Value: []byte(event.ServiceName)},
			{Key: "action", Value: []byte(event.Action)},
		},
	}

	go func() {
		writeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := e.writer.WriteMessages(writeCtx, msg); err != nil {
			log.Printf("[Audit Emitter] Warning: Failed to publish audit event to Kafka: %v\n", err)
		}
	}()
}

func (e *KafkaEmitter) Close() error {
	if e.writer != nil {
		return e.writer.Close()
	}
	return nil
}
