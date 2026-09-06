package audit_test

import (
	"context"
	"testing"
	"time"

	pb "auditService/api/proto/v1"
	"auditService/internal/models"
	"auditService/internal/modules/audit"
	"auditService/internal/modules/audit/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.AuditLog{})
	require.NoError(t, err)

	return db
}

func TestAuditModule_CRUD_And_Listing(t *testing.T) {
	db := setupTestDB(t)
	mod := audit.NewAuditModule(db)
	ctx := context.Background()

	actorUUID := uuid.New()

	// 1. Record Audit Log
	recResp, err := mod.Controller.RecordAuditLog(ctx, &pb.RecordAuditLogRequest{
		ActorId:      actorUUID.String(),
		ActorType:    "ADMIN",
		ActorName:    "Super Admin",
		ActorEmail:   "admin@domain.com",
		ServiceName:  "accounts",
		Action:       "CREATE",
		Resource:     "USER",
		ResourceId:   "u-456",
		BeforeJson:   "",
		AfterJson:    `{"name":"John Doe"}`,
		IpAddress:    "127.0.0.1",
		UserAgent:    "Go-Test",
		Status:       "SUCCESS",
	})
	require.NoError(t, err)
	require.NotNil(t, recResp)
	assert.True(t, recResp.Success)
	assert.NotEmpty(t, recResp.AuditLog.Id)
	assert.Equal(t, actorUUID.String(), recResp.AuditLog.ActorId)
	assert.Equal(t, "accounts", recResp.AuditLog.ServiceName)

	logID := recResp.AuditLog.Id

	// 2. Get Audit Log by ID
	getResp, err := mod.Controller.GetAuditLog(ctx, &pb.GetAuditLogRequest{
		Id: logID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResp)
	assert.Equal(t, logID, getResp.AuditLog.Id)
	assert.Equal(t, "CREATE", getResp.AuditLog.Action)
	assert.Equal(t, "USER", getResp.AuditLog.Resource)

	// 3. List Audit Logs with filters
	listResp, err := mod.Controller.ListAuditLogs(ctx, &pb.ListAuditLogsRequest{
		Page:        1,
		Limit:       10,
		ServiceName: "accounts",
		Action:      "CREATE",
		Resource:    "USER",
	})
	require.NoError(t, err)
	require.NotNil(t, listResp)
	assert.Equal(t, int64(1), listResp.Total)
	assert.Len(t, listResp.AuditLogs, 1)

	// 4. Test Service Layer directly
	directLog := &models.AuditLog{
		ID:          uuid.New(),
		ActorEmail:  "direct@test.com",
		ServiceName: "subscriptions",
		Action:      "UPDATE",
		Resource:    "PLAN",
		ResourceID:  "p-1",
		Status:      "SUCCESS",
		CreatedAt:   time.Now().UTC(),
	}
	saved, err := mod.Service.RecordLog(ctx, directLog)
	require.NoError(t, err)
	assert.NotNil(t, saved)

	logs, total, err := mod.Service.ListLogs(ctx, repository.AuditFilter{
		ServiceName: "subscriptions",
		Page:        1,
		Limit:       10,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, logs, 1)
	assert.Equal(t, "subscriptions", logs[0].ServiceName)
}
