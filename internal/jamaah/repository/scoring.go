package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RegRef identifies a single package registration (the unit lead scoring acts
// on: one jamaah may sit in several packages at different stages).
type RegRef struct {
	JamaahID  uuid.UUID
	PackageID uuid.UUID
}

// ScoringBase holds the registration-local signals for lead scoring: the stage
// and when it was entered (drives stage base + staleness decay) plus the
// snapshotted contract value (payment-progress denominator when invoice
// balances aren't on hand, e.g. the event path).
type ScoringBase struct {
	Stage          string
	StageEnteredAt *time.Time
	PriceSnapshot  int64
	DiscountAmount int64
}

// GetScoringBase loads the registration-local scoring signals for one reg.
func (r *JamaahRepo) GetScoringBase(ctx context.Context, orgID, jamaahID, packageID uuid.UUID) (*ScoringBase, error) {
	var b ScoringBase
	err := r.pool.QueryRow(ctx, `SELECT pipeline_status, stage_entered_at, COALESCE(price_snapshot, 0), COALESCE(discount_amount, 0)
		FROM jamaah_package_registrations WHERE org_id = $1 AND jamaah_id = $2 AND package_id = $3`,
		orgID, jamaahID, packageID).Scan(&b.Stage, &b.StageEnteredAt, &b.PriceSnapshot, &b.DiscountAmount)
	if err == pgx.ErrNoRows {
		return nil, ErrRegistrationNotFound
	}
	return &b, err
}

// CountDocuments returns (total, received) document counts for a jamaah, where
// "received" is any document past the belum_diterima state.
func (r *JamaahRepo) CountDocuments(ctx context.Context, orgID, jamaahID uuid.UUID) (total, received int, err error) {
	err = r.pool.QueryRow(ctx, `SELECT COUNT(*), COUNT(*) FILTER (WHERE status <> 'belum_diterima')
		FROM jamaah_documents WHERE org_id = $1 AND jamaah_id = $2`, orgID, jamaahID).Scan(&total, &received)
	return total, received, err
}

// LastTouchAt is the most recent note or follow-up timestamp for a jamaah (the
// recency signal). nil means no contact recorded yet. GREATEST ignores NULLs in
// Postgres, returning NULL only when both sources are empty.
func (r *JamaahRepo) LastTouchAt(ctx context.Context, orgID, jamaahID uuid.UUID) (*time.Time, error) {
	var t *time.Time
	err := r.pool.QueryRow(ctx, `SELECT GREATEST(
		(SELECT MAX(created_at) FROM jamaah_notes WHERE org_id = $1 AND jamaah_id = $2),
		(SELECT MAX(created_at) FROM follow_ups WHERE org_id = $1 AND jamaah_id = $2)
	)`, orgID, jamaahID).Scan(&t)
	return t, err
}

// ListActiveRegistrationIDsForRecompute returns every still-open registration
// for an org (excludes terminal batal/selesai), used by the org-wide recompute
// triggered on payment.received.
func (r *JamaahRepo) ListActiveRegistrationIDsForRecompute(ctx context.Context, orgID uuid.UUID) ([]RegRef, error) {
	rows, err := r.pool.Query(ctx, `SELECT jamaah_id, package_id FROM jamaah_package_registrations
		WHERE org_id = $1 AND pipeline_status NOT IN ('batal', 'selesai')`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	refs := []RegRef{}
	for rows.Next() {
		var ref RegRef
		if err := rows.Scan(&ref.JamaahID, &ref.PackageID); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}

// ListActivePackagesForJamaah returns the package ids of a jamaah's still-open
// registrations, for the in-process recompute after a note/follow-up/document
// change (which carry no package context of their own).
func (r *JamaahRepo) ListActivePackagesForJamaah(ctx context.Context, orgID, jamaahID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `SELECT package_id FROM jamaah_package_registrations
		WHERE org_id = $1 AND jamaah_id = $2 AND pipeline_status NOT IN ('batal', 'selesai')`, orgID, jamaahID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	pkgs := []uuid.UUID{}
	for rows.Next() {
		var pkg uuid.UUID
		if err := rows.Scan(&pkg); err != nil {
			return nil, err
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs, rows.Err()
}

// UpdateLeadScore persists a recomputed score/temperature and stamps
// score_updated_at (the lazy-refresh staleness clock).
func (r *JamaahRepo) UpdateLeadScore(ctx context.Context, orgID, jamaahID, packageID uuid.UUID, score int, temp string) error {
	_, err := r.pool.Exec(ctx, `UPDATE jamaah_package_registrations
		SET lead_score = $4, lead_temp = $5, score_updated_at = NOW()
		WHERE org_id = $1 AND jamaah_id = $2 AND package_id = $3`,
		orgID, jamaahID, packageID, score, temp)
	return err
}
