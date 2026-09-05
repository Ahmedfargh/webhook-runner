package controller

import (
	"context"

	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/modules/country/presenter"
	"accounts/internal/repository"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CountryController struct {
	accountsv1.UnimplementedCountryServiceServer
	repo      repository.CountryRepository
	presenter presenter.CountryPresenter
}

func NewCountryController(repo repository.CountryRepository, presenter presenter.CountryPresenter) *CountryController {
	return &CountryController{
		repo:      repo,
		presenter: presenter,
	}
}

func (c *CountryController) ListCountries(ctx context.Context, req *accountsv1.ListCountriesRequest) (*accountsv1.ListCountriesResponse, error) {
	countries, err := c.repo.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list countries: %v", err)
	}

	return &accountsv1.ListCountriesResponse{
		Countries: c.presenter.ToProtoList(countries),
	}, nil
}

func (c *CountryController) GetCountry(ctx context.Context, req *accountsv1.GetCountryRequest) (*accountsv1.CountryResponse, error) {
	if req.Id != "" {
		id, err := uuid.Parse(req.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid country ID: %v", err)
		}
		country, err := c.repo.FindByID(ctx, id)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to find country: %v", err)
		}
		if country == nil {
			return nil, status.Errorf(codes.NotFound, "country not found")
		}
		return c.presenter.ToProto(country), nil
	}

	if req.CountryCode != "" {
		country, err := c.repo.FindByCode(ctx, req.CountryCode)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to find country: %v", err)
		}
		if country == nil {
			return nil, status.Errorf(codes.NotFound, "country not found")
		}
		return c.presenter.ToProto(country), nil
	}

	return nil, status.Errorf(codes.InvalidArgument, "either id or country_code must be provided")
}
