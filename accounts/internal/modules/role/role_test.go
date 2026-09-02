package role_test

import (
	"context"
	"testing"

	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/models"
	"accounts/internal/modules/permission"
	"accounts/internal/modules/role"

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

func TestRoleModule_CRUD_And_Permissions(t *testing.T) {
	db := setupTestDB(t)
	permMod := permission.NewPermissionModule(db)
	roleMod := role.NewRoleModule(db, permMod.Repository, permMod.Presenter)
	ctx := context.Background()

	// 1. Create permissions
	p1, err := permMod.Controller.CreatePermission(ctx, &accountsv1.CreatePermissionRequest{Name: "roles.create"})
	require.NoError(t, err)
	p2, err := permMod.Controller.CreatePermission(ctx, &accountsv1.CreatePermissionRequest{Name: "roles.delete"})
	require.NoError(t, err)

	// 2. Create Role with permission
	created, err := roleMod.Controller.CreateRole(ctx, &accountsv1.CreateRoleRequest{
		Name:          "RoleManager",
		PermissionIds: []string{p1.Id},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.Id)
	assert.Equal(t, "RoleManager", created.Name)
	assert.Len(t, created.Permissions, 1)

	// 3. Assign additional permission to Role
	updated, err := roleMod.Controller.AssignPermissionsToRole(ctx, &accountsv1.AssignPermissionsToRoleRequest{
		RoleId:        created.Id,
		PermissionIds: []string{p1.Id, p2.Id},
	})
	require.NoError(t, err)
	assert.Len(t, updated.Permissions, 2)

	// 4. Get Role
	fetched, err := roleMod.Controller.GetRole(ctx, &accountsv1.GetRoleRequest{Id: created.Id})
	require.NoError(t, err)
	assert.Equal(t, created.Id, fetched.Id)
	assert.Len(t, fetched.Permissions, 2)

	// 5. Delete Role
	delResp, err := roleMod.Controller.DeleteRole(ctx, &accountsv1.DeleteRoleRequest{Id: created.Id})
	require.NoError(t, err)
	assert.True(t, delResp.Success)

	// 6. Verify Not Found
	_, err = roleMod.Controller.GetRole(ctx, &accountsv1.GetRoleRequest{Id: created.Id})
	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}
