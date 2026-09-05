package webhook

import (
	pb "webhookRunner/api/proto/v1"
	"webhookRunner/internal/engine"
	appRepo "webhookRunner/internal/modules/app/repository"
	"webhookRunner/internal/modules/webhook/controller"
	"webhookRunner/internal/modules/webhook/presenter"
	"webhookRunner/internal/modules/webhook/repository"
	"webhookRunner/internal/modules/webhook/service"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type WebhookModule struct {
	Repository repository.WebhookRepository
	Service    service.WebhookService
	Presenter  presenter.WebhookPresenter
	Controller *controller.WebhookController
}

func NewWebhookModule(db *gorm.DB, aRepo appRepo.AppRepository, dispatcher *engine.Dispatcher) *WebhookModule {
	repo := repository.NewWebhookRepository(db)
	pres := presenter.NewWebhookPresenter()
	svc := service.NewWebhookService(repo, aRepo, dispatcher)
	ctrl := controller.NewWebhookController(svc, pres)

	return &WebhookModule{
		Repository: repo,
		Service:    svc,
		Presenter:  pres,
		Controller: ctrl,
	}
}

func (m *WebhookModule) RegisterGRPC(server *grpc.Server) {
	pb.RegisterWebhookRunnerServiceServer(server, m.Controller)
}
