package presenter

import (
	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/models"
	permissionPresenter "accounts/internal/modules/permission/presenter"
	"math"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// RolePresenter formats role domain models to gRPC responses
type RolePresenter interface {
	ToProto(role *models.Role) *accountsv1.RoleResponse
	ToListProto(roles []models.Role, total int64, page, pageSize int) *accountsv1.ListRolesResponse
}

type rolePresenter struct {
	permPresenter permissionPresenter.PermissionPresenter
}

// NewRolePresenter creates a new RolePresenter instance
func NewRolePresenter(permPresenter permissionPresenter.PermissionPresenter) RolePresenter {
	return &rolePresenter{
		permPresenter: permPresenter,
	}
}

func (p *rolePresenter) ToProto(role *models.Role) *accountsv1.RoleResponse {
	if role == nil {
		return nil
	}

	protoPerms := make([]*accountsv1.PermissionResponse, 0, len(role.Permissions))
	for i := range role.Permissions {
		protoPerms = append(protoPerms, p.permPresenter.ToProto(&role.Permissions[i]))
	}

	return &accountsv1.RoleResponse{
		Id:          role.ID.String(),
		Name:        role.Name,
		Permissions: protoPerms,
		CreatedAt:   timestamppb.New(role.CreatedAt),
		UpdatedAt:   timestamppb.New(role.UpdatedAt),
	}
}

func (p *rolePresenter) ToListProto(roles []models.Role, total int64, page, pageSize int) *accountsv1.ListRolesResponse {
	protoList := make([]*accountsv1.RoleResponse, 0, len(roles))
	for i := range roles {
		protoList = append(protoList, p.ToProto(&roles[i]))
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	return &accountsv1.ListRolesResponse{
		Roles: protoList,
		Pagination: &accountsv1.PaginationMetadata{
			CurrentPage: int32(page),
			PageSize:    int32(pageSize),
			TotalItems:  total,
			TotalPages:  int32(totalPages),
		},
	}
}
