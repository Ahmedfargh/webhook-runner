package plan

import (
	"subscriptions/internal/modules/plan/controller"
	"subscriptions/internal/modules/plan/presenter"
	"subscriptions/internal/modules/plan/repository"
	"subscriptions/internal/modules/plan/service"

	"gorm.io/gorm"
)

type PlanModule struct {
	Repo       repository.PlanRepository
	Service    service.PlanService
	Presenter  presenter.PlanPresenter
	Controller *controller.PlanController
}

func NewPlanModule(db *gorm.DB) *PlanModule {
	repo := repository.NewPlanRepository(db)
	svc := service.NewPlanService(repo)
	pres := presenter.NewPlanPresenter()
	ctrl := controller.NewPlanController(svc, pres)

	return &PlanModule{
		Repo:       repo,
		Service:    svc,
		Presenter:  pres,
		Controller: ctrl,
	}
}
