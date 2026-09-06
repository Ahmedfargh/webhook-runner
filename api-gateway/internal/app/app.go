package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"webhookApiGateway/internal/config"
	"webhookApiGateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

type App struct {
	cfg       *config.Config
	container *Container
	router    *gin.Engine
}

func New(cfg *config.Config) (*App, error) {
	container, err := NewContainer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize container: %w", err)
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))
	router.Use(middleware.AuditMiddleware(container.AuditEmitter))
	router.Use(middleware.RequestTrackerMiddleware(container.RequestTrackerEmitter))

	app := &App{
		cfg:       cfg,
		container: container,
		router:    router,
	}

	app.setupRoutes()

	return app, nil
}

func (a *App) setupRoutes() {
	// Health & Reference Routes
	a.router.GET("/health", a.container.HealthHandler.HealthCheck)
	a.router.GET("/api/v1/health", a.container.HealthHandler.HealthCheck)
	a.router.POST("/webhooks/test-receiver", a.container.WebhookHandler.TestReceiver)
	a.router.GET("/api/v1/countries", a.container.CountryHandler.ListCountries)

	// API v1 Group
	v1 := a.router.Group("/api/v1")
	{
		v1.POST("/webhooks/test-receiver", a.container.WebhookHandler.TestReceiver)

		// Public Auth
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", a.container.AuthHandler.Login)
			authGroup.POST("/register", a.container.AuthHandler.Register)
		}

		// Public Plans
		v1.GET("/plans", a.container.PlanHandler.ListPlans)
		v1.GET("/plans/:id", a.container.PlanHandler.GetPlan)

		// Protected Routes
		protected := v1.Group("")
		protected.Use(middleware.JWTAuthMiddleware(a.cfg.JWTSecret))
		{
			// Auth Profile
			protected.GET("/auth/me", a.container.AuthHandler.Me)

			// Users
			users := protected.Group("/users")
			{
				users.GET("", a.container.UserHandler.ListUsers)
				users.POST("", a.container.UserHandler.CreateUser)
				users.GET("/:id", a.container.UserHandler.GetUser)
				users.PUT("/:id", a.container.UserHandler.UpdateUser)
				users.DELETE("/:id", a.container.UserHandler.DeleteUser)
			}

			// Admins
			admins := protected.Group("/admins")
			{
				admins.GET("", a.container.AdminHandler.ListAdmins)
				admins.POST("", a.container.AdminHandler.CreateAdmin)
				admins.GET("/:id", a.container.AdminHandler.GetAdmin)
				admins.PUT("/:id", a.container.AdminHandler.UpdateAdmin)
				admins.DELETE("/:id", a.container.AdminHandler.DeleteAdmin)
				admins.POST("/:id/roles", a.container.AdminHandler.AssignRoles)
				admins.POST("/:id/permissions", a.container.AdminHandler.AssignPermissions)
			}

			// Roles
			roles := protected.Group("/roles")
			{
				roles.GET("", a.container.RoleHandler.ListRoles)
				roles.POST("", a.container.RoleHandler.CreateRole)
				roles.GET("/:id", a.container.RoleHandler.GetRole)
				roles.PUT("/:id", a.container.RoleHandler.UpdateRole)
				roles.DELETE("/:id", a.container.RoleHandler.DeleteRole)
				roles.POST("/:id/permissions", a.container.RoleHandler.AssignPermissions)
			}

			// Permissions
			perms := protected.Group("/permissions")
			{
				perms.GET("", a.container.PermHandler.ListPermissions)
				perms.POST("", a.container.PermHandler.CreatePermission)
				perms.GET("/:id", a.container.PermHandler.GetPermission)
				perms.PUT("/:id", a.container.PermHandler.UpdatePermission)
				perms.DELETE("/:id", a.container.PermHandler.DeletePermission)
			}

			// Admin Plans
			adminPlans := protected.Group("/admin/plans")
			{
				adminPlans.POST("", a.container.PlanHandler.CreatePlan)
				adminPlans.PUT("/:id", a.container.PlanHandler.UpdatePlan)
				adminPlans.DELETE("/:id", a.container.PlanHandler.DeletePlan)
			}

			// Subscriptions
			subscriptions := protected.Group("/subscriptions")
			{
				subscriptions.GET("/current", a.container.SubscriptionHandler.GetCurrentSubscription)
				subscriptions.POST("/subscribe", a.container.SubscriptionHandler.Subscribe)
				subscriptions.POST("/cancel", a.container.SubscriptionHandler.CancelSubscription)
				subscriptions.GET("/admin/all", a.container.SubscriptionHandler.ListSubscriptions)
				subscriptions.POST("/admin/override", a.container.SubscriptionHandler.AdminOverrideSubscription)
			}

			// Invoices
			invoices := protected.Group("/invoices")
			{
				invoices.GET("", a.container.InvoiceHandler.GetMyInvoices)
				invoices.GET("/:id", a.container.InvoiceHandler.GetInvoice)
				invoices.GET("/admin/all", a.container.InvoiceHandler.ListAllInvoices)
				invoices.POST("/admin/create", a.container.InvoiceHandler.CreateManualInvoice)
				invoices.POST("/admin/:id/mark-paid", a.container.InvoiceHandler.MarkInvoicePaid)
				invoices.POST("/admin/:id/void", a.container.InvoiceHandler.VoidInvoice)
			}

			// Manual Payments
			manualPayments := protected.Group("/manual-payments")
			{
				manualPayments.POST("", a.container.ManualPaymentHandler.SubmitManualPayment)
				manualPayments.GET("/admin/all", a.container.ManualPaymentHandler.ListManualPayments)
				manualPayments.POST("/admin/:id/review", a.container.ManualPaymentHandler.ReviewManualPayment)
			}

			// Applications
			apps := protected.Group("/apps")
			{
				apps.GET("", a.container.AppHandler.ListApps)
				apps.POST("", a.container.AppHandler.CreateApp)
				apps.GET("/:id", a.container.AppHandler.GetApp)
				apps.PUT("/:id", a.container.AppHandler.UpdateApp)
				apps.DELETE("/:id", a.container.AppHandler.DeleteApp)
				apps.POST("/:id/rotate-secrets", a.container.AppHandler.RotateSecrets)
			}

			// Webhooks
			webhooks := protected.Group("/webhooks")
			{
				webhooks.POST("/dispatch", a.container.WebhookHandler.SendWebhook)
				webhooks.POST("/send", a.container.WebhookHandler.SendWebhook)
				webhooks.GET("/dispatch", a.container.WebhookHandler.SendWebhook)
				webhooks.GET("/send", a.container.WebhookHandler.SendWebhook)
				webhooks.GET("/calls", a.container.WebhookHandler.ListWebhookCalls)
				webhooks.GET("/calls/:id", a.container.WebhookHandler.GetWebhookCall)
				webhooks.POST("/calls/:id/retry", a.container.WebhookHandler.RetryWebhookCall)
			}

			// Audit Logs & Action Trail
			auditLogs := protected.Group("/audit-logs")
			{
				auditLogs.GET("", a.container.AuditHandler.ListAuditLogs)
				auditLogs.GET("/:id", a.container.AuditHandler.GetAuditLog)
			}

			// Request Traces & Telemetry
			traces := protected.Group("/request-traces")
			{
				traces.GET("", a.container.RequestTraceHandler.ListTraces)
				traces.GET("/stats", a.container.RequestTraceHandler.GetStats)
				traces.GET("/:id", a.container.RequestTraceHandler.GetTrace)
			}
		}
	}
}

func (a *App) Run(ctx context.Context) error {
	serverAddr := fmt.Sprintf(":%s", a.cfg.Port)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      a.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("API Gateway HTTP server running on http://localhost%s\n", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Context cancelled. Initiating graceful shutdown...")
	case sig := <-stopChan:
		log.Printf("Received signal %v. Initiating graceful shutdown...\n", sig)
	case err := <-serverErr:
		return fmt.Errorf("HTTP server error: %w", err)
	}

	log.Println("Shutting down API Gateway...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Warning: Server forced to shutdown: %v\n", err)
	}

	a.container.Close()
	log.Println("API Gateway exited cleanly.")
	return nil
}
