package manual_payment

import (
	invoiceService "subscriptions/internal/modules/invoice/service"
	"subscriptions/internal/modules/manual_payment/controller"
	"subscriptions/internal/modules/manual_payment/presenter"
	"subscriptions/internal/modules/manual_payment/repository"
	"subscriptions/internal/modules/manual_payment/service"
	subRepo "subscriptions/internal/modules/subscription/repository"

	"gorm.io/gorm"
)

type ManualPaymentModule struct {
	Repo       repository.ManualPaymentRepository
	Service    service.ManualPaymentService
	Presenter  presenter.ManualPaymentPresenter
	Controller *controller.ManualPaymentController
}

func NewManualPaymentModule(
	db *gorm.DB,
	invoiceSvc invoiceService.InvoiceService,
	subRepo subRepo.SubscriptionRepository,
) *ManualPaymentModule {
	repo := repository.NewManualPaymentRepository(db)
	svc := service.NewManualPaymentService(repo, invoiceSvc, subRepo)
	pres := presenter.NewManualPaymentPresenter()
	ctrl := controller.NewManualPaymentController(svc, pres)

	return &ManualPaymentModule{
		Repo:       repo,
		Service:    svc,
		Presenter:  pres,
		Controller: ctrl,
	}
}
