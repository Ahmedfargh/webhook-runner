package user

import (
	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/modules/user/controller"
	"accounts/internal/modules/user/presenter"
	"accounts/internal/modules/user/repository"
	"accounts/internal/modules/user/service"
	repo "accounts/internal/repository"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// UserModule encapsulates all layers of the User HMVC component
type UserModule struct {
	Repository repository.UserRepository
	Presenter  presenter.UserPresenter
	Service    service.UserService
	Controller *controller.UserController
}

// NewUserModule initializes and wires up the User HMVC module
func NewUserModule(db *gorm.DB, countryRepo repo.CountryRepository) *UserModule {
	userRepo := repository.NewUserRepository(db)
	pres := presenter.NewUserPresenter()
	svc := service.NewUserService(userRepo, countryRepo)
	ctrl := controller.NewUserController(svc, pres)

	return &UserModule{
		Repository: userRepo,
		Presenter:  pres,
		Service:    svc,
		Controller: ctrl,
	}
}

// RegisterGRPC registers the module's controller with the gRPC server
func (m *UserModule) RegisterGRPC(server *grpc.Server) {
	accountsv1.RegisterUserServiceServer(server, m.Controller)
}
