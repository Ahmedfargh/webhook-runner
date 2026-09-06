package kafka_test

import (
	"encoding/json"
	"testing"
	"time"

	"auditService/internal/kafka"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsumerDisabled(t *testing.T) {
	c := kafka.NewConsumer("", "audit-events", "test-group", false, nil)
	assert.False(t, c.IsEnabled())
	assert.NoError(t, c.Close())
}

func TestAuditEventSerialization(t *testing.T) {
	eventID := uuid.New().String()
	actorID := uuid.New().String()

	evt := kafka.AuditEvent{
		ID:          eventID,
		ActorID:     actorID,
		ActorType:   "ADMIN",
		ActorName:   "Admin User",
		ActorEmail:  "admin@test.com",
		ServiceName: "accounts",
		Action:      "DELETE",
		Resource:    "USER",
		ResourceID:  "u-123",
		BeforeJSON:  `{"status":"active"}`,
		AfterJSON:   `{"status":"deleted"}`,
		IPAddress:   "192.168.1.1",
		UserAgent:   "Mozilla/5.0",
		Status:      "SUCCESS",
		Timestamp:   time.Now().Unix(),
	}

	bytes, err := json.Marshal(evt)
	require.NoError(t, err)

	var parsed kafka.AuditEvent
	err = json.Unmarshal(bytes, &parsed)
	require.NoError(t, err)

	assert.Equal(t, eventID, parsed.ID)
	assert.Equal(t, actorID, parsed.ActorID)
	assert.Equal(t, "accounts", parsed.ServiceName)
	assert.Equal(t, "DELETE", parsed.Action)
	assert.Equal(t, "USER", parsed.Resource)
}
