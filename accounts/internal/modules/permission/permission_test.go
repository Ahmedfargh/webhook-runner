package permission_test

import (
	"context"
	"testing"

	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/models"
	"accounts/internal/modules/permission"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.Permission{},
		&models.Role{},
		&models.Country{},
		&models.User{},
		&models.Admin{},
	)
	require.NoError(t, err)
	return db
}

func TestPermissionModule_CRUD(t *testing.T) {
	db := setupTestDB(t)
	mod := permission.NewPermissionModule(db)
	ctx := context.Background()

	// 1. Create Permission
	created, err := mod.Controller.CreatePermission(ctx, &accountsv1.CreatePermissionRequest{
		Name: "users.read",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.Id)
	assert.Equal(t, "users.read", created.Name)

	// 2. Duplicate Check
	_, err = mod.Controller.CreatePermission(ctx, &accountsv1.CreatePermissionRequest{
		Name: "users.read",
	})
	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, st.Code())

	// 3. Get Permission
	fetched, err := mod.Controller.GetPermission(ctx, &accountsv1.GetPermissionRequest{
		Id: created.Id,
	})
	require.NoError(t, err)
	assert.Equal(t, created.Id, fetched.Id)
	assert.Equal(t, "users.read", fetched.Name)

	// 4. Update Permission
	updated, err := mod.Controller.UpdatePermission(ctx, &accountsv1.UpdatePermissionRequest{
		Id:   created.Id,
		Name: "users.write",
	})
	require.NoError(t, err)
	assert.Equal(t, "users.write", updated.Name)

	// 5. List Permissions
	list, err := mod.Controller.ListPermissions(ctx, &accountsv1.ListPermissionsRequest{
		Pagination: &accountsv1.PaginationRequest{
			Page:     1,
			PageSize: 10,
		},
	})
	require.NoError(t, err)
	assert.Len(t, list.Permissions, 1)
	assert.Equal(t, int64(1), list.Pagination.TotalItems)

	// 6. Delete Permission
	delResp, err := mod.Controller.DeletePermission(ctx, &accountsv1.DeletePermissionRequest{
		Id: created.Id,
	})
	require.NoError(t, err)
	assert.True(t, delResp.Success)

	// 7. Verify Not Found
	_, err = mod.Controller.GetPermission(ctx, &accountsv1.GetPermissionRequest{
		Id: created.Id,
	})
	assert.Error(t, err)
	st, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}
