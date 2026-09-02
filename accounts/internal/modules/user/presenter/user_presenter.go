package presenter

import (
	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/models"
	"math"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserPresenter formats user domain models into protobuf responses
type UserPresenter interface {
	ToProto(user *models.User) *accountsv1.UserResponse
	ToListProto(users []models.User, total int64, page, pageSize int) *accountsv1.ListUsersResponse
}

type userPresenter struct{}

// NewUserPresenter creates a new UserPresenter instance
func NewUserPresenter() UserPresenter {
	return &userPresenter{}
}

func (p *userPresenter) ToProto(user *models.User) *accountsv1.UserResponse {
	if user == nil {
		return nil
	}

	var countryProto *accountsv1.CountryResponse
	if user.Country.ID.String() != "" && user.Country.CountryCode != "" {
		countryProto = &accountsv1.CountryResponse{
			Id:          user.Country.ID.String(),
			NameAr:      user.Country.Name.AR,
			NameEn:      user.Country.Name.EN,
			CountryCode: user.Country.CountryCode,
		}
	}

	return &accountsv1.UserResponse{
		Id:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		CountryId: user.CountryID.String(),
		Country:   countryProto,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

func (p *userPresenter) ToListProto(users []models.User, total int64, page, pageSize int) *accountsv1.ListUsersResponse {
	protoList := make([]*accountsv1.UserResponse, 0, len(users))
	for i := range users {
		protoList = append(protoList, p.ToProto(&users[i]))
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	return &accountsv1.ListUsersResponse{
		Users: protoList,
		Pagination: &accountsv1.PaginationMetadata{
			CurrentPage: int32(page),
			PageSize:    int32(pageSize),
			TotalItems:  total,
			TotalPages:  int32(totalPages),
		},
	}
}
