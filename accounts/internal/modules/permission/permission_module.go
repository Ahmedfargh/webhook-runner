package permission

import (
	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/modules/permission/controller"
	"accounts/internal/modules/permission/presenter"
	"accounts/internal/modules/permission/repository"
	"accounts/internal/modules/permission/service"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// PermissionModule encapsulates all layers of the Permission HMVC component
type PermissionModule struct {
	Repository repository.PermissionRepository
	Presenter  presenter.PermissionPresenter
	Service    service.PermissionService
	Controller *controller.PermissionController
}

// NewPermissionModule initializes and wires up the Permission HMVC module
func NewPermissionModule(db *gorm.DB) *PermissionModule {
	repo := repository.NewPermissionRepository(db)
	pres := presenter.NewPermissionPresenter()
	svc := service.NewPermissionService(repo)
	ctrl := controller.NewPermissionController(svc, pres)

	return &PermissionModule{
		Repository: repo,
		Presenter:  pres,
		Service:    svc,
		Controller: ctrl,
	}
}

// RegisterGRPC registers the module's controller with the gRPC server
func (m *PermissionModule) RegisterGRPC(server *grpc.Server) {
	accountsv1.RegisterPermissionServiceServer(server, m.Controller)
}
