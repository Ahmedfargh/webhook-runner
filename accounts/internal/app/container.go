package app

import (
	"accounts/internal/audit"
	"accounts/internal/config"
	"accounts/internal/modules/admin"
	"accounts/internal/modules/country"
	"accounts/internal/modules/permission"
	"accounts/internal/modules/role"
	"accounts/internal/modules/user"
	"accounts/internal/repository"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Container struct {
	DB               *gorm.DB
	AuditEmitter     *audit.KafkaEmitter
	CountryModule    *country.CountryModule
	PermissionModule *permission.PermissionModule
	RoleModule       *role.RoleModule
	UserModule       *user.UserModule
	AdminModule      *admin.AdminModule
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	countryRepo := repository.NewCountryRepository(db)

	countryMod := country.NewCountryModule(db)
	permissionMod := permission.NewPermissionModule(db)
	roleMod := role.NewRoleModule(db, permissionMod.Repository, permissionMod.Presenter)
	userMod := user.NewUserModule(db, countryRepo)
	adminMod := admin.NewAdminModule(
		db,
		countryRepo,
		roleMod.Repository,
		permissionMod.Repository,
		roleMod.Presenter,
		permissionMod.Presenter,
	)

	auditEmitter := audit.NewEmitter(cfg.KafkaBrokers, cfg.KafkaTopicAudit, "accounts", cfg.KafkaEnabled)

	return &Container{
		DB:               db,
		AuditEmitter:     auditEmitter,
		CountryModule:    countryMod,
		PermissionModule: permissionMod,
		RoleModule:       roleMod,
		UserModule:       userMod,
		AdminModule:      adminMod,
	}
}

func (c *Container) Close() {
	if c.AuditEmitter != nil {
		_ = c.AuditEmitter.Close()
	}
}

func (c *Container) RegisterGRPCServices(server *grpc.Server) {
	c.CountryModule.RegisterGRPC(server)
	c.PermissionModule.RegisterGRPC(server)
	c.RoleModule.RegisterGRPC(server)
	c.UserModule.RegisterGRPC(server)
	c.AdminModule.RegisterGRPC(server)
}
