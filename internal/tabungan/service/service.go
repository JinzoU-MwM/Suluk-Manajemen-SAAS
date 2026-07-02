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
	// Defense-in-depth: the handler already rejects amount < 1, but guard here too
	// — a non-positive amount would decrement the balance via the deposit path,
	// bypassing the withdrawal/overdraw checks.
	if req.Amount <= 0 {
		return nil, repository.ErrInvalidAmount
	}
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

// Convert applies savings balance to settle an invoice. Order matters: it
// reserves (locks + decrements) the tabungan balance FIRST, so a concurrent
// second Convert() attempt sees the reduced balance instead of racing on a
// stale read (T1/T7) — only then does it call invoice-service, with an
// idempotency key tied to the reservation so a lost-response retry (the
// shared httpclient retries automatically on 5xx) replays safely instead of
// double-applying (T2). If the invoice-service call fails outright, the
// reservation is fully compensated (credited back); if it applies less than
// reserved (invoice had less remaining than requested), the shortfall is
// credited back. The applied amount is min(requested|balance, invoice remaining).
func (s *Service) Convert(ctx context.Context, orgID, userID, accountID uuid.UUID, req model.ConvertRequest) (*model.SavingsAccount, error) {
	acct, err := s.repo.GetAccount(ctx, orgID, accountID)
	if err != nil {
		return nil, err
	}
	if acct.Status != model.StatusAktif {
		return nil, repository.ErrNotActive
	}
	invoiceID, err := uuid.Parse(req.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice_id tidak valid")
	}
	want := req.Amount
	if want <= 0 || want > acct.Balance {
		want = acct.Balance
	}
	if want <= 0 {
		return nil, repository.ErrInsufficient
	}

	depositID, err := s.repo.ReserveForConvert(ctx, orgID, accountID, want, &userID)
	if err != nil {
		return nil, err
	}
	idempotencyKey := depositID.String()

	applied, settleErr := s.settleInvoice(ctx, orgID, acct.JamaahID, invoiceID, want, idempotencyKey)
	if settleErr == nil && applied <= 0 {
		settleErr = fmt.Errorf("invoice sudah lunas / tidak ada sisa untuk dilunasi")
	}
	if settleErr != nil {
		if _, compErr := s.repo.CompensateConvert(ctx, orgID, accountID, depositID, want); compErr != nil {
			s.log.Errorw("CONVERSION INCONSISTENCY: reserved but neither settled nor compensated",
				"org_id", orgID, "account_id", accountID, "deposit_id", depositID, "invoice_id", req.InvoiceID, "err", compErr)
		}
		return nil, settleErr
	}

	payload, _ := json.Marshal(map[string]any{"amount": applied})
	evt := outbox.Event{
		OrgID:         orgID,
		AggregateType: "savings_account",
		AggregateID:   accountID,
		EventType:     events.EventSavingsConverted,
		Payload:       payload,
	}
	acct, err = s.repo.ConfirmConvert(ctx, orgID, accountID, depositID, want, applied, req.InvoiceID, evt)
	if err != nil {
		s.log.Errorw("CONVERSION INCONSISTENCY: invoice settled but savings confirm failed",
			"org_id", orgID, "account_id", accountID, "deposit_id", depositID, "invoice_id", req.InvoiceID, "applied", applied, "err", err)
		return nil, err
	}
	return acct, nil
}

type settleReq struct {
	InvoiceID      string `json:"invoice_id"`
	OrgID          string `json:"org_id"`
	JamaahID       string `json:"jamaah_id"`
	Amount         int64  `json:"amount"`
	IdempotencyKey string `json:"idempotency_key"`
}
type settleResp struct {
	Applied int64 `json:"applied"`
}

func (s *Service) settleInvoice(ctx context.Context, orgID, jamaahID, invoiceID uuid.UUID, amount int64, idempotencyKey string) (int64, error) {
	if s.invoiceAddr == "" || s.internalKey == "" {
		return 0, fmt.Errorf("invoice service / internal key not configured")
	}
	var out settleResp
	err := s.httpc.PostJSON(ctx, s.invoiceAddr, "/api/v1/invoices/internal/settle",
		map[string]string{"X-Internal-Key": s.internalKey},
		settleReq{InvoiceID: invoiceID.String(), OrgID: orgID.String(), JamaahID: jamaahID.String(), Amount: amount, IdempotencyKey: idempotencyKey}, &out)
	if err != nil {
		return 0, err
	}
	return out.Applied, nil
}
