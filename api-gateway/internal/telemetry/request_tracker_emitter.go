package telemetry

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type TracePayload struct {
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

type RequestTrackerEmitter struct {
	writer      *kafka.Writer
	serviceName string
	enabled     bool
	ch          chan TracePayload
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

func NewRequestTrackerEmitter(brokersStr, topic, serviceName string, enabled bool) *RequestTrackerEmitter {
	if !enabled || strings.TrimSpace(brokersStr) == "" {
		return &RequestTrackerEmitter{serviceName: serviceName, enabled: false}
	}

	brokers := strings.Split(brokersStr, ",")
	for i, b := range brokers {
		brokers[i] = strings.TrimSpace(b)
	}

	if topic == "" {
		topic = "http-request-traces"
	}
	if serviceName == "" {
		serviceName = "api-gateway"
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 20 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        true,
	}

	emitter := &RequestTrackerEmitter{
		writer:      writer,
		serviceName: serviceName,
		enabled:     true,
		ch:          make(chan TracePayload, 50000),
		stopCh:      make(chan struct{}),
	}

	emitter.wg.Add(1)
	go emitter.worker()

	return emitter
}

func (e *RequestTrackerEmitter) worker() {
	defer e.wg.Done()
	batch := make([]kafka.Message, 0, 100)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 || e.writer == nil {
			return
		}
		if err := e.writer.WriteMessages(context.Background(), batch...); err != nil {
			log.Printf("[Request Tracker Emitter] Error writing batch to Kafka: %v\n", err)
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-e.stopCh:
			// Drain remaining messages
			for {
				select {
				case payload := <-e.ch:
					if bytes, err := json.Marshal(payload); err == nil {
						batch = append(batch, kafka.Message{
							Key:   []byte(payload.TraceID),
							Value: bytes,
						})
					}
				default:
					flush()
					return
				}
			}

		case <-ticker.C:
			flush()

		case payload := <-e.ch:
			bytes, err := json.Marshal(payload)
			if err != nil {
				continue
			}
			batch = append(batch, kafka.Message{
				Key:   []byte(payload.TraceID),
				Value: bytes,
			})
			if len(batch) >= 100 {
				flush()
			}
		}
	}
}

func (e *RequestTrackerEmitter) EmitAsync(payload TracePayload) {
	if !e.enabled || e.ch == nil {
		return
	}

	if payload.ID == "" {
		payload.ID = uuid.New().String()
	}
	if payload.ServiceName == "" {
		payload.ServiceName = e.serviceName
	}

	select {
	case e.ch <- payload:
	default:
		// Queue is full, avoid blocking gateway thread
		log.Println("[Request Tracker Emitter] Warning: trace buffer is full, dropping trace event")
	}
}

func (e *RequestTrackerEmitter) Close() error {
	if !e.enabled {
		return nil
	}
	close(e.stopCh)
	e.wg.Wait()
	if e.writer != nil {
		return e.writer.Close()
	}
	return nil
}
