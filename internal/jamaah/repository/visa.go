package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/jamaah-in/v2/internal/jamaah/model"
	"github.com/jamaah-in/v2/internal/shared/outbox"
)

var ErrVisaNotFound = fmt.Errorf("visa application not found")

const visaCols = `id, org_id, jamaah_id, package_id, status, provider, reference_no,
	submitted_at, decided_at, expiry_date, reject_reason, notes, created_at, updated_at`

func scanVisa(row rowScanner) (*model.VisaApplication, error) {
	v := &model.VisaApplication{}
	err := row.Scan(&v.ID, &v.OrgID, &v.JamaahID, &v.PackageID, &v.Status, &v.Provider, &v.ReferenceNo,
		&v.SubmittedAt, &v.DecidedAt, &v.ExpiryDate, &v.RejectReason, &v.Notes, &v.CreatedAt, &v.UpdatedAt)
	return v, err
}

// GetVisaByJamaah returns a jamaah's single visa application (one per jamaah).
func (r *JamaahRepo) GetVisaByJamaah(ctx context.Context, orgID, jamaahID uuid.UUID) (*model.VisaApplication, error) {
	v, err := scanVisa(r.pool.QueryRow(ctx,
		`SELECT `+visaCols+` FROM visa_applications WHERE org_id = $1 AND jamaah_id = $2`, orgID, jamaahID))
	if err == pgx.ErrNoRows {
		return nil, ErrVisaNotFound
	}
	return v, err
}

// UpsertVisaDraft creates the visa application (status draft) or, if it exists,
// patches its editable fields. Status is never changed here — that's Transition.
func (r *JamaahRepo) UpsertVisaDraft(ctx context.Context, v *model.VisaApplication) (*model.VisaApplication, error) {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO visa_applications (org_id, jamaah_id, package_id, status, provider, reference_no, expiry_date, notes)
		VALUES ($1, $2, $3, 'draft', $4, $5, $6, $7)
		ON CONFLICT (org_id, jamaah_id) DO UPDATE SET
			package_id = COALESCE(EXCLUDED.package_id, visa_applications.package_id),
			provider = EXCLUDED.provider,
			reference_no = EXCLUDED.reference_no,
			expiry_date = COALESCE(EXCLUDED.expiry_date, visa_applications.expiry_date),
			notes = EXCLUDED.notes,
			updated_at = NOW()`,
		v.OrgID, v.JamaahID, v.PackageID, v.Provider, v.ReferenceNo, v.ExpiryDate, v.Notes)
	if err != nil {
		return nil, fmt.Errorf("upsert visa: %w", err)
	}
	return r.GetVisaByJamaah(ctx, v.OrgID, v.JamaahID)
}

// ListVisas returns visa applications for the board, joined with the jamaah name.
func (r *JamaahRepo) ListVisas(ctx context.Context, orgID uuid.UUID, status, search string, offset, limit int) ([]model.VisaApplication, int, error) {
	where := `WHERE v.org_id = $1`
	args := []any{orgID}
	if status != "" {
		args = append(args, status)
		where += fmt.Sprintf(` AND v.status = $%d`, len(args))
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		where += fmt.Sprintf(` AND (p.nama ILIKE $%[1]d OR v.reference_no ILIKE $%[1]d)`, len(args))
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM visa_applications v JOIN jamaah_profiles p ON p.id = v.jamaah_id `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := `SELECT v.id, v.org_id, v.jamaah_id, v.package_id, v.status, v.provider, v.reference_no,
		v.submitted_at, v.decided_at, v.expiry_date, v.reject_reason, v.notes, v.created_at, v.updated_at, p.nama
		FROM visa_applications v JOIN jamaah_profiles p ON p.id = v.jamaah_id ` + where +
		fmt.Sprintf(` ORDER BY v.updated_at DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []model.VisaApplication{}
	for rows.Next() {
		var v model.VisaApplication
		if err := rows.Scan(&v.ID, &v.OrgID, &v.JamaahID, &v.PackageID, &v.Status, &v.Provider, &v.ReferenceNo,
			&v.SubmittedAt, &v.DecidedAt, &v.ExpiryDate, &v.RejectReason, &v.Notes, &v.CreatedAt, &v.UpdatedAt, &v.JamaahName); err != nil {
			return nil, 0, err
		}
		out = append(out, v)
	}
	return out, total, rows.Err()
}

// VisaTransition carries a validated status change to persist.
type VisaTransition struct {
	VisaID       uuid.UUID
	OrgID        uuid.UUID
	JamaahID     uuid.UUID
	FromStatus   string
	ToStatus     string
	SubmittedAt  *time.Time
	DecidedAt    *time.Time
	ExpiryDate   *time.Time
	ReferenceNo  string // applied when non-empty
	RejectReason string
	Reason       string // free-text for history
	ChangedBy    *uuid.UUID
	EventType    string
	Payload      []byte
}

// TransitionVisaTx applies a status change, writes a history row, and enqueues
// the visa.* outbox event — all in one transaction.
func (r *JamaahRepo) TransitionVisaTx(ctx context.Context, t VisaTransition) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ct, err := tx.Exec(ctx, `
		UPDATE visa_applications SET
			status = $3,
			submitted_at = COALESCE($4, submitted_at),
			decided_at = COALESCE($5, decided_at),
			expiry_date = COALESCE($6, expiry_date),
			reference_no = CASE WHEN $7 <> '' THEN $7 ELSE reference_no END,
			reject_reason = $8,
			updated_at = NOW()
		WHERE id = $1 AND org_id = $2`,
		t.VisaID, t.OrgID, t.ToStatus, t.SubmittedAt, t.DecidedAt, t.ExpiryDate, t.ReferenceNo, t.RejectReason)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrVisaNotFound
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO visa_history (org_id, visa_id, jamaah_id, from_status, to_status, reason, changed_by)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, NULLIF($6, ''), $7)`,
		t.OrgID, t.VisaID, t.JamaahID, t.FromStatus, t.ToStatus, t.Reason, t.ChangedBy); err != nil {
		return err
	}

	if t.EventType != "" {
		if err := outbox.Insert(ctx, tx, outbox.Event{
			OrgID:         t.OrgID,
			AggregateType: "visa_application",
			AggregateID:   t.VisaID,
			EventType:     t.EventType,
			Payload:       t.Payload,
		}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ListVisaHistory returns the audit trail for a visa application.
func (r *JamaahRepo) ListVisaHistory(ctx context.Context, orgID, visaID uuid.UUID) ([]model.VisaHistory, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, org_id, visa_id, jamaah_id, from_status, to_status, reason, changed_by, created_at
		FROM visa_history WHERE org_id = $1 AND visa_id = $2 ORDER BY created_at`, orgID, visaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.VisaHistory{}
	for rows.Next() {
		var h model.VisaHistory
		if err := rows.Scan(&h.ID, &h.OrgID, &h.VisaID, &h.JamaahID, &h.FromStatus, &h.ToStatus, &h.Reason, &h.ChangedBy, &h.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// HasReceivedDocument reports whether a jamaah has a given document past the
// belum_diterima state (the gate for submitting a visa: passport must be in).
func (r *JamaahRepo) HasReceivedDocument(ctx context.Context, orgID, jamaahID uuid.UUID, docType string) (bool, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM jamaah_documents
		WHERE org_id = $1 AND jamaah_id = $2 AND doc_type = $3 AND status <> 'belum_diterima'`,
		orgID, jamaahID, docType).Scan(&n)
	return n > 0, err
}

// UpdateProfileVisaFields mirrors an approved visa onto the jamaah profile (so
// the profile's no_visa / tanggal_visa(_akhir) stay in sync for manifests/PDFs).
func (r *JamaahRepo) UpdateProfileVisaFields(ctx context.Context, orgID, jamaahID uuid.UUID, noVisa string, tglVisa, tglVisaAkhir *time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE jamaah_profiles SET
		no_visa = CASE WHEN $3 <> '' THEN $3 ELSE no_visa END,
		tanggal_visa = COALESCE($4, tanggal_visa),
		tanggal_visa_akhir = COALESCE($5, tanggal_visa_akhir),
		updated_at = NOW()
		WHERE id = $2 AND org_id = $1`, orgID, jamaahID, noVisa, tglVisa, tglVisaAkhir)
	return err
}

// ScanVisaExpiredAll finds approved visas whose expiry has passed (for the job
// to auto-transition them to 'expired').
func (r *JamaahRepo) ScanVisaExpiredAll(ctx context.Context) ([]model.VisaApplication, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+visaCols+` FROM visa_applications
		WHERE status = 'approved' AND expiry_date IS NOT NULL AND expiry_date < CURRENT_DATE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.VisaApplication{}
	for rows.Next() {
		v, err := scanVisa(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}

// GetVisaExpiring returns approved visas within withinDays of expiry (per-org,
// for dashboard alerts).
func (r *JamaahRepo) GetVisaExpiring(ctx context.Context, orgID uuid.UUID, withinDays int) ([]model.VisaApplication, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT v.id, v.org_id, v.jamaah_id, v.package_id, v.status, v.provider, v.reference_no,
			v.submitted_at, v.decided_at, v.expiry_date, v.reject_reason, v.notes, v.created_at, v.updated_at, p.nama
		FROM visa_applications v JOIN jamaah_profiles p ON p.id = v.jamaah_id
		WHERE v.org_id = $1 AND v.status = 'approved' AND v.expiry_date IS NOT NULL
		  AND v.expiry_date <= (CURRENT_DATE + make_interval(days => $2)) AND v.expiry_date >= CURRENT_DATE
		ORDER BY v.expiry_date`, orgID, withinDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.VisaApplication{}
	for rows.Next() {
		var v model.VisaApplication
		if err := rows.Scan(&v.ID, &v.OrgID, &v.JamaahID, &v.PackageID, &v.Status, &v.Provider, &v.ReferenceNo,
			&v.SubmittedAt, &v.DecidedAt, &v.ExpiryDate, &v.RejectReason, &v.Notes, &v.CreatedAt, &v.UpdatedAt, &v.JamaahName); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// ExpiringSubject is one item the reminder job may notify about (cross-org).
type ExpiringSubject struct {
	OrgID    uuid.UUID
	JamaahID uuid.UUID
	Name     string
	DaysLeft int
}

// ScanVisaExpiringAll finds approved visas expiring within withinDays across all
// orgs (for the daily reminder job).
func (r *JamaahRepo) ScanVisaExpiringAll(ctx context.Context, withinDays int) ([]ExpiringSubject, error) {
	return r.scanExpiring(ctx, `
		SELECT v.org_id, v.jamaah_id, p.nama, (v.expiry_date - CURRENT_DATE)
		FROM visa_applications v JOIN jamaah_profiles p ON p.id = v.jamaah_id
		WHERE v.status = 'approved' AND v.expiry_date IS NOT NULL
		  AND v.expiry_date <= (CURRENT_DATE + make_interval(days => $1)) AND v.expiry_date >= CURRENT_DATE`, withinDays)
}

// ScanPassportExpiringAll finds passports expiring within withinDays across all
// orgs (passport validity assumed 5 years from issue, matching dashboard alerts).
func (r *JamaahRepo) ScanPassportExpiringAll(ctx context.Context, withinDays int) ([]ExpiringSubject, error) {
	return r.scanExpiring(ctx, `
		SELECT org_id, id, nama, ((tanggal_paspor + INTERVAL '5 years')::date - CURRENT_DATE)
		FROM jamaah_profiles
		WHERE tanggal_paspor IS NOT NULL
		  AND (tanggal_paspor + INTERVAL '5 years')::date <= (CURRENT_DATE + make_interval(days => $1))
		  AND (tanggal_paspor + INTERVAL '5 years')::date >= CURRENT_DATE`, withinDays)
}

func (r *JamaahRepo) scanExpiring(ctx context.Context, query string, withinDays int) ([]ExpiringSubject, error) {
	rows, err := r.pool.Query(ctx, query, withinDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ExpiringSubject{}
	for rows.Next() {
		var s ExpiringSubject
		if err := rows.Scan(&s.OrgID, &s.JamaahID, &s.Name, &s.DaysLeft); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// TryRecordReminder inserts a reminder marker, returning true only if it was new
// (so the daily job notifies each subject+milestone exactly once).
func (r *JamaahRepo) TryRecordReminder(ctx context.Context, orgID, subjectID uuid.UUID, subjectType, milestone string) (bool, error) {
	ct, err := r.pool.Exec(ctx, `
		INSERT INTO lifecycle_reminders (org_id, subject_type, subject_id, milestone)
		VALUES ($1, $2, $3, $4) ON CONFLICT (org_id, subject_type, subject_id, milestone) DO NOTHING`,
		orgID, subjectType, subjectID, milestone)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}
