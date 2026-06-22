package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/jamaah-in/v2/internal/accounting/model"
	"github.com/jamaah-in/v2/internal/accounting/repository"
	"github.com/jamaah-in/v2/internal/shared/ai"
	"github.com/jamaah-in/v2/internal/shared/events"
)

type Service struct {
	repo   *repository.Repo
	log    *zap.SugaredLogger
	ai     ai.Generator
	narr   *narrativeCache // memoized AI narratives, keyed by (org, prompt fingerprint)
	seeded sync.Map        // orgID -> struct{}: in-process cache so we don't re-check COA every event
}

func NewService(repo *repository.Repo, log *zap.SugaredLogger) *Service {
	return &Service{repo: repo, log: log, narr: newNarrativeCache(6*time.Hour, 512)}
}

// WithAI attaches an optional text Generator for copilot narratives. A nil
// generator (no key for the selected provider) is fine — insights still return
// the rule-based findings.
func (s *Service) WithAI(c ai.Generator) *Service {
	s.ai = c
	return s
}

// EnsureCOA upserts the standard chart of accounts for an org on first use.
// It always runs SeedAccounts (idempotent: ON CONFLICT DO NOTHING) rather than
// only when the org has zero accounts — this BACKFILLS accounts added to
// StandardCOA after an org was first seeded. Without it, a posting referencing a
// newly-added account would fail with ErrUnknownAccount and the journal would be
// dropped, silently understating the org's GL. Cached per process so it runs at
// most once per org per process.
func (s *Service) EnsureCOA(ctx context.Context, orgID uuid.UUID) error {
	if _, ok := s.seeded.Load(orgID); ok {
		return nil
	}
	inserted, err := s.repo.SeedAccounts(ctx, orgID, StandardCOA())
	if err != nil {
		return err
	}
	if inserted > 0 && s.log != nil {
		s.log.Infow("seeded standard COA", "org_id", orgID, "accounts_added", inserted)
	}
	s.seeded.Store(orgID, struct{}{})
	return nil
}

// PostFromEvent maps an event to a balanced journal and persists it
// idempotently. Returns (posted, error). posted=false means either a duplicate
// (already processed) or no template for the event type — both are non-errors
// the consumer can ACK.
func (s *Service) PostFromEvent(ctx context.Context, env *events.Envelope) (bool, error) {
	orgID, err := uuid.Parse(env.OrgID)
	if err != nil {
		return false, fmt.Errorf("invalid org_id in event: %w", err)
	}

	p, err := buildPosting(env)
	if err != nil {
		if err == ErrNoTemplate {
			return false, nil // skip unmapped event types
		}
		return false, err
	}

	if err := s.EnsureCOA(ctx, orgID); err != nil {
		return false, fmt.Errorf("ensure COA: %w", err)
	}

	var srcRef *uuid.UUID
	if aid, perr := uuid.Parse(env.AggregateID); perr == nil {
		srcRef = &aid
	}

	in := repository.PostInput{
		OrgID:         orgID,
		JournalNo:     journalNo(p.module, env.OccurredAt),
		Date:          env.OccurredAt,
		SourceModule:  p.module,
		SourceRefID:   srcRef,
		SourceEventID: env.EventID,
		EventType:     env.EventType,
		Description:   p.description,
		Lines:         p.lines,
	}
	_, posted, err := s.repo.Post(ctx, in)
	return posted, err
}

func journalNo(module string, t time.Time) string {
	if t.IsZero() {
		t = time.Now()
	}
	// Random 48-bit suffix guarantees uniqueness within (org, journal_no) without
	// a sequence table. A timestamp-derived suffix collided when two distinct
	// same-module events shared a 100µs window, failing the UNIQUE(org_id,
	// journal_no) constraint and dropping the journal. Idempotency for the SAME
	// event is handled separately via source_event_id, so a non-deterministic
	// journal_no is safe here. 12 hex chars keeps it within journal_no's VARCHAR(40).
	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")[:12]
	return fmt.Sprintf("JRN-%s-%s-%s", module, t.Format("20060102"), suffix)
}

// ---- COA passthrough ----

func (s *Service) ListAccounts(ctx context.Context, orgID uuid.UUID) ([]model.Account, error) {
	if err := s.EnsureCOA(ctx, orgID); err != nil {
		return nil, err
	}
	return s.repo.ListAccounts(ctx, orgID)
}

func (s *Service) CreateAccount(ctx context.Context, a *model.Account) error {
	return s.repo.CreateAccount(ctx, a)
}

// ---- Journals ----

func (s *Service) ListJournals(ctx context.Context, orgID uuid.UUID, page, limit int) ([]model.Journal, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.ListJournals(ctx, orgID, (page-1)*limit, limit)
}

func (s *Service) GetJournal(ctx context.Context, orgID, id uuid.UUID) (*model.Journal, error) {
	return s.repo.GetJournal(ctx, orgID, id)
}

func (s *Service) GeneralLedger(ctx context.Context, orgID, accountID uuid.UUID, from, to time.Time) ([]model.JournalLine, error) {
	return s.repo.GeneralLedger(ctx, orgID, accountID, from, to)
}

func (s *Service) TrialBalance(ctx context.Context, orgID uuid.UUID, asOf time.Time) ([]model.TrialBalanceRow, error) {
	if err := s.EnsureCOA(ctx, orgID); err != nil {
		return nil, err
	}
	return s.repo.TrialBalance(ctx, orgID, asOf)
}

// BalanceSheet (Neraca) as of a date. Net income for the period up to asOf is
// folded into equity as "Laba (Rugi) Berjalan" so assets == liabilities+equity.
func (s *Service) BalanceSheet(ctx context.Context, orgID uuid.UUID, asOf time.Time) (*model.BalanceSheet, error) {
	rows, err := s.TrialBalance(ctx, orgID, asOf)
	if err != nil {
		return nil, err
	}
	bs := &model.BalanceSheet{AsOf: asOf.Format("2006-01-02")}
	var netIncome int64
	for _, r := range rows {
		switch r.Type {
		case model.TypeAsset:
			if r.Balance != 0 {
				bs.Assets = append(bs.Assets, model.StatementLine{Code: r.Code, Name: r.Name, Amount: r.Balance})
			}
			bs.TotalAssets += r.Balance
		case model.TypeLiability:
			if r.Balance != 0 {
				bs.Liabilities = append(bs.Liabilities, model.StatementLine{Code: r.Code, Name: r.Name, Amount: r.Balance})
			}
			bs.TotalLiabilities += r.Balance
		case model.TypeEquity:
			if r.Balance != 0 {
				bs.Equity = append(bs.Equity, model.StatementLine{Code: r.Code, Name: r.Name, Amount: r.Balance})
			}
			bs.TotalEquity += r.Balance
		case model.TypeRevenue:
			netIncome += r.Balance
		case model.TypeExpense:
			netIncome -= r.Balance
		}
	}
	if netIncome != 0 {
		bs.Equity = append(bs.Equity, model.StatementLine{Code: "LBR", Name: "Laba (Rugi) Berjalan", Amount: netIncome})
		bs.TotalEquity += netIncome
	}
	bs.Balanced = bs.TotalAssets == bs.TotalLiabilities+bs.TotalEquity
	return bs, nil
}

// IncomeStatement (Laba Rugi) for [from, to].
func (s *Service) IncomeStatement(ctx context.Context, orgID uuid.UUID, from, to time.Time) (*model.IncomeStatement, error) {
	if err := s.EnsureCOA(ctx, orgID); err != nil {
		return nil, err
	}
	rows, err := s.repo.AccountActivity(ctx, orgID, from, to)
	if err != nil {
		return nil, err
	}
	is := &model.IncomeStatement{From: from.Format("2006-01-02"), To: to.Format("2006-01-02")}
	for _, r := range rows {
		switch r.Type {
		case model.TypeRevenue:
			if r.Balance != 0 {
				is.Revenue = append(is.Revenue, model.StatementLine{Code: r.Code, Name: r.Name, Amount: r.Balance})
			}
			is.TotalRevenue += r.Balance
		case model.TypeExpense:
			if r.Balance != 0 {
				is.Expenses = append(is.Expenses, model.StatementLine{Code: r.Code, Name: r.Name, Amount: r.Balance})
			}
			is.TotalExpense += r.Balance
		}
	}
	is.NetIncome = is.TotalRevenue - is.TotalExpense
	return is, nil
}
