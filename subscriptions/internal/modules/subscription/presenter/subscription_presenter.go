package presenter

import (
	"time"

	subscriptionsv1 "subscriptions/api/proto/v1"
	"subscriptions/internal/models"
	planPresenter "subscriptions/internal/modules/plan/presenter"
)

type SubscriptionPresenter interface {
	ToProto(sub *models.Subscription) *subscriptionsv1.SubscriptionResponse
	ToListProto(subs []models.Subscription, total int64, page, pageSize int) *subscriptionsv1.ListSubscriptionsResponse
}

type subscriptionPresenter struct {
	planPres planPresenter.PlanPresenter
}

func NewSubscriptionPresenter(planPres planPresenter.PlanPresenter) SubscriptionPresenter {
	return &subscriptionPresenter{planPres: planPres}
}

func (p *subscriptionPresenter) ToProto(sub *models.Subscription) *subscriptionsv1.SubscriptionResponse {
	if sub == nil {
		return nil
	}

	trialEndsAtStr := ""
	if sub.TrialEndsAt != nil {
		trialEndsAtStr = sub.TrialEndsAt.Format(time.RFC3339)
	}

	return &subscriptionsv1.SubscriptionResponse{
		Id:                 sub.ID.String(),
		UserId:             sub.UserID.String(),
		PlanId:             sub.PlanID.String(),
		Plan:               p.planPres.ToProto(&sub.Plan),
		Status:             string(sub.Status),
		BillingCycle:       string(sub.BillingCycle),
		CurrentPeriodStart: sub.CurrentPeriodStart.Format(time.RFC3339),
		CurrentPeriodEnd:   sub.CurrentPeriodEnd.Format(time.RFC3339),
		TrialEndsAt:        trialEndsAtStr,
		CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
		CustomNotes:        sub.CustomNotes,
		CreatedAt:          sub.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          sub.UpdatedAt.Format(time.RFC3339),
	}
}

func (p *subscriptionPresenter) ToListProto(subs []models.Subscription, total int64, page, pageSize int) *subscriptionsv1.ListSubscriptionsResponse {
	res := make([]*subscriptionsv1.SubscriptionResponse, 0, len(subs))
	for i := range subs {
		res = append(res, p.ToProto(&subs[i]))
	}

	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))

	return &subscriptionsv1.ListSubscriptionsResponse{
		Subscriptions: res,
		TotalItems:    total,
		CurrentPage:   int32(page),
		TotalPages:    totalPages,
	}
}
