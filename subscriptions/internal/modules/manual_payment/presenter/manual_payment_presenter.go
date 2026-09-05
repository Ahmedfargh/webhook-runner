package presenter

import (
	"time"

	subscriptionsv1 "subscriptions/api/proto/v1"
	"subscriptions/internal/models"
)

type ManualPaymentPresenter interface {
	ToProto(payment *models.ManualPaymentRecord) *subscriptionsv1.ManualPaymentResponse
	ToListProto(payments []models.ManualPaymentRecord, total int64, page, pageSize int) *subscriptionsv1.ListManualPaymentsResponse
}

type manualPaymentPresenter struct{}

func NewManualPaymentPresenter() ManualPaymentPresenter {
	return &manualPaymentPresenter{}
}

func (p *manualPaymentPresenter) ToProto(pmt *models.ManualPaymentRecord) *subscriptionsv1.ManualPaymentResponse {
	if pmt == nil {
		return nil
	}

	adminIDStr := ""
	if pmt.RecordedByAdminID != nil {
		adminIDStr = pmt.RecordedByAdminID.String()
	}

	return &subscriptionsv1.ManualPaymentResponse{
		Id:                   pmt.ID.String(),
		InvoiceId:            pmt.InvoiceID.String(),
		UserId:               pmt.UserID.String(),
		Amount:               pmt.Amount,
		Currency:             pmt.Currency,
		PaymentMethod:        pmt.PaymentMethod,
		TransactionReference: pmt.TransactionReference,
		PayerName:            pmt.PayerName,
		PayerNotes:           pmt.PayerNotes,
		RecordedByAdminId:    adminIDStr,
		AdminNotes:           pmt.AdminNotes,
		Status:               string(pmt.Status),
		CreatedAt:            pmt.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            pmt.UpdatedAt.Format(time.RFC3339),
	}
}

func (p *manualPaymentPresenter) ToListProto(payments []models.ManualPaymentRecord, total int64, page, pageSize int) *subscriptionsv1.ListManualPaymentsResponse {
	res := make([]*subscriptionsv1.ManualPaymentResponse, 0, len(payments))
	for i := range payments {
		res = append(res, p.ToProto(&payments[i]))
	}

	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))

	return &subscriptionsv1.ListManualPaymentsResponse{
		Payments:    res,
		TotalItems:  total,
		CurrentPage: int32(page),
		TotalPages:  totalPages,
	}
}
