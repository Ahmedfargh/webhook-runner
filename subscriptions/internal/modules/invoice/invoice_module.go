package invoice

import (
	"subscriptions/internal/modules/invoice/controller"
	"subscriptions/internal/modules/invoice/presenter"
	"subscriptions/internal/modules/invoice/repository"
	"subscriptions/internal/modules/invoice/service"

	"gorm.io/gorm"
)

type InvoiceModule struct {
	Repo       repository.InvoiceRepository
	Service    service.InvoiceService
	Presenter  presenter.InvoicePresenter
	Controller *controller.InvoiceController
}

func NewInvoiceModule(db *gorm.DB) *InvoiceModule {
	repo := repository.NewInvoiceRepository(db)
	svc := service.NewInvoiceService(repo)
	pres := presenter.NewInvoicePresenter()
	ctrl := controller.NewInvoiceController(svc, pres)

	return &InvoiceModule{
		Repo:       repo,
		Service:    svc,
		Presenter:  pres,
		Controller: ctrl,
	}
}
