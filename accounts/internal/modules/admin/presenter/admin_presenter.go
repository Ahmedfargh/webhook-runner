package presenter

import (
	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/models"
	permPresenter "accounts/internal/modules/permission/presenter"
	rolePresenter "accounts/internal/modules/role/presenter"
	"math"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// AdminPresenter formats admin domain models into protobuf responses
type AdminPresenter interface {
	ToProto(admin *models.Admin) *accountsv1.AdminResponse
	ToListProto(admins []models.Admin, total int64, page, pageSize int) *accountsv1.ListAdminsResponse
}

type adminPresenter struct {
	rolePres rolePresenter.RolePresenter
	permPres permPresenter.PermissionPresenter
}

// NewAdminPresenter creates a new AdminPresenter instance
func NewAdminPresenter(rolePres rolePresenter.RolePresenter, permPres permPresenter.PermissionPresenter) AdminPresenter {
	return &adminPresenter{
		rolePres: rolePres,
		permPres: permPres,
	}
}

func (p *adminPresenter) ToProto(admin *models.Admin) *accountsv1.AdminResponse {
	if admin == nil {
		return nil
	}

	var countryProto *accountsv1.CountryResponse
	if admin.Country.ID.String() != "" && admin.Country.CountryCode != "" {
		countryProto = &accountsv1.CountryResponse{
			Id:          admin.Country.ID.String(),
			NameAr:      admin.Country.Name.AR,
			NameEn:      admin.Country.Name.EN,
			CountryCode: admin.Country.CountryCode,
		}
	}

	rolesProto := make([]*accountsv1.RoleResponse, 0, len(admin.Roles))
	for i := range admin.Roles {
		rolesProto = append(rolesProto, p.rolePres.ToProto(&admin.Roles[i]))
	}

	permsProto := make([]*accountsv1.PermissionResponse, 0, len(admin.Permissions))
	for i := range admin.Permissions {
		permsProto = append(permsProto, p.permPres.ToProto(&admin.Permissions[i]))
	}

	return &accountsv1.AdminResponse{
		Id:          admin.ID.String(),
		Name:        admin.Name,
		Email:       admin.Email,
		Phone:       admin.Phone,
		CountryId:   admin.CountryID.String(),
		Country:     countryProto,
		Roles:       rolesProto,
		Permissions: permsProto,
		CreatedAt:   timestamppb.New(admin.CreatedAt),
		UpdatedAt:   timestamppb.New(admin.UpdatedAt),
	}
}

func (p *adminPresenter) ToListProto(admins []models.Admin, total int64, page, pageSize int) *accountsv1.ListAdminsResponse {
	protoList := make([]*accountsv1.AdminResponse, 0, len(admins))
	for i := range admins {
		protoList = append(protoList, p.ToProto(&admins[i]))
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	return &accountsv1.ListAdminsResponse{
		Admins: protoList,
		Pagination: &accountsv1.PaginationMetadata{
			CurrentPage: int32(page),
			PageSize:    int32(pageSize),
			TotalItems:  total,
			TotalPages:  int32(totalPages),
		},
	}
}
