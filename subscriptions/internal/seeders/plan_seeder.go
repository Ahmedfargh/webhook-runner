package seeders

import (
	"encoding/json"
	"log"

	"subscriptions/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedPlans(db *gorm.DB) {
	var count int64
	db.Model(&models.Plan{}).Count(&count)
	if count > 0 {
		return
	}

	freeFeatures, _ := json.Marshal([]string{
		"3 Webhook Endpoints",
		"5,000 Events / month",
		"Basic Retry Logic (3 retries)",
		"Standard Email Support",
		"7-day Log Retention",
	})

	starterFeatures, _ := json.Marshal([]string{
		"15 Webhook Endpoints",
		"50,000 Events / month",
		"Exponential Backoff Retries",
		"Custom Payload Headers & Secrets",
		"30-day Log Retention",
		"3 Team Members",
	})

	proFeatures, _ := json.Marshal([]string{
		"50 Webhook Endpoints",
		"250,000 Events / month",
		"Advanced Filtering & Transformations",
		"Custom SSL / TLS Verification",
		"90-day Log Retention",
		"10 Team Members",
		"Priority Support & Live Chat",
	})

	enterpriseFeatures, _ := json.Marshal([]string{
		"Unlimited Webhook Endpoints",
		"2,000,000 Events / month",
		"Dedicated Ingress IP Addresses",
		"Custom Rate Limiting & Circuit Breaking",
		"365-day Log Retention",
		"Unlimited Team Members",
		"99.99% Uptime SLA & Dedicated Account Manager",
	})

	plans := []models.Plan{
		{
			ID:                uuid.New(),
			Name:              "Free Developer",
			Code:              "free",
			Description:       "Ideal for testing, small pet projects, and local developer environments.",
			PriceMonthly:      0.00,
			PriceAnnually:     0.00,
			Currency:          "USD",
			MaxWebhooks:       3,
			MaxEventsPerMonth: 5000,
			MaxTeamMembers:    1,
			FeaturesJSON:      string(freeFeatures),
			IsActive:          true,
			IsPopular:         false,
			TierLevel:         1,
		},
		{
			ID:                uuid.New(),
			Name:              "Starter Tier",
			Code:              "starter",
			Description:       "Perfect for fast-growing startups and essential production webhook flows.",
			PriceMonthly:      19.00,
			PriceAnnually:     190.00,
			Currency:          "USD",
			MaxWebhooks:       15,
			MaxEventsPerMonth: 50000,
			MaxTeamMembers:    3,
			FeaturesJSON:      string(starterFeatures),
			IsActive:          true,
			IsPopular:         false,
			TierLevel:         2,
		},
		{
			ID:                uuid.New(),
			Name:              "Professional Business",
			Code:              "pro",
			Description:       "Recommended for modern SaaS applications requiring high throughput and fine-grained security.",
			PriceMonthly:      49.00,
			PriceAnnually:     490.00,
			Currency:          "USD",
			MaxWebhooks:       50,
			MaxEventsPerMonth: 250000,
			MaxTeamMembers:    10,
			FeaturesJSON:      string(proFeatures),
			IsActive:          true,
			IsPopular:         true,
			TierLevel:         3,
		},
		{
			ID:                uuid.New(),
			Name:              "Enterprise Scale",
			Code:              "enterprise",
			Description:       "Mission-critical reliability, multi-region failover, custom SLA, and dedicated compliance.",
			PriceMonthly:      149.00,
			PriceAnnually:     1490.00,
			Currency:          "USD",
			MaxWebhooks:       -1, // Unlimited
			MaxEventsPerMonth: 2000000,
			MaxTeamMembers:    -1, // Unlimited
			FeaturesJSON:      string(enterpriseFeatures),
			IsActive:          true,
			IsPopular:         false,
			TierLevel:         4,
		},
	}

	for _, p := range plans {
		if err := db.Create(&p).Error; err != nil {
			log.Println("Error seeding plan:", p.Name, err)
		}
	}

	log.Println("Seeded default subscription plans successfully")
}
