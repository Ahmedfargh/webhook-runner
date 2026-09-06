package app

import (
	subscriptionsv1 "subscriptions/api/proto/v1"
	"subscriptions/internal/audit"
	"subscriptions/internal/config"
	"subscriptions/internal/modules/invoice"
	"subscriptions/internal/modules/manual_payment"
	"subscriptions/internal/modules/plan"
	"subscriptions/internal/modules/subscription"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Container struct {
	DB                  *gorm.DB
	AuditEmitter        *audit.KafkaEmitter
	PlanModule          *plan.PlanModule
	InvoiceModule       *invoice.InvoiceModule
	SubscriptionModule  *subscription.SubscriptionModule
	ManualPaymentModule *manual_payment.ManualPaymentModule
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	planMod := plan.NewPlanModule(db)
	invoiceMod := invoice.NewInvoiceModule(db)
	subMod := subscription.NewSubscriptionModule(db, planMod.Repo, planMod.Presenter, invoiceMod.Service)
	manualPaymentMod := manual_payment.NewManualPaymentModule(db, invoiceMod.Service, subMod.Repo)

	auditEmitter := audit.NewEmitter(cfg.KafkaBrokers, cfg.KafkaTopicAudit, "subscriptions", cfg.KafkaEnabled)

	return &Container{
		DB:                  db,
		AuditEmitter:        auditEmitter,
		PlanModule:          planMod,
		InvoiceModule:       invoiceMod,
		SubscriptionModule:  subMod,
		ManualPaymentModule: manualPaymentMod,
	}
}

func (c *Container) Close() {
	if c.AuditEmitter != nil {
		_ = c.AuditEmitter.Close()
	}
}

func (c *Container) RegisterGRPCServices(server *grpc.Server) {
	subscriptionsv1.RegisterPlanServiceServer(server, c.PlanModule.Controller)
	subscriptionsv1.RegisterInvoiceServiceServer(server, c.InvoiceModule.Controller)
	subscriptionsv1.RegisterSubscriptionServiceServer(server, c.SubscriptionModule.Controller)
	subscriptionsv1.RegisterManualPaymentServiceServer(server, c.ManualPaymentModule.Controller)
}
