package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jamaah-in/v2/internal/jamaah/model"
)

// ErrLimitReached is returned by the transactional create helpers when the
// per-org row count is already at the supplied cap. The service maps it to the
// user-facing plan-limit error.
var ErrLimitReached = errors.New("limit reached")

// Advisory-lock classes namespace the per-org locks so jamaah and group creates
// don't block each other.
const (
	lockClassJamaah int32 = 1
	lockClassGroup  int32 = 2
)

// querier is satisfied by both *pgxpool.Pool and pgx.Tx, letting the insert
// helpers run either standalone or inside a transaction.
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// CreateProfileTx atomically enforces the per-org jamaah cap and inserts the
// profile in a single transaction. A per-org transaction-scoped advisory lock
// serializes concurrent creates for the same org, so the count->insert cannot
// race past the cap (closing the TOCTOU hole in the plain count-then-insert).
// maxJamaah == plan.Unlimited (-1) skips the lock/count entirely. Returns
// ErrLimitReached when the cap is already met.
func (r *JamaahRepo) CreateProfileTx(ctx context.Context, p *model.JamaahProfile, maxJamaah int) error {
	if maxJamaah < 0 { // unlimited — no gate needed
		return insertProfile(ctx, r.pool, p)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1, hashtext($2))`, lockClassJamaah, p.OrgID.String()); err != nil {
		return err
	}
	var n int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM jamaah_profiles WHERE org_id = $1`, p.OrgID).Scan(&n); err != nil {
		return err
	}
	if n >= maxJamaah {
		return ErrLimitReached
	}
	if err := insertProfile(ctx, tx, p); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// CreateGroupTx is the group-create counterpart of CreateProfileTx: it enforces
// the per-org group cap under a per-org advisory lock in one transaction.
func (r *JamaahRepo) CreateGroupTx(ctx context.Context, g *model.Group, maxGroups int) error {
	if maxGroups < 0 { // unlimited — no gate needed
		return insertGroup(ctx, r.pool, g)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1, hashtext($2))`, lockClassGroup, g.OrgID.String()); err != nil {
		return err
	}
	var n int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM groups WHERE org_id = $1`, g.OrgID).Scan(&n); err != nil {
		return err
	}
	if n >= maxGroups {
		return ErrLimitReached
	}
	if err := insertGroup(ctx, tx, g); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
