package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/jamaah-in/v2/internal/shared/events"
	"github.com/jamaah-in/v2/internal/shared/httpclient"
	"github.com/jamaah-in/v2/internal/shared/outbox"
	"github.com/jamaah-in/v2/internal/tabungan/model"
	"github.com/jamaah-in/v2/internal/tabungan/repository"
)

type Service struct {
	repo        *repository.Repo
	log         *zap.SugaredLogger
	httpc       *httpclient.Client
	invoiceAddr string
	internalKey string
}

func NewService(repo *repository.Repo, log *zap.SugaredLogger, invoiceAddr, internalKey string) *Service {
	return &Service{repo: repo, log: log, httpc: httpclient.New(), invoiceAddr: invoiceAddr, internalKey: internalKey}
}

func (s *Service) CreateAccount(ctx context.Context, orgID uuid.UUID, req model.CreateAccountRequest) (*model.SavingsAccount, error) {
	jamaahID, err := uuid.Parse(req.JamaahID)
	if err != nil {
		return nil, fmt.Errorf("jamaah_id tidak valid")
	}
	a := &model.SavingsAccount{OrgID: orgID, JamaahID: jamaahID, JamaahName: req.JamaahName, TargetAmount: req.TargetAmount, Notes: req.Notes}
	if req.TargetPackageID != "" {
		if pid, perr := uuid.Parse(req.TargetPackageID); perr == nil {
			a.TargetPackageID = &pid
		}
	}
	if err := s.repo.CreateAccount(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) ListAccounts(ctx context.Context, orgID uuid.UUID, page, limit int) ([]model.SavingsAccount, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.ListAccounts(ctx, orgID, (page-1)*limit, limit)
}

func (s *Service) GetAccount(ctx context.Context, orgID, id uuid.UUID) (*model.SavingsAccount, error) {
	return s.repo.GetAccount(ctx, orgID, id)
}

// Deposit records a savings deposit and emits savings.deposited
// (Dr Kas/Bank, Cr Hutang Tabungan).
func (s *Service) Deposit(ctx context.Context, orgID, userID, accountID uuid.UUID, req model.DepositRequest) (*model.SavingsAccount, error) {
	method := req.Method
	if method == "" {
		method = "tunai"
	}
	d := &model.Deposit{Amount: req.Amount, Method: method, Reference: req.Reference, Notes: req.Notes, CreatedBy: &userID}
	payload, _ := json.Marshal(map[string]any{"amount": req.Amount, "payment_method": method})
	evt := outbox.Event{
		OrgID:         orgID,
		AggregateType: "savings_account",
		AggregateID:   accountID,
		EventType:     events.EventSavingsDeposited,
		Payload:       payload,
	}
	return s.repo.DepositTx(ctx, orgID, accountID, d, evt)
}

// Convert applies savings balance to settle an invoice: it first settles the
// invoice via the invoice-service internal endpoint (no cash event), then
// records the conversion + emits savings.converted (Dr Hutang Tabungan, Cr
// Piutang). The applied amount is min(requested|balance, invoice remaining).
func (s *Service) Convert(ctx context.Context, orgID, userID, accountID uuid.UUID, req model.ConvertRequest) (*model.SavingsAccount, error) {
	acct, err := s.repo.GetAccount(ctx, orgID, accountID)
	if err != nil {
		return nil, err
	}
	if acct.Status != model.StatusAktif {
		return nil, repository.ErrNotActive
	}
	want := req.Amount
	if want <= 0 || want > acct.Balance {
		want = acct.Balance
	}
	if want <= 0 {
		return nil, repository.ErrInsufficient
	}

	// 1) Settle the invoice (idempotent-ish, returns the actually applied amount).
	applied, err := s.settleInvoice(ctx, orgID, req.InvoiceID, want)
	if err != nil {
		return nil, fmt.Errorf("settle invoice: %w", err)
	}
	if applied <= 0 {
		return nil, fmt.Errorf("invoice sudah lunas / tidak ada sisa untuk dilunasi")
	}

	// 2) Reduce savings + emit savings.converted for the applied amount.
	payload, _ := json.Marshal(map[string]any{"amount": applied})
	evt := outbox.Event{
		OrgID:         orgID,
		AggregateType: "savings_account",
		AggregateID:   accountID,
		EventType:     events.EventSavingsConverted,
		Payload:       payload,
	}
	acct, err = s.repo.ConvertTx(ctx, orgID, accountID, applied, &userID, req.InvoiceID, evt)
	if err != nil {
		// Invoice already settled but tabungan tx failed — surface for manual fix.
		s.log.Errorw("CONVERSION INCONSISTENCY: invoice settled but savings not reduced",
			"org_id", orgID, "account_id", accountID, "invoice_id", req.InvoiceID, "applied", applied, "err", err)
		return nil, err
	}
	return acct, nil
}

type settleReq struct {
	InvoiceID string `json:"invoice_id"`
	OrgID     string `json:"org_id"`
	Amount    int64  `json:"amount"`
}
type settleResp struct {
	Applied int64 `json:"applied"`
}

func (s *Service) settleInvoice(ctx context.Context, orgID uuid.UUID, invoiceID string, amount int64) (int64, error) {
	if s.invoiceAddr == "" || s.internalKey == "" {
		return 0, fmt.Errorf("invoice service / internal key not configured")
	}
	var out settleResp
	err := s.httpc.PostJSON(ctx, s.invoiceAddr, "/api/v1/invoices/internal/settle",
		map[string]string{"X-Internal-Key": s.internalKey},
		settleReq{InvoiceID: invoiceID, OrgID: orgID.String(), Amount: amount}, &out)
	if err != nil {
		return 0, err
	}
	return out.Applied, nil
}
