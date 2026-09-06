package audit

import (
	pb "auditService/api/proto/v1"
	"auditService/internal/modules/audit/controller"
	"auditService/internal/modules/audit/presenter"
	"auditService/internal/modules/audit/repository"
	"auditService/internal/modules/audit/service"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type AuditModule struct {
	Repository repository.AuditRepository
	Service    service.AuditService
	Presenter  presenter.AuditPresenter
	Controller *controller.AuditController
}

func NewAuditModule(db *gorm.DB) *AuditModule {
	repo := repository.NewAuditRepository(db)
	pres := presenter.NewAuditPresenter()
	svc := service.NewAuditService(repo)
	ctrl := controller.NewAuditController(svc, pres)

	return &AuditModule{
		Repository: repo,
		Service:    svc,
		Presenter:  pres,
		Controller: ctrl,
	}
}

func (m *AuditModule) RegisterGRPC(server *grpc.Server) {
	pb.RegisterAuditServiceServer(server, m.Controller)
}
