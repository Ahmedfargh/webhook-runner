package country

import (
	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/modules/country/controller"
	"accounts/internal/modules/country/presenter"
	"accounts/internal/repository"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type CountryModule struct {
	Repository repository.CountryRepository
	Presenter  presenter.CountryPresenter
	Controller *controller.CountryController
}

func NewCountryModule(db *gorm.DB) *CountryModule {
	repo := repository.NewCountryRepository(db)
	pres := presenter.NewCountryPresenter()
	ctrl := controller.NewCountryController(repo, pres)

	return &CountryModule{
		Repository: repo,
		Presenter:  pres,
		Controller: ctrl,
	}
}

func (m *CountryModule) RegisterGRPC(server *grpc.Server) {
	accountsv1.RegisterCountryServiceServer(server, m.Controller)
}
