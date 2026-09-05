package presenter

import (
	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/models"
)

type CountryPresenter interface {
	ToProto(country *models.Country) *accountsv1.CountryResponse
	ToProtoList(countries []models.Country) []*accountsv1.CountryResponse
}

type countryPresenter struct{}

func NewCountryPresenter() CountryPresenter {
	return &countryPresenter{}
}

func (p *countryPresenter) ToProto(country *models.Country) *accountsv1.CountryResponse {
	if country == nil {
		return nil
	}

	return &accountsv1.CountryResponse{
		Id:          country.ID.String(),
		NameAr:      country.Name.AR,
		NameEn:      country.Name.EN,
		CountryCode: country.CountryCode,
	}
}

func (p *countryPresenter) ToProtoList(countries []models.Country) []*accountsv1.CountryResponse {
	result := make([]*accountsv1.CountryResponse, len(countries))
	for i := range countries {
		result[i] = p.ToProto(&countries[i])
	}
	return result
}
