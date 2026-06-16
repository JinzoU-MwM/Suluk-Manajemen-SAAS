// Package model holds the accounting domain types: chart of accounts, journals,
// journal lines, and the report shapes (trial balance, neraca, laba rugi).
package model

import (
	"time"

	"github.com/google/uuid"
)

// Account types and normal balances.
const (
	TypeAsset     = "asset"
	TypeLiability = "liability"
	TypeEquity    = "equity"
	TypeRevenue   = "revenue"
	TypeExpense   = "expense"

	BalanceDebit  = "debit"
	BalanceCredit = "credit"

	StatusPosted = "posted"
	StatusVoid   = "void"
	StatusDraft  = "draft"
)

// Account is one Chart-of-Accounts row.
type Account struct {
	ID            uuid.UUID  `json:"id"`
	OrgID         uuid.UUID  `json:"org_id"`
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	NormalBalance string     `json:"normal_balance"`
	ParentID      *uuid.UUID `json:"parent_id,omitempty"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// JournalLine is one debit or credit posting against an account.
type JournalLine struct {
	ID         uuid.UUID `json:"id"`
	JournalID  uuid.UUID `json:"journal_id"`
	OrgID      uuid.UUID `json:"org_id"`
	AccountID  uuid.UUID `json:"account_id"`
	AccountCode string   `json:"account_code,omitempty"`
	AccountName string   `json:"account_name,omitempty"`
	Debit      int64     `json:"debit"`
	Credit     int64     `json:"credit"`
	Memo       string    `json:"memo"`
}

// Journal is a balanced set of lines from a single source transaction/event.
type Journal struct {
	ID            uuid.UUID     `json:"id"`
	OrgID         uuid.UUID     `json:"org_id"`
	JournalNo     string        `json:"journal_no"`
	JournalDate   time.Time     `json:"journal_date"`
	SourceModule  string        `json:"source_module"`
	SourceRefID   *uuid.UUID    `json:"source_ref_id,omitempty"`
	SourceEventID *string       `json:"source_event_id,omitempty"`
	Description   string        `json:"description"`
	Status        string        `json:"status"`
	CreatedBy     *uuid.UUID    `json:"created_by,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	Lines         []JournalLine `json:"lines,omitempty"`
}

// PostingLine is an instruction from the posting engine: a debit or credit on a
// COA account identified by its code, with an amount.
type PostingLine struct {
	AccountCode string
	Debit       int64
	Credit      int64
	Memo        string
}

// ---- Report shapes ----

// TrialBalanceRow is one account's net debit/credit balance.
type TrialBalanceRow struct {
	AccountID   uuid.UUID `json:"account_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Debit       int64     `json:"debit"`
	Credit      int64     `json:"credit"`
	Balance     int64     `json:"balance"` // signed by normal balance
}

// StatementLine is an account's amount in a financial statement.
type StatementLine struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Amount int64  `json:"amount"`
}

// BalanceSheet (Neraca).
type BalanceSheet struct {
	AsOf             string          `json:"as_of"`
	Assets           []StatementLine `json:"assets"`
	Liabilities      []StatementLine `json:"liabilities"`
	Equity           []StatementLine `json:"equity"`
	TotalAssets      int64           `json:"total_assets"`
	TotalLiabilities int64           `json:"total_liabilities"`
	TotalEquity      int64           `json:"total_equity"`
	Balanced         bool            `json:"balanced"`
}

// IncomeStatement (Laba Rugi).
type IncomeStatement struct {
	From         string          `json:"from"`
	To           string          `json:"to"`
	Revenue      []StatementLine `json:"revenue"`
	Expenses     []StatementLine `json:"expenses"`
	TotalRevenue int64           `json:"total_revenue"`
	TotalExpense int64           `json:"total_expense"`
	NetIncome    int64           `json:"net_income"`
}

// Insight is one finding from the financial-insights engine.
type Insight struct {
	Severity string `json:"severity"` // critical|warning|info|good
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

// InsightMetrics are the headline numbers the insights are derived from.
type InsightMetrics struct {
	TotalAssets          int64 `json:"total_assets"`
	TotalLiabilities     int64 `json:"total_liabilities"`
	TotalEquity          int64 `json:"total_equity"`
	Cash                 int64 `json:"cash"`        // 1101 + 1102
	Receivables          int64 `json:"receivables"` // 1201
	Revenue              int64 `json:"revenue"`
	Expense              int64 `json:"expense"`
	NetIncome            int64 `json:"net_income"`
	BalanceSheetBalanced bool  `json:"balance_sheet_balanced"`
	LedgerBalanced       bool  `json:"ledger_balanced"`
}

// InsightReport is the financial-insights payload (rule-based + optional AI).
type InsightReport struct {
	AsOf        string         `json:"as_of"`
	PeriodFrom  string         `json:"period_from"`
	PeriodTo    string         `json:"period_to"`
	Metrics     InsightMetrics `json:"metrics"`
	Anomalies   []Insight      `json:"anomalies"`
	Highlights  []Insight      `json:"highlights"`
	AINarrative string         `json:"ai_narrative"`
	AIAvailable bool           `json:"ai_available"`
}
