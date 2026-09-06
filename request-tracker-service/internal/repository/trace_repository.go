package repository

import (
	"context"
	"strings"
	"time"

	"requestTrackerService/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TraceFilter struct {
	ActorType     string
	ActorID       string
	Method        string
	Route         string
	StatusCode    int
	MinLifetimeMs float64
	Search        string
	StartDate     string
	EndDate       string
	Page          int
	Limit         int
}

type TraceStats struct {
	TotalRequests int64
	AvgLifetimeMs float64
	P95LifetimeMs float64
	P99LifetimeMs float64
	ErrorCount    int64
	ErrorRate     float64
}

type TraceRepository interface {
	Create(ctx context.Context, trace *models.RequestTrace) error
	BatchCreate(ctx context.Context, traces []*models.RequestTrace) error
	FindByID(ctx context.Context, idOrTraceID string) (*models.RequestTrace, error)
	List(ctx context.Context, filter TraceFilter) ([]*models.RequestTrace, int64, error)
	GetStats(ctx context.Context, startDate, endDate string) (*TraceStats, error)
}

type traceRepository struct {
	db *gorm.DB
}

func NewTraceRepository(db *gorm.DB) TraceRepository {
	return &traceRepository{db: db}
}

func (r *traceRepository) Create(ctx context.Context, trace *models.RequestTrace) error {
	return r.db.WithContext(ctx).Create(trace).Error
}

func (r *traceRepository) BatchCreate(ctx context.Context, traces []*models.RequestTrace) error {
	if len(traces) == 0 {
		return nil
	}
	for _, t := range traces {
		if t.ID == uuid.Nil {
			t.ID = uuid.New()
		}
		if t.CreatedAt.IsZero() {
			t.CreatedAt = time.Now().UTC()
		}
	}
	return r.db.WithContext(ctx).CreateInBatches(traces, 100).Error
}

func (r *traceRepository) FindByID(ctx context.Context, idOrTraceID string) (*models.RequestTrace, error) {
	var trace models.RequestTrace
	query := r.db.WithContext(ctx)
	if _, err := uuid.Parse(idOrTraceID); err == nil {
		query = query.Where("id = ? OR trace_id = ?", idOrTraceID, idOrTraceID)
	} else {
		query = query.Where("trace_id = ? OR request_id = ?", idOrTraceID, idOrTraceID)
	}

	if err := query.First(&trace).Error; err != nil {
		return nil, err
	}
	return &trace, nil
}

func (r *traceRepository) List(ctx context.Context, filter TraceFilter) ([]*models.RequestTrace, int64, error) {
	var traces []*models.RequestTrace
	var total int64

	query := r.db.WithContext(ctx).Model(&models.RequestTrace{})

	if filter.ActorType != "" {
		query = query.Where("actor_type = ?", strings.ToUpper(filter.ActorType))
	}
	if filter.ActorID != "" {
		query = query.Where("actor_id = ?", filter.ActorID)
	}
	if filter.Method != "" {
		query = query.Where("method = ?", strings.ToUpper(filter.Method))
	}
	if filter.Route != "" {
		query = query.Where("route LIKE ?", "%"+filter.Route+"%")
	}
	if filter.StatusCode > 0 {
		query = query.Where("status_code = ?", filter.StatusCode)
	}
	if filter.MinLifetimeMs > 0 {
		query = query.Where("lifetime_ms >= ?", filter.MinLifetimeMs)
	}
	if filter.Search != "" {
		pattern := "%" + filter.Search + "%"
		query = query.Where("trace_id LIKE ? OR path LIKE ? OR actor_email LIKE ? OR actor_name LIKE ?", pattern, pattern, pattern, pattern)
	}
	if filter.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, filter.StartDate); err == nil {
			query = query.Where("received_at >= ?", t)
		}
	}
	if filter.EndDate != "" {
		if t, err := time.Parse(time.RFC3339, filter.EndDate); err == nil {
			query = query.Where("received_at <= ?", t)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	if err := query.Order("received_at DESC").Offset(offset).Limit(limit).Find(&traces).Error; err != nil {
		return nil, 0, err
	}

	return traces, total, nil
}

func (r *traceRepository) GetStats(ctx context.Context, startDate, endDate string) (*TraceStats, error) {
	query := r.db.WithContext(ctx).Model(&models.RequestTrace{})
	if startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			query = query.Where("received_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			query = query.Where("received_at <= ?", t)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if total == 0 {
		return &TraceStats{}, nil
	}

	var avgLifetime float64
	query.Select("COALESCE(AVG(lifetime_ms), 0)").Row().Scan(&avgLifetime)

	var errorCount int64
	query.Where("status_code >= 400").Count(&errorCount)

	errorRate := 0.0
	if total > 0 {
		errorRate = float64(errorCount) / float64(total) * 100.0
	}

	// Approximate P95 and P99 latency
	p95Offset := int64(float64(total) * 0.05)
	p99Offset := int64(float64(total) * 0.01)

	var p95Lifetime float64
	r.db.WithContext(ctx).Model(&models.RequestTrace{}).Order("lifetime_ms DESC").Offset(int(p95Offset)).Limit(1).Pluck("lifetime_ms", &p95Lifetime)

	var p99Lifetime float64
	r.db.WithContext(ctx).Model(&models.RequestTrace{}).Order("lifetime_ms DESC").Offset(int(p99Offset)).Limit(1).Pluck("lifetime_ms", &p99Lifetime)

	return &TraceStats{
		TotalRequests: total,
		AvgLifetimeMs: avgLifetime,
		P95LifetimeMs: p95Lifetime,
		P99LifetimeMs: p99Lifetime,
		ErrorCount:    errorCount,
		ErrorRate:     errorRate,
	}, nil
}
