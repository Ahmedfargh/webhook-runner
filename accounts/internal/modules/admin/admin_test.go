package admin_test

import (
	"context"
	"testing"

	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/helpers/passwords"
	"accounts/internal/models"
	"accounts/internal/modules/admin"
	"accounts/internal/modules/permission"
	"accounts/internal/modules/role"
	"accounts/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, models.Country) {
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

	country := models.Country{
		ID:          uuid.New(),
		CountryCode: "US",
		Name: models.LocaleName{
			AR: "الولايات المتحدة",
			EN: "United States",
		},
	}
	err = db.Create(&country).Error
	require.NoError(t, err)

	return db, country
}

func TestAdminModule_CRUD_And_Relations(t *testing.T) {
	db, country := setupTestDB(t)
	countryRepo := repository.NewCountryRepository(db)
	permMod := permission.NewPermissionModule(db)
	roleMod := role.NewRoleModule(db, permMod.Repository, permMod.Presenter)
	adminMod := admin.NewAdminModule(
		db,
		countryRepo,
		roleMod.Repository,
		permMod.Repository,
		roleMod.Presenter,
		permMod.Presenter,
	)
	ctx := context.Background()

	// 1. Setup role and permission
	perm, err := permMod.Controller.CreatePermission(ctx, &accountsv1.CreatePermissionRequest{Name: "system.admin"})
	require.NoError(t, err)

	r, err := roleMod.Controller.CreateRole(ctx, &accountsv1.CreateRoleRequest{
		Name:          "SuperAdminRole",
		PermissionIds: []string{perm.Id},
	})
	require.NoError(t, err)

	// 2. Create Admin
	created, err := adminMod.Controller.CreateAdmin(ctx, &accountsv1.CreateAdminRequest{
		Name:          "Root Admin",
		Email:         "admin@domain.com",
		Phone:         "+12025550123",
		Password:      "Secret123456",
		CountryId:     country.ID.String(),
		RoleIds:       []string{r.Id},
		PermissionIds: []string{perm.Id},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.Id)
	assert.Equal(t, "Root Admin", created.Name)
	assert.Len(t, created.Roles, 1)
	assert.Len(t, created.Permissions, 1)

	// Check password hash
	var dbAdmin models.Admin
	err = db.First(&dbAdmin, "id = ?", created.Id).Error
	require.NoError(t, err)
	assert.True(t, passwords.CheckPasswordHash("Secret123456", dbAdmin.Password))

	// 3. Get Admin
	fetched, err := adminMod.Controller.GetAdmin(ctx, &accountsv1.GetAdminRequest{Id: created.Id})
	require.NoError(t, err)
	assert.Equal(t, created.Id, fetched.Id)
	assert.Equal(t, "Root Admin", fetched.Name)

	// 4. Update Admin
	updated, err := adminMod.Controller.UpdateAdmin(ctx, &accountsv1.UpdateAdminRequest{
		Id:   created.Id,
		Name: "Root Administrator",
	})
	require.NoError(t, err)
	assert.Equal(t, "Root Administrator", updated.Name)

	// 5. Delete Admin
	delResp, err := adminMod.Controller.DeleteAdmin(ctx, &accountsv1.DeleteAdminRequest{Id: created.Id})
	require.NoError(t, err)
	assert.True(t, delResp.Success)

	// 6. Verify Not Found
	_, err = adminMod.Controller.GetAdmin(ctx, &accountsv1.GetAdminRequest{Id: created.Id})
	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}
