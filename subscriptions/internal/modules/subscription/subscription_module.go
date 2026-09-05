package subscription

import (
	invoiceService "subscriptions/internal/modules/invoice/service"
	planPresenter "subscriptions/internal/modules/plan/presenter"
	planRepo "subscriptions/internal/modules/plan/repository"
	"subscriptions/internal/modules/subscription/controller"
	"subscriptions/internal/modules/subscription/presenter"
	"subscriptions/internal/modules/subscription/repository"
	"subscriptions/internal/modules/subscription/service"

	"gorm.io/gorm"
)

type SubscriptionModule struct {
	Repo       repository.SubscriptionRepository
	Service    service.SubscriptionService
	Presenter  presenter.SubscriptionPresenter
	Controller *controller.SubscriptionController
}

func NewSubscriptionModule(
	db *gorm.DB,
	planRepo planRepo.PlanRepository,
	planPres planPresenter.PlanPresenter,
	invoiceSvc invoiceService.InvoiceService,
) *SubscriptionModule {
	repo := repository.NewSubscriptionRepository(db)
	svc := service.NewSubscriptionService(repo, planRepo, invoiceSvc)
	pres := presenter.NewSubscriptionPresenter(planPres)
	ctrl := controller.NewSubscriptionController(svc, pres)

	return &SubscriptionModule{
		Repo:       repo,
		Service:    svc,
		Presenter:  pres,
		Controller: ctrl,
	}
}
