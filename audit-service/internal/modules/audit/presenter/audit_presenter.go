package presenter

import (
	pb "auditService/api/proto/v1"
	"auditService/internal/models"
	"time"

	"github.com/google/uuid"
)

type AuditPresenter interface {
	ToProto(m *models.AuditLog) *pb.AuditLog
	ToProtoList(logs []models.AuditLog) []*pb.AuditLog
	FromRecordRequest(req *pb.RecordAuditLogRequest) *models.AuditLog
}

type auditPresenter struct{}

func NewAuditPresenter() AuditPresenter {
	return &auditPresenter{}
}

func (p *auditPresenter) ToProto(m *models.AuditLog) *pb.AuditLog {
	if m == nil {
		return nil
	}

	actorIDStr := ""
	if m.ActorID != uuid.Nil {
		actorIDStr = m.ActorID.String()
	}

	return &pb.AuditLog{
		Id:           m.ID.String(),
		ActorId:      actorIDStr,
		ActorType:    m.ActorType,
		ActorName:    m.ActorName,
		ActorEmail:   m.ActorEmail,
		ServiceName:  m.ServiceName,
		Action:       m.Action,
		Resource:     m.Resource,
		ResourceId:   m.ResourceID,
		BeforeJson:   m.BeforeJSON,
		AfterJson:    m.AfterJSON,
		IpAddress:    m.IPAddress,
		UserAgent:    m.UserAgent,
		Status:       m.Status,
		ErrorMessage: m.ErrorMessage,
		CreatedAt:    m.CreatedAt.Format(time.RFC3339),
	}
}

func (p *auditPresenter) ToProtoList(logs []models.AuditLog) []*pb.AuditLog {
	list := make([]*pb.AuditLog, len(logs))
	for i, l := range logs {
		list[i] = p.ToProto(&l)
	}
	return list
}

func (p *auditPresenter) FromRecordRequest(req *pb.RecordAuditLogRequest) *models.AuditLog {
	if req == nil {
		return nil
	}

	var actorID uuid.UUID
	if req.ActorId != "" {
		if parsed, err := uuid.Parse(req.ActorId); err == nil {
			actorID = parsed
		}
	}

	return &models.AuditLog{
		ID:           uuid.New(),
		ActorID:      actorID,
		ActorType:    req.ActorType,
		ActorName:    req.ActorName,
		ActorEmail:   req.ActorEmail,
		ServiceName:  req.ServiceName,
		Action:       req.Action,
		Resource:     req.Resource,
		ResourceID:   req.ResourceId,
		BeforeJSON:   req.BeforeJson,
		AfterJSON:    req.AfterJson,
		IPAddress:    req.IpAddress,
		UserAgent:    req.UserAgent,
		Status:       req.Status,
		ErrorMessage: req.ErrorMessage,
		CreatedAt:    time.Now().UTC(),
	}
}
