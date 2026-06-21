package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/invoice/model"
	"github.com/jamaah-in/v2/internal/shared/outbox"
)

var (
	ErrSessionExists   = errors.New("an open cash session already exists for this user")
	ErrSessionNotFound = errors.New("cash session not found")
	ErrSessionClosed   = errors.New("cash session already closed")
)

const sessionCols = `id, org_id, user_id, opening_float, expected_cash, counted_cash, difference, status, opened_at, closed_at, notes`

func scanSession(row interface{ Scan(...any) error }) (*model.CashSession, error) {
	var s model.CashSession
	err := row.Scan(&s.ID, &s.OrgID, &s.UserID, &s.OpeningFloat, &s.ExpectedCash, &s.CountedCash, &s.Difference, &s.Status, &s.OpenedAt, &s.ClosedAt, &s.Notes)
	return &s, err
}

// OpenSession opens a new cash session; fails if the user already has one open.
func (r *InvoiceRepo) OpenSession(ctx context.Context, orgID, userID uuid.UUID, openingFloat int64, notes string) (*model.CashSession, error) {
	s := &model.CashSession{}
	err := r.pool.QueryRow(ctx, `INSERT INTO cash_sessions (org_id, user_id, opening_float, notes)
		VALUES ($1,$2,$3,$4) RETURNING `+sessionCols,
		orgID, userID, openingFloat, notes).Scan(
		&s.ID, &s.OrgID, &s.UserID, &s.OpeningFloat, &s.ExpectedCash, &s.CountedCash, &s.Difference, &s.Status, &s.OpenedAt, &s.ClosedAt, &s.Notes)
	if err != nil {
		if isDuplicate(err) {
			return nil, ErrSessionExists
		}
		return nil, err
	}
	return s, nil
}

// GetActiveSession returns the user's currently-open session, or nil if none.
func (r *InvoiceRepo) GetActiveSession(ctx context.Context, orgID, userID uuid.UUID) (*model.CashSession, error) {
	s, err := scanSession(r.pool.QueryRow(ctx, `SELECT `+sessionCols+` FROM cash_sessions WHERE org_id=$1 AND user_id=$2 AND status='open'`, orgID, userID))
	if err != nil {
		return nil, nil // no active session
	}
	return s, nil
}

// ActiveSessionID returns the id of the user's open session (or nil), used to
// stamp cash payments.
func (r *InvoiceRepo) ActiveSessionID(ctx context.Context, orgID, userID uuid.UUID) *uuid.UUID {
	var id uuid.UUID
	if err := r.pool.QueryRow(ctx, `SELECT id FROM cash_sessions WHERE org_id=$1 AND user_id=$2 AND status='open'`, orgID, userID).Scan(&id); err != nil {
		return nil
	}
	return &id
}

func (r *InvoiceRepo) ListSessions(ctx context.Context, orgID uuid.UUID, limit int) ([]model.CashSession, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+sessionCols+` FROM cash_sessions WHERE org_id=$1 ORDER BY opened_at DESC LIMIT $2`, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.CashSession{}
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *s)
	}
	return out, rows.Err()
}

// CloseSessionTx closes a session: computes expected cash (opening float + cash
// payments taken during the session), records the counted amount + difference,
// and — if there is a discrepancy — emits a pos.cash.session.closed event so the
// accounting engine posts the cash over/short adjustment. All in one tx.
func (r *InvoiceRepo) CloseSessionTx(ctx context.Context, orgID, sessionID uuid.UUID, counted int64, eventType string) (*model.CashSession, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var opening int64
	var status string
	if err := tx.QueryRow(ctx, `SELECT opening_float, status FROM cash_sessions WHERE id=$1 AND org_id=$2 FOR UPDATE`, sessionID, orgID).Scan(&opening, &status); err != nil {
		return nil, ErrSessionNotFound
	}
	if status != "open" {
		return nil, ErrSessionClosed
	}

	var cashTaken int64
	if err := tx.QueryRow(ctx, `SELECT COALESCE(SUM(amount),0) FROM payments WHERE cash_session_id=$1`, sessionID).Scan(&cashTaken); err != nil {
		return nil, err
	}
	expected := opening + cashTaken
	diff := counted - expected

	if _, err := tx.Exec(ctx, `UPDATE cash_sessions SET expected_cash=$3, counted_cash=$4, difference=$5, status='closed', closed_at=NOW()
		WHERE id=$1 AND org_id=$2`, sessionID, orgID, expected, counted, diff); err != nil {
		return nil, err
	}

	if diff != 0 {
		payload, _ := json.Marshal(map[string]any{"difference": diff})
		if err := outbox.Insert(ctx, tx, outbox.Event{
			OrgID:         orgID,
			AggregateType: "cash_session",
			AggregateID:   sessionID,
			EventType:     eventType,
			Payload:       payload,
		}); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	s, err := scanSession(r.pool.QueryRow(ctx, `SELECT `+sessionCols+` FROM cash_sessions WHERE id=$1 AND org_id=$2`, sessionID, orgID))
	if err != nil {
		return nil, err
	}
	return s, nil
}
