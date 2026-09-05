package modules_test

import (
	"context"
	"testing"
	"time"

	"subscriptions/internal/models"
	"subscriptions/internal/modules/invoice"
	"subscriptions/internal/modules/manual_payment"
	"subscriptions/internal/modules/plan"
	"subscriptions/internal/modules/subscription"
	subService "subscriptions/internal/modules/subscription/service"
	mpService "subscriptions/internal/modules/manual_payment/service"
	"subscriptions/internal/seeders"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.Plan{},
		&models.Subscription{},
		&models.Invoice{},
		&models.InvoiceItem{},
		&models.ManualPaymentRecord{},
	)
	require.NoError(t, err)

	seeders.SeedPlans(db)
	return db
}

func TestSubscriptionAndManualPaymentFlow(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	planMod := plan.NewPlanModule(db)
	invoiceMod := invoice.NewInvoiceModule(db)
	subMod := subscription.NewSubscriptionModule(db, planMod.Repo, planMod.Presenter, invoiceMod.Service)
	mpMod := manual_payment.NewManualPaymentModule(db, invoiceMod.Service, subMod.Repo)

	// 1. Verify Seeded Plans
	plans, err := planMod.Service.ListPlans(ctx, false)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(plans), 4)

	var proPlan *models.Plan
	for i := range plans {
		if plans[i].Code == "pro" {
			proPlan = &plans[i]
			break
		}
	}
	require.NotNil(t, proPlan)
	assert.Equal(t, 49.00, proPlan.PriceMonthly)

	// 2. User Subscribes to Pro Plan (Monthly)
	testUserID := uuid.New()
	subResult, err := subMod.Service.Subscribe(ctx, subService.SubscribeInput{
		UserID:        testUserID,
		PlanID:        proPlan.ID,
		BillingCycle:  models.BillingCycleMonthly,
		PaymentMethod: "bank_transfer",
		Notes:         "Testing manual subscription purchase",
	})
	require.NoError(t, err)
	require.NotNil(t, subResult)
	assert.Equal(t, models.StatusPendingManualPayment, subResult.Subscription.Status)
	assert.NotEmpty(t, subResult.InvoiceID)
	assert.Equal(t, 49.00, subResult.AmountDue)
	assert.Contains(t, subResult.PaymentInstructions, "Offline Bank Transfer Instructions")

	// 3. User Submits Offline Payment Proof
	invUUID, err := uuid.Parse(subResult.InvoiceID)
	require.NoError(t, err)

	paymentRecord, err := mpMod.Service.SubmitPayment(ctx, mpService.SubmitPaymentInput{
		InvoiceID:            invUUID,
		UserID:               testUserID,
		Amount:               49.00,
		Currency:             "USD",
		PaymentMethod:        "bank_wire",
		TransactionReference: "WIRE-TX-998822",
		PayerName:            "Acme Corp",
		PayerNotes:           "Sent from Standard Bank account",
	})
	require.NoError(t, err)
	require.NotNil(t, paymentRecord)
	assert.Equal(t, models.PaymentStatusSubmitted, paymentRecord.Status)

	// 4. Admin Reviews and Approves Offline Payment
	adminID := uuid.New()
	reviewedPmt, err := mpMod.Service.ReviewPayment(ctx, mpService.ReviewPaymentInput{
		PaymentID:  paymentRecord.ID,
		Approve:    true,
		AdminNotes: "Bank transfer funds verified in corporate account",
		AdminID:    &adminID,
	})
	require.NoError(t, err)
	assert.Equal(t, models.PaymentStatusApproved, reviewedPmt.Status)

	// 5. Verify Invoice is now Paid
	inv, err := invoiceMod.Service.GetInvoice(ctx, invUUID)
	require.NoError(t, err)
	assert.Equal(t, models.InvoiceStatusPaid, inv.Status)
	assert.NotNil(t, inv.PaidAt)

	// 6. Verify Subscription is now Active with future period end
	activeSub, err := subMod.Service.GetUserSubscription(ctx, testUserID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusActive, activeSub.Status)
	assert.True(t, activeSub.CurrentPeriodEnd.After(time.Now()))
}
