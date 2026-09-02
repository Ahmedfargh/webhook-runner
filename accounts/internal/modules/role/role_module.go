package role

import (
	accountsv1 "accounts/api/proto/v1"
	permPresenter "accounts/internal/modules/permission/presenter"
	permRepo "accounts/internal/modules/permission/repository"
	"accounts/internal/modules/role/controller"
	"accounts/internal/modules/role/presenter"
	"accounts/internal/modules/role/repository"
	"accounts/internal/modules/role/service"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// RoleModule encapsulates all layers of the Role HMVC component
type RoleModule struct {
	Repository repository.RoleRepository
	Presenter  presenter.RolePresenter
	Service    service.RoleService
	Controller *controller.RoleController
}

// NewRoleModule initializes and wires up the Role HMVC module
func NewRoleModule(db *gorm.DB, permRepository permRepo.PermissionRepository, permPresenter permPresenter.PermissionPresenter) *RoleModule {
	repo := repository.NewRoleRepository(db)
	pres := presenter.NewRolePresenter(permPresenter)
	svc := service.NewRoleService(repo, permRepository)
	ctrl := controller.NewRoleController(svc, pres)

	return &RoleModule{
		Repository: repo,
		Presenter:  pres,
		Service:    svc,
		Controller: ctrl,
	}
}

// RegisterGRPC registers the module's controller with the gRPC server
func (m *RoleModule) RegisterGRPC(server *grpc.Server) {
	accountsv1.RegisterRoleServiceServer(server, m.Controller)
}
