package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/shared/events"
	"github.com/jamaah-in/v2/internal/shared/httpclient"
	"github.com/jamaah-in/v2/internal/shared/outbox"
	"github.com/jamaah-in/v2/internal/vendor_svc/model"
	"github.com/jamaah-in/v2/internal/vendor_svc/repository"
)

type VendorService struct {
	repo        *repository.VendorRepo
	packageAddr string
	httpc       *httpclient.Client
}

var (
	ErrVendorNotFound  = repository.ErrVendorNotFound
	ErrPackageNotFound = errors.New("package not found for this organization")
)

func NewVendorService(repo *repository.VendorRepo, packageAddr string) *VendorService {
	return &VendorService{repo: repo, packageAddr: packageAddr, httpc: httpclient.New()}
}

// --- Vendor Master ---

func (s *VendorService) CreateVendor(ctx context.Context, orgID uuid.UUID, req model.CreateVendorRequest) (*model.Vendor, error) {
	v := &model.Vendor{
		ID:                uuid.New(),
		OrgID:             orgID,
		Name:              req.Name,
		Type:              req.Type,
		NPWP:              strPtr(req.NPWP),
		Address:           strPtr(req.Address),
		PICName:           strPtr(req.PICName),
		PICPhone:          strPtr(req.PICPhone),
		PICEmail:          strPtr(req.PICEmail),
		BankName:          strPtr(req.BankName),
		BankAccountNumber: strPtr(req.BankAccountNumber),
		BankAccountName:   strPtr(req.BankAccountName),
		Notes:             strPtr(req.Notes),
	}
	if err := s.repo.CreateVendor(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *VendorService) GetVendor(ctx context.Context, id, orgID uuid.UUID) (*model.Vendor, error) {
	return s.repo.GetVendorByID(ctx, id, orgID)
}

func (s *VendorService) UpdateVendor(ctx context.Context, id, orgID uuid.UUID, req model.UpdateVendorRequest) (*model.Vendor, error) {
	v, err := s.repo.GetVendorByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		v.Name = *req.Name
	}
	if req.Type != nil {
		v.Type = *req.Type
	}
	if req.NPWP != nil {
		v.NPWP = req.NPWP
	}
	if req.Address != nil {
		v.Address = req.Address
	}
	if req.PICName != nil {
		v.PICName = req.PICName
	}
	if req.PICPhone != nil {
		v.PICPhone = req.PICPhone
	}
	if req.PICEmail != nil {
		v.PICEmail = req.PICEmail
	}
	if req.BankName != nil {
		v.BankName = req.BankName
	}
	if req.BankAccountNumber != nil {
		v.BankAccountNumber = req.BankAccountNumber
	}
	if req.BankAccountName != nil {
		v.BankAccountName = req.BankAccountName
	}
	if req.Notes != nil {
		v.Notes = req.Notes
	}
	if err := s.repo.UpdateVendor(ctx, v); err != nil {
		return nil, err
	}
	return s.repo.GetVendorByID(ctx, id, orgID)
}

func (s *VendorService) DeleteVendor(ctx context.Context, id, orgID uuid.UUID) error {
	return s.repo.DeleteVendor(ctx, id, orgID)
}

func (s *VendorService) ListVendors(ctx context.Context, orgID uuid.UUID, vendorType, search string, page, limit int) ([]model.Vendor, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.repo.ListVendors(ctx, orgID, vendorType, search, offset, limit)
}

// --- Vendor Bills ---

func (s *VendorService) CreateBill(ctx context.Context, orgID uuid.UUID, req model.CreateBillRequest, authToken string) (*model.VendorBill, error) {
	// The vendor must belong to the caller's org — authoritative, not
	// best-effort. vendor_bills.vendor_id has no org-scoped FK, so without this
	// a foreign-org vendor_id would be accepted and later leak that vendor's
	// name/type into this org's bill reads via the (un-org-scoped) JOIN.
	vendor, err := s.repo.GetVendorByID(ctx, req.VendorID, orgID)
	if err != nil {
		return nil, err // ErrVendorNotFound
	}

	// The package lives in package-service (separate DB) — no local FK can scope
	// it by org. Confirm it belongs to the caller's org so a bill can't be
	// attached to a foreign-org package_id and contaminate package debt rollups.
	if err := s.validatePackage(ctx, req.PackageID, authToken); err != nil {
		return nil, err
	}

	if req.Currency == "" {
		req.Currency = "IDR"
	}
	if req.ExchangeRate == 0 {
		req.ExchangeRate = 1.0
	}

	var dueDate *time.Time
	if req.DueDate != "" {
		d, err := repository.ParseDate(req.DueDate)
		if err != nil {
			return nil, err
		}
		dueDate = d
	}

	b := &model.VendorBill{
		ID:           uuid.New(),
		OrgID:        orgID,
		VendorID:     req.VendorID,
		PackageID:    req.PackageID,
		Description:  req.Description,
		Amount:       req.Amount,
		Currency:     req.Currency,
		ExchangeRate: req.ExchangeRate,
		DueDate:      dueDate,
		Status:       "belum_bayar",
	}

	// vendor.bill.created → Dr Beban / Cr Hutang Vendor at the IDR-converted
	// amount (GL is IDR).
	amountIDR := int64(float64(req.Amount) * req.ExchangeRate)
	payload, _ := json.Marshal(map[string]any{"amount": amountIDR, "vendor_name": vendor.Name})
	evt := outbox.Event{
		OrgID:         orgID,
		AggregateType: "vendor_bill",
		AggregateID:   b.ID,
		EventType:     events.EventVendorBillCreated,
		Payload:       payload,
	}
	if err := s.repo.CreateBillTx(ctx, b, evt); err != nil {
		return nil, err
	}
	return s.repo.GetBillByID(ctx, b.ID, orgID)
}

func (s *VendorService) GetBill(ctx context.Context, id, orgID uuid.UUID) (*model.VendorBill, error) {
	return s.repo.GetBillByID(ctx, id, orgID)
}

func (s *VendorService) UpdateBill(ctx context.Context, id, orgID uuid.UUID, req model.UpdateBillRequest) (*model.VendorBill, error) {
	b, err := s.repo.GetBillByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if req.Description != nil {
		b.Description = *req.Description
	}
	if req.Amount != nil {
		b.Amount = *req.Amount
	}
	if req.Currency != nil {
		b.Currency = *req.Currency
	}
	if req.ExchangeRate != nil {
		b.ExchangeRate = *req.ExchangeRate
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			b.DueDate = nil
		} else {
			d, err := repository.ParseDate(*req.DueDate)
			if err != nil {
				return nil, err
			}
			b.DueDate = d
		}
	}
	if err := s.repo.UpdateBill(ctx, b); err != nil {
		return nil, err
	}

	// Status is derived, never client-set. Editing amount/exchange_rate shifts
	// the DB-computed amount_idr, so recompute against paid_amount to keep
	// lunas/sebagian/belum_bayar consistent with reality.
	updated, err := s.repo.GetBillByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if status := deriveBillStatus(updated.PaidAmount, updated.AmountIDR); status != updated.Status {
		if err := s.repo.UpdateBillStatus(ctx, id, status); err != nil {
			return nil, err
		}
		updated.Status = status
	}
	return updated, nil
}

func (s *VendorService) DeleteBill(ctx context.Context, id, orgID uuid.UUID) error {
	return s.repo.DeleteBill(ctx, id, orgID)
}

func (s *VendorService) ListBills(ctx context.Context, orgID uuid.UUID, vendorID, packageID *uuid.UUID, status string, page, limit int) ([]model.VendorBill, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.repo.ListBills(ctx, orgID, vendorID, packageID, status, offset, limit)
}

func (s *VendorService) GetOverdueBills(ctx context.Context, orgID uuid.UUID) ([]model.VendorBill, error) {
	return s.repo.GetOverdueBills(ctx, orgID)
}

func (s *VendorService) GetBillsDueSoon(ctx context.Context, orgID uuid.UUID, withinDays int) ([]model.VendorBill, error) {
	if withinDays < 1 {
		withinDays = 7
	}
	if withinDays > 365 {
		withinDays = 365
	}
	return s.repo.GetBillsDueSoon(ctx, orgID, withinDays)
}

func (s *VendorService) GetDebtSummary(ctx context.Context, orgID uuid.UUID, vendorID *uuid.UUID) (*model.VendorDebtSummary, error) {
	return s.repo.GetDebtSummary(ctx, orgID, vendorID)
}

func (s *VendorService) GetPackageBillSummary(ctx context.Context, orgID, packageID uuid.UUID) (*model.PackageBillSummary, error) {
	return s.repo.GetPackageBillSummary(ctx, orgID, packageID)
}

// --- Vendor Payments ---

func (s *VendorService) CreatePayment(ctx context.Context, orgID uuid.UUID, req model.CreatePaymentRequest) (*model.VendorPayment, error) {
	bill, err := s.repo.GetBillByID(ctx, req.VendorBillID, orgID)
	if err != nil {
		return nil, err
	}

	if req.Currency == "" {
		req.Currency = "IDR"
	}
	if req.ExchangeRate == 0 {
		req.ExchangeRate = 1.0
	}

	paymentDate, err := repository.ParseDate(req.PaymentDate)
	if err != nil {
		return nil, err
	}
	if paymentDate == nil {
		return nil, errors.New("payment_date is required")
	}

	p := &model.VendorPayment{
		ID:               uuid.New(),
		OrgID:            orgID,
		VendorBillID:     req.VendorBillID,
		VendorID:         bill.VendorID,
		PaymentDate:      *paymentDate,
		Amount:           req.Amount,
		Currency:         req.Currency,
		ExchangeRate:     req.ExchangeRate,
		SourceAccount:    strPtr(req.SourceAccount),
		TransferProofURL: strPtr(req.TransferProofURL),
		Notes:            strPtr(req.Notes),
	}

	if err := s.repo.CreatePayment(ctx, p); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateBillPaidAmount(ctx, bill.ID); err != nil {
		return nil, err
	}

	updatedBill, err := s.repo.GetBillByID(ctx, bill.ID, orgID)
	if err != nil {
		return p, nil
	}
	_ = s.repo.UpdateBillStatus(ctx, bill.ID, deriveBillStatus(updatedBill.PaidAmount, updatedBill.AmountIDR))

	return p, nil
}

func (s *VendorService) GetPayment(ctx context.Context, id, orgID uuid.UUID) (*model.VendorPayment, error) {
	return s.repo.GetPaymentByID(ctx, id, orgID)
}

func (s *VendorService) DeletePayment(ctx context.Context, id, orgID uuid.UUID) error {
	p, err := s.repo.GetPaymentByID(ctx, id, orgID)
	if err != nil {
		return err
	}
	billID := p.VendorBillID

	if err := s.repo.DeletePayment(ctx, id, orgID); err != nil {
		return err
	}

	if err := s.repo.UpdateBillPaidAmount(ctx, billID); err != nil {
		return err
	}

	updatedBill, err := s.repo.GetBillByID(ctx, billID, orgID)
	if err != nil {
		return nil
	}
	_ = s.repo.UpdateBillStatus(ctx, billID, deriveBillStatus(updatedBill.PaidAmount, updatedBill.AmountIDR))
	return nil
}

// deriveBillStatus is the single source of truth for a bill's payment status:
// lunas once paid_amount covers the IDR total, sebagian while partially paid,
// belum_bayar otherwise. Used by create/delete payment and bill update so the
// status can never be set inconsistently with paid_amount.
func deriveBillStatus(paidAmount, amountIDR int64) string {
	switch {
	case paidAmount >= amountIDR:
		return "lunas"
	case paidAmount > 0:
		return "sebagian"
	default:
		return "belum_bayar"
	}
}

func (s *VendorService) ListPaymentsByBill(ctx context.Context, billID, orgID uuid.UUID) ([]model.VendorPayment, error) {
	return s.repo.ListPaymentsByBill(ctx, billID, orgID)
}

func (s *VendorService) ListPaymentsByVendor(ctx context.Context, vendorID, orgID uuid.UUID, page, limit int) ([]model.VendorPayment, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.repo.ListPaymentsByVendor(ctx, vendorID, orgID, offset, limit)
}

// validatePackage confirms the package belongs to the caller's org by fetching
// it from package-service with the caller's own (org-scoped) token: a foreign or
// missing package returns 404 there, surfaced here as ErrPackageNotFound, which
// blocks a cross-org bill. If package-service isn't configured (dev/test) or no
// token is present, it no-ops — mirroring jamaah-service's seat reservation.
func (s *VendorService) validatePackage(ctx context.Context, packageID uuid.UUID, authToken string) error {
	if s.packageAddr == "" || authToken == "" {
		return nil
	}
	if _, err := s.httpc.GetRaw(ctx, s.packageAddr, "/api/v1/packages/"+packageID.String(), authToken); err != nil {
		return ErrPackageNotFound
	}
	return nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
