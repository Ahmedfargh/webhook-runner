package presenter

import (
	"encoding/json"
	"time"

	subscriptionsv1 "subscriptions/api/proto/v1"
	"subscriptions/internal/models"
)

type PlanPresenter interface {
	ToProto(plan *models.Plan) *subscriptionsv1.PlanResponse
	ToListProto(plans []models.Plan) *subscriptionsv1.ListPlansResponse
}

type planPresenter struct{}

func NewPlanPresenter() PlanPresenter {
	return &planPresenter{}
}

func (p *planPresenter) ToProto(plan *models.Plan) *subscriptionsv1.PlanResponse {
	if plan == nil {
		return nil
	}

	var features []string
	if plan.FeaturesJSON != "" {
		_ = json.Unmarshal([]byte(plan.FeaturesJSON), &features)
	}

	return &subscriptionsv1.PlanResponse{
		Id:                plan.ID.String(),
		Name:              plan.Name,
		Code:              plan.Code,
		Description:       plan.Description,
		PriceMonthly:      plan.PriceMonthly,
		PriceAnnually:     plan.PriceAnnually,
		Currency:          plan.Currency,
		MaxWebhooks:       plan.MaxWebhooks,
		MaxEventsPerMonth: plan.MaxEventsPerMonth,
		MaxTeamMembers:    plan.MaxTeamMembers,
		Features:          features,
		IsActive:          plan.IsActive,
		IsPopular:         plan.IsPopular,
		TierLevel:         plan.TierLevel,
		CreatedAt:         plan.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         plan.UpdatedAt.Format(time.RFC3339),
	}
}

func (p *planPresenter) ToListProto(plans []models.Plan) *subscriptionsv1.ListPlansResponse {
	res := make([]*subscriptionsv1.PlanResponse, 0, len(plans))
	for i := range plans {
		res = append(res, p.ToProto(&plans[i]))
	}
	return &subscriptionsv1.ListPlansResponse{
		Plans: res,
	}
}
