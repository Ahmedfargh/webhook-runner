package app

import (
	pb "webhookRunner/api/proto/v1"
	"webhookRunner/internal/modules/app/controller"
	"webhookRunner/internal/modules/app/presenter"
	"webhookRunner/internal/modules/app/repository"
	"webhookRunner/internal/modules/app/service"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type AppModule struct {
	Repository repository.AppRepository
	Service    service.AppService
	Presenter  presenter.AppPresenter
	Controller *controller.AppController
}

func NewAppModule(db *gorm.DB) *AppModule {
	repo := repository.NewAppRepository(db)
	pres := presenter.NewAppPresenter()
	svc := service.NewAppService(repo)
	ctrl := controller.NewAppController(svc, pres)

	return &AppModule{
		Repository: repo,
		Service:    svc,
		Presenter:  pres,
		Controller: ctrl,
	}
}

func (m *AppModule) RegisterGRPC(server *grpc.Server) {
	pb.RegisterAppServiceServer(server, m.Controller)
}
