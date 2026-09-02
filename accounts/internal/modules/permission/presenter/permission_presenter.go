package presenter

import (
	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/models"
	"math"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// PermissionPresenter handles formatting domain models into gRPC protobuf responses
type PermissionPresenter interface {
	ToProto(permission *models.Permission) *accountsv1.PermissionResponse
	ToListProto(permissions []models.Permission, total int64, page, pageSize int) *accountsv1.ListPermissionsResponse
}

type permissionPresenter struct{}

// NewPermissionPresenter creates a new PermissionPresenter instance
func NewPermissionPresenter() PermissionPresenter {
	return &permissionPresenter{}
}

func (p *permissionPresenter) ToProto(permission *models.Permission) *accountsv1.PermissionResponse {
	if permission == nil {
		return nil
	}

	return &accountsv1.PermissionResponse{
		Id:        permission.ID.String(),
		Name:      permission.Name,
		CreatedAt: timestamppb.New(permission.CreatedAt),
		UpdatedAt: timestamppb.New(permission.UpdatedAt),
	}
}

func (p *permissionPresenter) ToListProto(permissions []models.Permission, total int64, page, pageSize int) *accountsv1.ListPermissionsResponse {
	protoList := make([]*accountsv1.PermissionResponse, 0, len(permissions))
	for i := range permissions {
		protoList = append(protoList, p.ToProto(&permissions[i]))
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	return &accountsv1.ListPermissionsResponse{
		Permissions: protoList,
		Pagination: &accountsv1.PaginationMetadata{
			CurrentPage: int32(page),
			PageSize:    int32(pageSize),
			TotalItems:  total,
			TotalPages:  int32(totalPages),
		},
	}
}
