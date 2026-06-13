package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jamaah-in/v2/internal/accounting/model"
)

var (
	ErrUnbalanced     = errors.New("journal not balanced: sum(debit) != sum(credit)")
	ErrEmptyJournal   = errors.New("journal has no lines")
	ErrUnknownAccount = errors.New("account code not found in org COA")
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

func (r *Repo) Ping(ctx context.Context) error { return r.pool.Ping(ctx) }

// ---- Chart of Accounts ----

// CountAccounts reports how many COA rows an org has (used to decide seeding).
func (r *Repo) CountAccounts(ctx context.Context, orgID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM chart_of_accounts WHERE org_id=$1`, orgID).Scan(&n)
	return n, err
}

// SeedAccounts bulk-inserts the standard COA for an org, skipping any code that
// already exists (idempotent). Runs in one transaction.
func (r *Repo) SeedAccounts(ctx context.Context, orgID uuid.UUID, accts []model.Account) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, a := range accts {
		_, err := tx.Exec(ctx, `INSERT INTO chart_of_accounts (org_id, code, name, type, normal_balance)
			VALUES ($1,$2,$3,$4,$5) ON CONFLICT (org_id, code) DO NOTHING`,
			orgID, a.Code, a.Name, a.Type, a.NormalBalance)
		if err != nil {
			return fmt.Errorf("seed account %s: %w", a.Code, err)
		}
	}
	return tx.Commit(ctx)
}

func (r *Repo) ListAccounts(ctx context.Context, orgID uuid.UUID) ([]model.Account, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, org_id, code, name, type, normal_balance, parent_id, is_active, created_at, updated_at
		FROM chart_of_accounts WHERE org_id=$1 ORDER BY code`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Account{}
	for rows.Next() {
		var a model.Account
		if err := rows.Scan(&a.ID, &a.OrgID, &a.Code, &a.Name, &a.Type, &a.NormalBalance, &a.ParentID, &a.IsActive, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *Repo) CreateAccount(ctx context.Context, a *model.Account) error {
	if a.NormalBalance == "" {
		a.NormalBalance = normalBalanceFor(a.Type)
	}
	err := r.pool.QueryRow(ctx, `INSERT INTO chart_of_accounts (org_id, code, name, type, normal_balance, parent_id, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at, updated_at`,
		a.OrgID, a.Code, a.Name, a.Type, a.NormalBalance, a.ParentID, true).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if isDuplicate(err) {
			return fmt.Errorf("account code %s already exists", a.Code)
		}
		return err
	}
	a.IsActive = true
	return nil
}

func normalBalanceFor(t string) string {
	switch t {
	case model.TypeAsset, model.TypeExpense:
		return model.BalanceDebit
	default:
		return model.BalanceCredit
	}
}

// ---- Posting ----

// PostInput is one balanced journal to persist, addressing accounts by COA code.
type PostInput struct {
	OrgID         uuid.UUID
	JournalNo     string
	Date          time.Time
	SourceModule  string
	SourceRefID   *uuid.UUID
	SourceEventID string // "" for manual/opening journals
	EventType     string
	Description   string
	CreatedBy     *uuid.UUID
	Lines         []model.PostingLine
}

// Post writes a balanced journal + its lines in a single transaction. When
// SourceEventID is set it first claims the event in processed_events
// (ON CONFLICT DO NOTHING): if already claimed it returns posted=false without
// error (idempotent skip). Asserts sum(debit)==sum(credit)>0 before writing.
func (r *Repo) Post(ctx context.Context, in PostInput) (journalID uuid.UUID, posted bool, err error) {
	if len(in.Lines) == 0 {
		return uuid.Nil, false, ErrEmptyJournal
	}
	var sumD, sumC int64
	for _, l := range in.Lines {
		sumD += l.Debit
		sumC += l.Credit
	}
	if sumD != sumC || sumD == 0 {
		return uuid.Nil, false, ErrUnbalanced
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, false, err
	}
	defer tx.Rollback(ctx)

	// Idempotency: claim the event first.
	if in.SourceEventID != "" {
		tag, derr := tx.Exec(ctx, `INSERT INTO processed_events (event_id, org_id, event_type)
			VALUES ($1,$2,$3) ON CONFLICT (event_id) DO NOTHING`, in.SourceEventID, in.OrgID, in.EventType)
		if derr != nil {
			return uuid.Nil, false, fmt.Errorf("claim event: %w", derr)
		}
		if tag.RowsAffected() == 0 {
			// already processed — nothing to do
			return uuid.Nil, false, nil
		}
	}

	// Resolve account codes → ids for this org.
	codeToID, err := loadAccountIDs(ctx, tx, in.OrgID)
	if err != nil {
		return uuid.Nil, false, err
	}

	jid := uuid.New()
	var srcEvent *string
	if in.SourceEventID != "" {
		srcEvent = &in.SourceEventID
	}
	if in.Date.IsZero() {
		in.Date = time.Now()
	}
	_, err = tx.Exec(ctx, `INSERT INTO journals (id, org_id, journal_no, journal_date, source_module, source_ref_id, source_event_id, description, status, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'posted',$9)`,
		jid, in.OrgID, in.JournalNo, in.Date, in.SourceModule, in.SourceRefID, srcEvent, in.Description, in.CreatedBy)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("insert journal: %w", err)
	}

	for _, l := range in.Lines {
		accID, ok := codeToID[l.AccountCode]
		if !ok {
			return uuid.Nil, false, fmt.Errorf("%w: %s", ErrUnknownAccount, l.AccountCode)
		}
		_, err = tx.Exec(ctx, `INSERT INTO journal_lines (journal_id, org_id, account_id, debit, credit, memo)
			VALUES ($1,$2,$3,$4,$5,$6)`, jid, in.OrgID, accID, l.Debit, l.Credit, l.Memo)
		if err != nil {
			return uuid.Nil, false, fmt.Errorf("insert line %s: %w", l.AccountCode, err)
		}
	}

	if in.SourceEventID != "" {
		if _, err = tx.Exec(ctx, `UPDATE processed_events SET journal_id=$2 WHERE event_id=$1`, in.SourceEventID, jid); err != nil {
			return uuid.Nil, false, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return uuid.Nil, false, err
	}
	return jid, true, nil
}

func loadAccountIDs(ctx context.Context, tx pgx.Tx, orgID uuid.UUID) (map[string]uuid.UUID, error) {
	rows, err := tx.Query(ctx, `SELECT code, id FROM chart_of_accounts WHERE org_id=$1`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[string]uuid.UUID{}
	for rows.Next() {
		var code string
		var id uuid.UUID
		if err := rows.Scan(&code, &id); err != nil {
			return nil, err
		}
		m[code] = id
	}
	return m, rows.Err()
}

// ---- Journals (read) ----

func (r *Repo) ListJournals(ctx context.Context, orgID uuid.UUID, offset, limit int) ([]model.Journal, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM journals WHERE org_id=$1`, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `SELECT id, org_id, journal_no, journal_date, source_module, source_ref_id, source_event_id, description, status, created_by, created_at
		FROM journals WHERE org_id=$1 ORDER BY journal_date DESC, created_at DESC LIMIT $2 OFFSET $3`, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []model.Journal{}
	for rows.Next() {
		var j model.Journal
		if err := rows.Scan(&j.ID, &j.OrgID, &j.JournalNo, &j.JournalDate, &j.SourceModule, &j.SourceRefID, &j.SourceEventID, &j.Description, &j.Status, &j.CreatedBy, &j.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, j)
	}
	return out, total, rows.Err()
}

func (r *Repo) GetJournal(ctx context.Context, orgID, id uuid.UUID) (*model.Journal, error) {
	var j model.Journal
	err := r.pool.QueryRow(ctx, `SELECT id, org_id, journal_no, journal_date, source_module, source_ref_id, source_event_id, description, status, created_by, created_at
		FROM journals WHERE org_id=$1 AND id=$2`, orgID, id).Scan(
		&j.ID, &j.OrgID, &j.JournalNo, &j.JournalDate, &j.SourceModule, &j.SourceRefID, &j.SourceEventID, &j.Description, &j.Status, &j.CreatedBy, &j.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("journal not found")
	}
	lines, err := r.journalLines(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	j.Lines = lines
	return &j, nil
}

func (r *Repo) journalLines(ctx context.Context, journalID, orgID uuid.UUID) ([]model.JournalLine, error) {
	rows, err := r.pool.Query(ctx, `SELECT l.id, l.journal_id, l.org_id, l.account_id, a.code, a.name, l.debit, l.credit, l.memo
		FROM journal_lines l JOIN chart_of_accounts a ON a.id=l.account_id
		WHERE l.journal_id=$1 AND l.org_id=$2 ORDER BY l.debit DESC`, journalID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.JournalLine{}
	for rows.Next() {
		var l model.JournalLine
		if err := rows.Scan(&l.ID, &l.JournalID, &l.OrgID, &l.AccountID, &l.AccountCode, &l.AccountName, &l.Debit, &l.Credit, &l.Memo); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// GeneralLedger returns posted lines for one account within a date range.
func (r *Repo) GeneralLedger(ctx context.Context, orgID, accountID uuid.UUID, from, to time.Time) ([]model.JournalLine, error) {
	rows, err := r.pool.Query(ctx, `SELECT l.id, l.journal_id, l.org_id, l.account_id, a.code, a.name, l.debit, l.credit, l.memo
		FROM journal_lines l
		JOIN chart_of_accounts a ON a.id=l.account_id
		JOIN journals j ON j.id=l.journal_id
		WHERE l.org_id=$1 AND l.account_id=$2 AND j.status='posted' AND j.journal_date BETWEEN $3 AND $4
		ORDER BY j.journal_date, j.created_at`, orgID, accountID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.JournalLine{}
	for rows.Next() {
		var l model.JournalLine
		if err := rows.Scan(&l.ID, &l.JournalID, &l.OrgID, &l.AccountID, &l.AccountCode, &l.AccountName, &l.Debit, &l.Credit, &l.Memo); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// TrialBalance returns each account's summed debit/credit from posted journals
// up to and including asOf.
func (r *Repo) TrialBalance(ctx context.Context, orgID uuid.UUID, asOf time.Time) ([]model.TrialBalanceRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id, a.code, a.name, a.type, a.normal_balance,
		       COALESCE(SUM(l.debit),0), COALESCE(SUM(l.credit),0)
		FROM chart_of_accounts a
		LEFT JOIN journal_lines l ON l.account_id=a.id AND l.org_id=a.org_id
		LEFT JOIN journals j ON j.id=l.journal_id AND j.status='posted' AND j.journal_date <= $2
		WHERE a.org_id=$1
		GROUP BY a.id, a.code, a.name, a.type, a.normal_balance
		ORDER BY a.code`, orgID, asOf)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.TrialBalanceRow{}
	for rows.Next() {
		var row model.TrialBalanceRow
		var normal string
		if err := rows.Scan(&row.AccountID, &row.Code, &row.Name, &row.Type, &normal, &row.Debit, &row.Credit); err != nil {
			return nil, err
		}
		if normal == model.BalanceDebit {
			row.Balance = row.Debit - row.Credit
		} else {
			row.Balance = row.Credit - row.Debit
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// AccountActivity returns per-account debit/credit summed from posted journals
// within [from, to] — used for period reports (Laba Rugi / Arus Kas).
func (r *Repo) AccountActivity(ctx context.Context, orgID uuid.UUID, from, to time.Time) ([]model.TrialBalanceRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id, a.code, a.name, a.type, a.normal_balance,
		       COALESCE(SUM(l.debit),0), COALESCE(SUM(l.credit),0)
		FROM chart_of_accounts a
		LEFT JOIN journal_lines l ON l.account_id=a.id AND l.org_id=a.org_id
		LEFT JOIN journals j ON j.id=l.journal_id AND j.status='posted' AND j.journal_date BETWEEN $2 AND $3
		WHERE a.org_id=$1
		GROUP BY a.id, a.code, a.name, a.type, a.normal_balance
		ORDER BY a.code`, orgID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.TrialBalanceRow{}
	for rows.Next() {
		var row model.TrialBalanceRow
		var normal string
		if err := rows.Scan(&row.AccountID, &row.Code, &row.Name, &row.Type, &normal, &row.Debit, &row.Credit); err != nil {
			return nil, err
		}
		if normal == model.BalanceDebit {
			row.Balance = row.Debit - row.Credit
		} else {
			row.Balance = row.Credit - row.Debit
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func isDuplicate(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return contains(s, "duplicate key") || contains(s, "unique constraint")
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
