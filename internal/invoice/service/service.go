package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/invoice/model"
	"github.com/jamaah-in/v2/internal/invoice/repository"
	"github.com/jamaah-in/v2/internal/shared/config"
	"github.com/jamaah-in/v2/internal/shared/events"
	"github.com/jamaah-in/v2/internal/shared/httpclient"
	"github.com/jamaah-in/v2/internal/shared/outbox"
)

type InvoiceService struct {
	repo        *repository.InvoiceRepo
	pakasir     config.PakasirConfig
	internalKey string
	authAddr    string
	publicURL   string
	httpc       *httpclient.Client
}

// PaymentDeps carries the settings the payment/subscription flow needs.
type PaymentDeps struct {
	Pakasir     config.PakasirConfig
	InternalKey string
	AuthAddr    string
	PublicURL   string
}

func NewInvoiceService(repo *repository.InvoiceRepo) *InvoiceService {
	return &InvoiceService{repo: repo, httpc: httpclient.New()}
}

// WithPayments configures the Pakasir + cross-service activation dependencies.
func (s *InvoiceService) WithPayments(d PaymentDeps) *InvoiceService {
	s.pakasir = d.Pakasir
	s.internalKey = d.InternalKey
	s.authAddr = d.AuthAddr
	s.publicURL = d.PublicURL
	return s
}

func (s *InvoiceService) CreateInvoice(ctx context.Context, orgID, userID uuid.UUID, req model.CreateInvoiceRequest) (*model.Invoice, error) {
	totalAmount := req.PriceSnapshot - req.DiscountAmount + req.SurchargeAmount
	if totalAmount < 0 {
		totalAmount = 0
	}

	inv := &model.Invoice{
		ID:              uuid.New(),
		OrgID:           orgID,
		InvoiceNumber:   repository.GenerateInvoiceNumber(orgID),
		JamaahID:        req.JamaahID,
		PackageID:       req.PackageID,
		RegistrationID:  req.RegistrationID,
		RoomType:        req.RoomType,
		PriceSnapshot:   req.PriceSnapshot,
		DiscountAmount:  req.DiscountAmount,
		SurchargeAmount: req.SurchargeAmount,
		TotalAmount:     totalAmount,
		AmountPaid:      0,
		AmountRemaining: totalAmount,
		PaymentScheme:   req.PaymentScheme,
		Status:          string(model.InvoiceStatusBelumBayar),
		Notes:           req.Notes,
	}

	if req.DueDate != "" {
		t, err := parseDate(req.DueDate)
		if err != nil {
			return nil, fmt.Errorf("due_date: %w", err)
		}
		inv.DueDate = t
	}

	// Emit invoice.issued so the accounting engine recognizes
	// Piutang Jemaah (D) / Pendapatan Paket (K). The outbox row is written in the
	// same tx as the invoice insert — no lost events.
	payload, _ := json.Marshal(map[string]any{
		"total_amount":   inv.TotalAmount,
		"invoice_number": inv.InvoiceNumber,
	})
	evt := outbox.Event{
		OrgID:         orgID,
		AggregateType: "invoice",
		AggregateID:   inv.ID,
		EventType:     events.EventInvoiceIssued,
		Payload:       payload,
		OccurredAt:    time.Now(),
	}
	if err := s.repo.CreateInvoiceTx(ctx, inv, evt); err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *InvoiceService) GetInvoice(ctx context.Context, id, orgID uuid.UUID) (*model.Invoice, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	schedules, _ := s.repo.GetPaymentSchedules(ctx, id)
	payments, _ := s.repo.GetPayments(ctx, id)
	inv.PaymentSchedules = schedules
	inv.Payments = payments
	return inv, nil
}

func (s *InvoiceService) GetInvoiceByNumber(ctx context.Context, orgID uuid.UUID, number string) (*model.Invoice, error) {
	inv, err := s.repo.GetInvoiceByNumber(ctx, orgID, number)
	if err != nil {
		return nil, err
	}
	schedules, _ := s.repo.GetPaymentSchedules(ctx, inv.ID)
	payments, _ := s.repo.GetPayments(ctx, inv.ID)
	inv.PaymentSchedules = schedules
	inv.Payments = payments
	return inv, nil
}

func (s *InvoiceService) ListInvoices(ctx context.Context, orgID uuid.UUID, status string, page, limit int) ([]model.Invoice, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.repo.ListInvoices(ctx, orgID, status, offset, limit)
}

func (s *InvoiceService) GetInvoicesByJamaah(ctx context.Context, orgID, jamaahID uuid.UUID) ([]model.Invoice, error) {
	return s.repo.GetInvoicesByJamaah(ctx, orgID, jamaahID)
}

func (s *InvoiceService) UpdateInvoice(ctx context.Context, id, orgID uuid.UUID, req model.UpdateInvoiceRequest) (*model.Invoice, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if inv.Status == string(model.InvoiceStatusBatal) {
		return nil, repository.ErrAlreadyCancelled
	}

	if req.DiscountAmount != nil {
		inv.DiscountAmount = *req.DiscountAmount
	}
	if req.SurchargeAmount != nil {
		inv.SurchargeAmount = *req.SurchargeAmount
	}
	if req.Notes != nil {
		inv.Notes = *req.Notes
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			inv.DueDate = nil
		} else {
			t, err := parseDate(*req.DueDate)
			if err != nil {
				return nil, fmt.Errorf("due_date: %w", err)
			}
			inv.DueDate = t
		}
	}

	inv.TotalAmount = inv.PriceSnapshot - inv.DiscountAmount + inv.SurchargeAmount
	inv.AmountRemaining = inv.TotalAmount - inv.AmountPaid
	if inv.AmountRemaining < 0 {
		inv.AmountRemaining = 0
	}

	if err := s.repo.UpdateInvoice(ctx, inv); err != nil {
		return nil, err
	}
	return s.repo.GetInvoiceByID(ctx, id, orgID)
}

func (s *InvoiceService) CancelInvoice(ctx context.Context, id, orgID uuid.UUID, reason string) error {
	return s.repo.CancelInvoice(ctx, id, orgID, reason)
}

func (s *InvoiceService) CreatePaymentSchedules(ctx context.Context, orgID, invoiceID uuid.UUID, req model.CreatePaymentScheduleRequest) ([]model.PaymentSchedule, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, invoiceID, orgID)
	if err != nil {
		return nil, err
	}
	if inv.Status == string(model.InvoiceStatusBatal) {
		return nil, repository.ErrAlreadyCancelled
	}

	totalScheduled := int64(0)
	schedules := make([]model.PaymentSchedule, 0, len(req.Installments))
	for _, inst := range req.Installments {
		totalScheduled += inst.Amount
	}

	if totalScheduled > inv.TotalAmount {
		return nil, fmt.Errorf("total scheduled amount (%d) exceeds invoice total (%d)", totalScheduled, inv.TotalAmount)
	}

	for i, inst := range req.Installments {
		ps := &model.PaymentSchedule{
			ID:             uuid.New(),
			InvoiceID:      invoiceID,
			InstallmentNum: i + 1,
			Amount:         inst.Amount,
			Description:    strPtr(inst.Description),
			IsPaid:         false,
		}
		if inst.DueDate != "" {
			t, err := parseDate(inst.DueDate)
			if err == nil {
				ps.DueDate = t
			}
		}
		if err := s.repo.CreatePaymentSchedule(ctx, ps); err != nil {
			return nil, err
		}
		schedules = append(schedules, *ps)
	}

	return schedules, nil
}

func (s *InvoiceService) GetPaymentSchedules(ctx context.Context, orgID, invoiceID uuid.UUID) ([]model.PaymentSchedule, error) {
	_, err := s.repo.GetInvoiceByID(ctx, invoiceID, orgID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetPaymentSchedules(ctx, invoiceID)
}

func (s *InvoiceService) RecordPayment(ctx context.Context, orgID, userID, invoiceID uuid.UUID, req model.RecordPaymentRequest) (*model.Payment, *model.Invoice, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, invoiceID, orgID)
	if err != nil {
		return nil, nil, err
	}
	if inv.Status == string(model.InvoiceStatusBatal) {
		return nil, nil, repository.ErrAlreadyCancelled
	}
	if inv.Status == string(model.InvoiceStatusLunas) {
		return nil, nil, repository.ErrAlreadyLunas
	}

	paidAt := time.Now()
	if req.PaidAt != "" {
		t, err := parseDate(req.PaidAt)
		if err == nil && t != nil {
			paidAt = *t
		}
	}

	payment := &model.Payment{
		ID:              uuid.New(),
		OrgID:           orgID,
		InvoiceID:       invoiceID,
		Amount:          req.Amount,
		PaymentMethod:   req.PaymentMethod,
		BankName:        strPtr(req.BankName),
		AccountNumber:   strPtr(req.AccountNumber),
		ReferenceNumber: strPtr(req.ReferenceNumber),
		ProofURL:        strPtr(req.ProofURL),
		Notes:           req.Notes,
		ReceivedBy:      userID,
		PaidAt:          paidAt,
	}
	// Link cash payments to the cashier's open POS session (for tutup-kas recon).
	if req.PaymentMethod == "tunai" || req.PaymentMethod == "cash" {
		payment.CashSessionID = s.repo.ActiveSessionID(ctx, orgID, userID)
	}

	inv.AmountPaid += req.Amount
	inv.AmountRemaining = inv.TotalAmount - inv.AmountPaid
	if inv.AmountRemaining <= 0 {
		inv.AmountRemaining = 0
		inv.Status = string(model.InvoiceStatusLunas)
	} else if inv.AmountPaid > 0 {
		inv.Status = string(model.InvoiceStatusSebagian)
	}

	// Persist payment + invoice update + payment.received outbox event in one tx.
	payload, _ := json.Marshal(map[string]any{
		"amount":         req.Amount,
		"payment_method": req.PaymentMethod,
		"invoice_number": inv.InvoiceNumber,
		"jamaah_id":      inv.JamaahID,
	})
	evt := outbox.Event{
		OrgID:         orgID,
		AggregateType: "invoice",
		AggregateID:   invoiceID,
		EventType:     events.EventPaymentReceived,
		Payload:       payload,
		OccurredAt:    paidAt,
	}
	if err := s.repo.RecordPaymentTx(ctx, payment, inv, evt); err != nil {
		return nil, nil, fmt.Errorf("record payment: %w", err)
	}

	// Allocate the cumulative paid amount across the installment schedule in
	// order (DP first, then cicilan): mark each schedule paid once the running
	// total of its amounts is covered. This enforces "DP before cicilan" — a
	// later installment is only settled after the earlier ones are fully paid.
	s.allocatePaymentsToSchedules(ctx, invoiceID, inv.AmountPaid)

	updated, err := s.repo.GetInvoiceByID(ctx, invoiceID, orgID)
	if err != nil {
		return payment, nil, nil
	}
	return payment, updated, nil
}

// allocatePaymentsToSchedules marks installment schedules paid in installment
// order as the cumulative amountPaid covers them. Best-effort: schedule sync
// must never fail the payment itself.
func (s *InvoiceService) allocatePaymentsToSchedules(ctx context.Context, invoiceID uuid.UUID, amountPaid int64) {
	schedules, err := s.repo.GetPaymentSchedules(ctx, invoiceID)
	if err != nil || len(schedules) == 0 {
		return
	}
	sort.Slice(schedules, func(i, j int) bool {
		return schedules[i].InstallmentNum < schedules[j].InstallmentNum
	})
	var cumulative int64
	for _, sch := range schedules {
		cumulative += sch.Amount
		if !sch.IsPaid && amountPaid >= cumulative {
			_ = s.repo.MarkSchedulePaid(ctx, sch.ID)
		}
	}
}

// SettleFromCredit applies a non-cash credit (savings conversion) to an invoice.
// The GL journal is posted by the originating module, so no event is emitted here.
func (s *InvoiceService) SettleFromCredit(ctx context.Context, invoiceID, orgID uuid.UUID, amount int64) (int64, error) {
	return s.repo.SettleFromCredit(ctx, invoiceID, orgID, amount)
}

func (s *InvoiceService) GetPayments(ctx context.Context, orgID, invoiceID uuid.UUID) ([]model.Payment, error) {
	_, err := s.repo.GetInvoiceByID(ctx, invoiceID, orgID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetPayments(ctx, invoiceID)
}

func (s *InvoiceService) GetSummary(ctx context.Context, orgID uuid.UUID) (*model.InvoiceSummary, error) {
	return s.repo.GetSummary(ctx, orgID)
}

func (s *InvoiceService) GetPackageRevenue(ctx context.Context, orgID, packageID uuid.UUID) (*model.PackageRevenueSummary, error) {
	return s.repo.GetPackageRevenue(ctx, orgID, packageID)
}

func (s *InvoiceService) GetPackageRevenueAll(ctx context.Context, orgID uuid.UUID) ([]model.PackageRevenueSummary, error) {
	return s.repo.GetPackageRevenueAll(ctx, orgID)
}

func (s *InvoiceService) GetMonthlyRevenue(ctx context.Context, orgID uuid.UUID, months int) ([]model.MonthlyRevenuePoint, error) {
	return s.repo.GetMonthlyRevenue(ctx, orgID, months)
}

func (s *InvoiceService) GetBalances(ctx context.Context, orgID uuid.UUID) ([]model.JamaahBalance, error) {
	return s.repo.GetBalancesByJamaah(ctx, orgID)
}

func (s *InvoiceService) ListInvoicesByPackage(ctx context.Context, orgID, packageID uuid.UUID) ([]model.Invoice, error) {
	return s.repo.ListInvoicesByPackage(ctx, orgID, packageID)
}

// ---- POS cash sessions ----

func (s *InvoiceService) OpenCashSession(ctx context.Context, orgID, userID uuid.UUID, openingFloat int64, notes string) (*model.CashSession, error) {
	return s.repo.OpenSession(ctx, orgID, userID, openingFloat, notes)
}

func (s *InvoiceService) GetActiveCashSession(ctx context.Context, orgID, userID uuid.UUID) (*model.CashSession, error) {
	return s.repo.GetActiveSession(ctx, orgID, userID)
}

func (s *InvoiceService) ListCashSessions(ctx context.Context, orgID uuid.UUID, limit int) ([]model.CashSession, error) {
	if limit < 1 || limit > 100 {
		limit = 30
	}
	return s.repo.ListSessions(ctx, orgID, limit)
}

func (s *InvoiceService) CloseCashSession(ctx context.Context, orgID, sessionID uuid.UUID, counted int64) (*model.CashSession, error) {
	return s.repo.CloseSessionTx(ctx, orgID, sessionID, counted, events.EventPosCashSessionClosed)
}

func parseDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	formats := []string{"2006-01-02", "2006-01-02T15:04:05Z", "2006-01-02T15:04:05"}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("invalid date format: %s", s)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
