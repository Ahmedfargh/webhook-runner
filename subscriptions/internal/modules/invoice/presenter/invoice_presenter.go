package presenter

import (
	"time"

	subscriptionsv1 "subscriptions/api/proto/v1"
	"subscriptions/internal/models"
)

type InvoicePresenter interface {
	ToProto(invoice *models.Invoice) *subscriptionsv1.InvoiceResponse
	ToListProto(invoices []models.Invoice, total int64, page, pageSize int) *subscriptionsv1.ListInvoicesResponse
}

type invoicePresenter struct{}

func NewInvoicePresenter() InvoicePresenter {
	return &invoicePresenter{}
}

func (p *invoicePresenter) ToProto(inv *models.Invoice) *subscriptionsv1.InvoiceResponse {
	if inv == nil {
		return nil
	}

	items := make([]*subscriptionsv1.InvoiceItemResponse, 0, len(inv.Items))
	for _, itm := range inv.Items {
		items = append(items, &subscriptionsv1.InvoiceItemResponse{
			Id:          itm.ID.String(),
			Description: itm.Description,
			Quantity:    itm.Quantity,
			UnitPrice:   itm.UnitPrice,
			Total:       itm.Total,
		})
	}

	paidAtStr := ""
	if inv.PaidAt != nil {
		paidAtStr = inv.PaidAt.Format(time.RFC3339)
	}

	return &subscriptionsv1.InvoiceResponse{
		Id:                      inv.ID.String(),
		InvoiceNumber:           inv.InvoiceNumber,
		SubscriptionId:          inv.SubscriptionID.String(),
		UserId:                  inv.UserID.String(),
		Amount:                  inv.Amount,
		Tax:                     inv.Tax,
		TotalAmount:             inv.TotalAmount,
		Currency:                inv.Currency,
		Status:                  string(inv.Status),
		DueDate:                 inv.DueDate.Format(time.RFC3339),
		PaidAt:                  paidAtStr,
		PaymentMethod:           inv.PaymentMethod,
		PaymentReference:        inv.PaymentReference,
		BankAccountInstructions: inv.BankAccountInstructions,
		Notes:                   inv.Notes,
		Items:                   items,
		CreatedAt:               inv.CreatedAt.Format(time.RFC3339),
		UpdatedAt:               inv.UpdatedAt.Format(time.RFC3339),
	}
}

func (p *invoicePresenter) ToListProto(invoices []models.Invoice, total int64, page, pageSize int) *subscriptionsv1.ListInvoicesResponse {
	res := make([]*subscriptionsv1.InvoiceResponse, 0, len(invoices))
	for i := range invoices {
		res = append(res, p.ToProto(&invoices[i]))
	}

	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))

	return &subscriptionsv1.ListInvoicesResponse{
		Invoices:    res,
		TotalItems:  total,
		CurrentPage: int32(page),
		TotalPages:  totalPages,
	}
}
