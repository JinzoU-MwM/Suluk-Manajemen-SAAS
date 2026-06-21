package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jamaah-in/v2/internal/shared/outbox"
	"github.com/jamaah-in/v2/internal/tabungan/model"
)

var (
	ErrNotFound      = errors.New("savings account not found")
	ErrInsufficient  = errors.New("insufficient savings balance")
	ErrNotActive     = errors.New("savings account is not active")
	ErrInvalidAmount = errors.New("amount harus lebih dari 0")
)

type Repo struct{ pool *pgxpool.Pool }

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

func (r *Repo) Ping(ctx context.Context) error { return r.pool.Ping(ctx) }

func (r *Repo) CreateAccount(ctx context.Context, a *model.SavingsAccount) error {
	return r.pool.QueryRow(ctx, `INSERT INTO savings_accounts (org_id, jamaah_id, jamaah_name, target_package_id, target_amount, notes)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, balance, status, created_at, updated_at`,
		a.OrgID, a.JamaahID, a.JamaahName, a.TargetPackageID, a.TargetAmount, a.Notes).
		Scan(&a.ID, &a.Balance, &a.Status, &a.CreatedAt, &a.UpdatedAt)
}

const acctCols = `id, org_id, jamaah_id, jamaah_name, target_package_id, target_amount, balance, status, notes, created_at, updated_at`

func scanAccount(row interface{ Scan(...any) error }) (*model.SavingsAccount, error) {
	var a model.SavingsAccount
	err := row.Scan(&a.ID, &a.OrgID, &a.JamaahID, &a.JamaahName, &a.TargetPackageID, &a.TargetAmount, &a.Balance, &a.Status, &a.Notes, &a.CreatedAt, &a.UpdatedAt)
	return &a, err
}

func (r *Repo) ListAccounts(ctx context.Context, orgID uuid.UUID, offset, limit int) ([]model.SavingsAccount, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM savings_accounts WHERE org_id=$1`, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`SELECT %s FROM savings_accounts WHERE org_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, acctCols), orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []model.SavingsAccount{}
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *a)
	}
	return out, total, rows.Err()
}

func (r *Repo) GetAccount(ctx context.Context, orgID, id uuid.UUID) (*model.SavingsAccount, error) {
	a, err := scanAccount(r.pool.QueryRow(ctx, fmt.Sprintf(`SELECT %s FROM savings_accounts WHERE org_id=$1 AND id=$2`, acctCols), orgID, id))
	if err != nil {
		return nil, ErrNotFound
	}
	deps, err := r.listDeposits(ctx, id)
	if err != nil {
		return nil, err
	}
	a.Deposits = deps
	return a, nil
}

func (r *Repo) listDeposits(ctx context.Context, accountID uuid.UUID) ([]model.Deposit, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, account_id, org_id, amount, direction, type, method, reference, notes, created_by, created_at
		FROM savings_deposits WHERE account_id=$1 ORDER BY created_at DESC`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Deposit{}
	for rows.Next() {
		var d model.Deposit
		if err := rows.Scan(&d.ID, &d.AccountID, &d.OrgID, &d.Amount, &d.Direction, &d.Type, &d.Method, &d.Reference, &d.Notes, &d.CreatedBy, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// DepositTx inserts a deposit, increases the balance, and writes a
// savings.deposited outbox event — atomically.
func (r *Repo) DepositTx(ctx context.Context, orgID, accountID uuid.UUID, d *model.Deposit, evt outbox.Event) (*model.SavingsAccount, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// lock + verify account active
	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM savings_accounts WHERE id=$1 AND org_id=$2 FOR UPDATE`, accountID, orgID).Scan(&status); err != nil {
		return nil, ErrNotFound
	}
	if status != model.StatusAktif {
		return nil, ErrNotActive
	}
	if _, err := tx.Exec(ctx, `INSERT INTO savings_deposits (account_id, org_id, amount, direction, type, method, reference, notes, created_by)
		VALUES ($1,$2,$3,'in','setor',$4,$5,$6,$7)`,
		accountID, orgID, d.Amount, d.Method, d.Reference, d.Notes, d.CreatedBy); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE savings_accounts SET balance = balance + $3, updated_at = NOW() WHERE id=$1 AND org_id=$2`, accountID, orgID, d.Amount); err != nil {
		return nil, err
	}
	if err := outbox.Insert(ctx, tx, evt); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetAccount(ctx, orgID, accountID)
}

// ConvertTx decrements the balance by amount, records a 'konversi' out movement,
// flips status to converted when balance hits 0, and writes a savings.converted
// outbox event — atomically. The caller settles the invoice separately.
func (r *Repo) ConvertTx(ctx context.Context, orgID, accountID uuid.UUID, amount int64, createdBy *uuid.UUID, ref string, evt outbox.Event) (*model.SavingsAccount, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var balance int64
	var status string
	if err := tx.QueryRow(ctx, `SELECT balance, status FROM savings_accounts WHERE id=$1 AND org_id=$2 FOR UPDATE`, accountID, orgID).Scan(&balance, &status); err != nil {
		return nil, ErrNotFound
	}
	if status != model.StatusAktif {
		return nil, ErrNotActive
	}
	if amount > balance {
		return nil, ErrInsufficient
	}
	if _, err := tx.Exec(ctx, `INSERT INTO savings_deposits (account_id, org_id, amount, direction, type, method, reference, notes, created_by)
		VALUES ($1,$2,$3,'out','konversi','tabungan',$4,'Konversi ke pelunasan invoice',$5)`,
		accountID, orgID, amount, ref, createdBy); err != nil {
		return nil, err
	}
	newStatus := status
	if balance-amount == 0 {
		newStatus = model.StatusConverted
	}
	if _, err := tx.Exec(ctx, `UPDATE savings_accounts SET balance = balance - $3, status=$4, updated_at = NOW() WHERE id=$1 AND org_id=$2`, accountID, orgID, amount, newStatus); err != nil {
		return nil, err
	}
	if err := outbox.Insert(ctx, tx, evt); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetAccount(ctx, orgID, accountID)
}
