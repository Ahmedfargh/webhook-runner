package user_test

import (
	"context"
	"testing"

	accountsv1 "accounts/api/proto/v1"
	"accounts/internal/helpers/passwords"
	"accounts/internal/models"
	"accounts/internal/modules/user"
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
		CountryCode: "EG",
		Name: models.LocaleName{
			AR: "مصر",
			EN: "Egypt",
		},
	}
	err = db.Create(&country).Error
	require.NoError(t, err)

	return db, country
}

func TestUserModule_CRUD(t *testing.T) {
	db, country := setupTestDB(t)
	countryRepo := repository.NewCountryRepository(db)
	userMod := user.NewUserModule(db, countryRepo)
	ctx := context.Background()

	// 1. Create User
	created, err := userMod.Controller.CreateUser(ctx, &accountsv1.CreateUserRequest{
		Name:      "Ahmed Ali",
		Email:     "ahmed@example.com",
		Phone:     "01012345678",
		Password:  "password123",
		CountryId: country.ID.String(),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.Id)
	assert.Equal(t, "Ahmed Ali", created.Name)
	assert.Equal(t, "ahmed@example.com", created.Email)
	assert.Equal(t, "+201012345678", created.Phone) // Normalized phone
	assert.NotNil(t, created.Country)
	assert.Equal(t, "Egypt", created.Country.NameEn)

	// Verify password was hashed in DB
	var dbUser models.User
	err = db.First(&dbUser, "id = ?", created.Id).Error
	require.NoError(t, err)
	assert.True(t, passwords.CheckPasswordHash("password123", dbUser.Password))

	// 2. Duplicate Email Check
	_, err = userMod.Controller.CreateUser(ctx, &accountsv1.CreateUserRequest{
		Name:      "Ahmed Clone",
		Email:     "ahmed@example.com",
		Phone:     "01012345679",
		Password:  "password123",
		CountryId: country.ID.String(),
	})
	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, st.Code())

	// 3. Get User
	fetched, err := userMod.Controller.GetUser(ctx, &accountsv1.GetUserRequest{Id: created.Id})
	require.NoError(t, err)
	assert.Equal(t, created.Id, fetched.Id)
	assert.Equal(t, "ahmed@example.com", fetched.Email)

	// 4. Update User
	pwd := "newsecret456"
	updated, err := userMod.Controller.UpdateUser(ctx, &accountsv1.UpdateUserRequest{
		Id:       created.Id,
		Name:     "Ahmed Updated",
		Password: &pwd,
	})
	require.NoError(t, err)
	assert.Equal(t, "Ahmed Updated", updated.Name)

	err = db.First(&dbUser, "id = ?", created.Id).Error
	require.NoError(t, err)
	assert.True(t, passwords.CheckPasswordHash("newsecret456", dbUser.Password))

	// 5. List Users
	list, err := userMod.Controller.ListUsers(ctx, &accountsv1.ListUsersRequest{
		Pagination: &accountsv1.PaginationRequest{
			Page:     1,
			PageSize: 10,
			Search:   "Ahmed",
		},
	})
	require.NoError(t, err)
	assert.Len(t, list.Users, 1)

	// 6. Delete User
	delResp, err := userMod.Controller.DeleteUser(ctx, &accountsv1.DeleteUserRequest{Id: created.Id})
	require.NoError(t, err)
	assert.True(t, delResp.Success)

	// 7. Verify Not Found
	_, err = userMod.Controller.GetUser(ctx, &accountsv1.GetUserRequest{Id: created.Id})
	assert.Error(t, err)
	st, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}
