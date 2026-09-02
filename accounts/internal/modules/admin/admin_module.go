package admin

import (
	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/modules/admin/controller"
	"accounts/internal/modules/admin/presenter"
	"accounts/internal/modules/admin/repository"
	"accounts/internal/modules/admin/service"
	permPresenter "accounts/internal/modules/permission/presenter"
	permRepo "accounts/internal/modules/permission/repository"
	rolePresenter "accounts/internal/modules/role/presenter"
	roleRepo "accounts/internal/modules/role/repository"
	repo "accounts/internal/repository"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// AdminModule encapsulates all layers of the Admin HMVC component
type AdminModule struct {
	Repository repository.AdminRepository
	Presenter  presenter.AdminPresenter
	Service    service.AdminService
	Controller *controller.AdminController
}

// NewAdminModule initializes and wires up the Admin HMVC module with its hierarchical dependencies
func NewAdminModule(
	db *gorm.DB,
	countryRepo repo.CountryRepository,
	roleRepository roleRepo.RoleRepository,
	permRepository permRepo.PermissionRepository,
	rolePres rolePresenter.RolePresenter,
	permPres permPresenter.PermissionPresenter,
) *AdminModule {
	adminRepo := repository.NewAdminRepository(db)
	pres := presenter.NewAdminPresenter(rolePres, permPres)
	svc := service.NewAdminService(adminRepo, countryRepo, roleRepository, permRepository)
	ctrl := controller.NewAdminController(svc, pres)

	return &AdminModule{
		Repository: adminRepo,
		Presenter:  pres,
		Service:    svc,
		Controller: ctrl,
	}
}

// RegisterGRPC registers the module's controller with the gRPC server
func (m *AdminModule) RegisterGRPC(server *grpc.Server) {
	accountsv1.RegisterAdminServiceServer(server, m.Controller)
}
