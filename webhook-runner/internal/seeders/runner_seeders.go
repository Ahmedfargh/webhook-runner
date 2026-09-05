package seeders

import (
	"fmt"
	"log"
	"time"

	"webhookRunner/internal/helpers"
	"webhookRunner/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedRunnerData(db *gorm.DB) {
	var count int64
	db.Model(&models.App{}).Count(&count)
	if count > 0 {
		return
	}

	log.Println("Seeding initial Webhook Runner applications and sample execution calls...")

	// Demo User ID (or default)
	demoUserID := uuid.MustParse("a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d")

	sampleApps := []struct {
		Name          string
		WebhookURL    string
		AppID         string
		AppSecret     string
		WebhookSecret string
	}{
		{
			Name:          "Stripe Payment Ingestion",
			WebhookURL:    "https://api.acme-enterprise.com/webhooks/stripe",
			AppID:         helpers.GenerateAppID(),
			AppSecret:     helpers.GenerateAppSecret(),
			WebhookSecret: helpers.GenerateWebhookSecret(),
		},
		{
			Name:          "Shopify Store Order Sync",
			WebhookURL:    "https://api.acme-enterprise.com/webhooks/shopify",
			AppID:         helpers.GenerateAppID(),
			AppSecret:     helpers.GenerateAppSecret(),
			WebhookSecret: helpers.GenerateWebhookSecret(),
		},
		{
			Name:          "GitHub CI/CD Deployment Hook",
			WebhookURL:    "https://ci.acme-enterprise.com/deployments/notify",
			AppID:         helpers.GenerateAppID(),
			AppSecret:     helpers.GenerateAppSecret(),
			WebhookSecret: helpers.GenerateWebhookSecret(),
		},
	}

	for _, sa := range sampleApps {
		app := models.App{
			ID:            uuid.New(),
			UserID:        demoUserID,
			Name:          sa.Name,
			AppID:         sa.AppID,
			AppSecret:     sa.AppSecret,
			WebhookURL:    sa.WebhookURL,
			WebhookSecret: sa.WebhookSecret,
			IsActive:      true,
			CreatedAt:     time.Now().Add(-48 * time.Hour),
			UpdatedAt:     time.Now(),
		}

		if err := db.Create(&app).Error; err != nil {
			log.Printf("Failed to seed app %s: %v\n", sa.Name, err)
			continue
		}

		// Seed some delivery calls
		calls := []models.WebhookCall{
			{
				ID:                 uuid.New(),
				AppID:              app.ID,
				EventName:          "order.created",
				TargetURL:          app.WebhookURL,
				PayloadJSON:        `{"event":"order.created","order_id":"ord_100982","amount":249.99,"currency":"USD"}`,
				HeadersJSON:        `{"X-Custom-Env":"production"}`,
				AttemptCount:       1,
				Status:             models.StatusSuccess,
				ResponseStatusCode: 200,
				ResponseBody:       `{"status":"received","processed_at":"` + time.Now().Format(time.RFC3339) + `"}`,
				ResponseLatencyMS:  48,
				CreatedAt:          time.Now().Add(-2 * time.Hour),
				UpdatedAt:          time.Now().Add(-2 * time.Hour),
			},
			{
				ID:                 uuid.New(),
				AppID:              app.ID,
				EventName:          "payment.captured",
				TargetURL:          app.WebhookURL,
				PayloadJSON:        `{"event":"payment.captured","charge_id":"ch_3M49aBc","amount":249.99}`,
				HeadersJSON:        `{"X-Idempotency-Key":"idem_882190"}`,
				AttemptCount:       1,
				Status:             models.StatusSuccess,
				ResponseStatusCode: 201,
				ResponseBody:       `{"ok":true,"code":201}`,
				ResponseLatencyMS:  62,
				CreatedAt:          time.Now().Add(-1 * time.Hour),
				UpdatedAt:          time.Now().Add(-1 * time.Hour),
			},
			{
				ID:                 uuid.New(),
				AppID:              app.ID,
				EventName:          "invoice.payment_failed",
				TargetURL:          app.WebhookURL,
				PayloadJSON:        `{"event":"invoice.payment_failed","invoice_id":"inv_44012","reason":"insufficient_funds"}`,
				AttemptCount:       2,
				Status:             models.StatusFailed,
				ResponseStatusCode: 500,
				ResponseBody:       `{"error":"Internal processing error on consumer endpoint"}`,
				ResponseLatencyMS:  320,
				ErrorMessage:       "Destination server returned HTTP 500",
				CreatedAt:          time.Now().Add(-30 * time.Minute),
				UpdatedAt:          time.Now().Add(-30 * time.Minute),
			},
		}

		for _, call := range calls {
			_ = db.Create(&call)
		}
	}

	fmt.Println("Webhook Runner applications & execution logs seeded successfully!")
}
