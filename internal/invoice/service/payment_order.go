package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jamaah-in/v2/internal/invoice/model"
)

func (s *InvoiceService) CreatePaymentOrder(ctx context.Context, orgID, userID uuid.UUID, req model.CreatePaymentOrderRequest) (*model.PaymentOrderResponse, error) {
	planType := req.PlanType
	if planType == "" {
		planType = "monthly"
	}
	if planType != "monthly" && planType != "yearly" {
		return nil, fmt.Errorf("invalid plan type: must be monthly or yearly")
	}

	var amount int64
	switch planType {
	case "monthly":
		amount = 299000
	case "yearly":
		amount = 2990000
	}

	orderID := uuid.New()
	redirectURL := fmt.Sprintf("/payment/%s", orderID.String())

	order := &model.PaymentOrder{
		ID:          orderID,
		OrgID:       orgID,
		UserID:      userID,
		PlanType:    planType,
		Amount:      amount,
		Status:      "pending",
		RedirectURL: &redirectURL,
	}

	if err := s.repo.CreatePaymentOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("create payment order: %w", err)
	}

	return &model.PaymentOrderResponse{
		OrderID:     orderID.String(),
		RedirectURL: redirectURL,
		Status:      "pending",
		Amount:      amount,
	}, nil
}

func (s *InvoiceService) CheckPaymentStatus(ctx context.Context, orderIDStr string, orgID uuid.UUID) (*model.PaymentStatusResponse, error) {
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid order id")
	}

	order, err := s.repo.GetPaymentOrder(ctx, orderID, orgID)
	if err != nil {
		return nil, err
	}

	return &model.PaymentStatusResponse{
		OrderID: order.ID.String(),
		Status:  order.Status,
		Amount:  order.Amount,
	}, nil
}
