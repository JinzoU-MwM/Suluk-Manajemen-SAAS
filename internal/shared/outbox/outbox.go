// Package outbox is the shared transactional-outbox toolkit used by every event
// producer (invoice, vendor, payroll, agent, ...). A producer inserts an event
// row in the SAME transaction as its business write (Insert), and a Relay drains
// unpublished rows to the Integration Bus. Every producer DB has an identical
// `outbox` table (see each service's outbox migration).
package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/jamaah-in/v2/internal/shared/events"
)

// Querier is satisfied by both *pgxpool.Pool and pgx.Tx, so Insert can run inside
// the caller's business transaction.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Event is a domain event to enqueue for publication.
type Event struct {
	OrgID         uuid.UUID
	AggregateType string
	AggregateID   uuid.UUID
	EventType     string
	Payload       []byte
	OccurredAt    time.Time
}

// Insert writes one outbox row using the supplied querier (pool or tx). Call it
// inside the same transaction as the business write to avoid lost events.
func Insert(ctx context.Context, q Querier, e Event) error {
	when := e.OccurredAt
	if when.IsZero() {
		when = time.Now()
	}
	_, err := q.Exec(ctx, `INSERT INTO outbox (org_id, aggregate_type, aggregate_id, event_type, payload, occurred_at)
		VALUES ($1,$2,$3,$4,$5,$6)`, e.OrgID, e.AggregateType, e.AggregateID, e.EventType, e.Payload, when)
	return err
}

// Row is an outbox record fetched for publication.
type Row struct {
	ID            uuid.UUID
	OrgID         uuid.UUID
	AggregateType string
	AggregateID   uuid.UUID
	EventType     string
	Payload       []byte
	OccurredAt    time.Time
}

// Store reads/updates the outbox table on a pool.
type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) FetchUnpublished(ctx context.Context, limit int) ([]Row, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, org_id, aggregate_type, aggregate_id, event_type, payload, occurred_at
		FROM outbox WHERE published_at IS NULL ORDER BY occurred_at LIMIT $1 FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Row{}
	for rows.Next() {
		var r Row
		if err := rows.Scan(&r.ID, &r.OrgID, &r.AggregateType, &r.AggregateID, &r.EventType, &r.Payload, &r.OccurredAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) MarkPublished(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `UPDATE outbox SET published_at = NOW(), attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

func (s *Store) MarkAttempt(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `UPDATE outbox SET attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

// ensure *pgxpool.Pool / pgx.Tx satisfy Querier at compile time.
var _ Querier = (*pgxpool.Pool)(nil)
var _ Querier = (pgx.Tx)(nil)

// Relay drains a Store to the bus on a ticker. One per service. Publishing is
// idempotent at the broker (Msg-Id = outbox row id), so re-publish is safe.
type Relay struct {
	store *Store
	bus   *events.Bus
	log   *zap.SugaredLogger
	name  string
}

func NewRelay(store *Store, bus *events.Bus, log *zap.SugaredLogger, name string) *Relay {
	return &Relay{store: store, bus: bus, log: log, name: name}
}

func (r *Relay) Start(ctx context.Context, interval time.Duration) {
	r.log.Infof("starting %s outbox relay (interval: %s)", r.name, interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	r.processBatch(ctx)
	for {
		select {
		case <-ticker.C:
			r.processBatch(ctx)
		case <-ctx.Done():
			r.log.Infof("%s outbox relay stopped", r.name)
			return
		}
	}
}

func (r *Relay) processBatch(ctx context.Context) {
	rows, err := r.store.FetchUnpublished(ctx, 100)
	if err != nil {
		r.log.Errorf("relay fetch outbox: %v", err)
		return
	}
	for _, e := range rows {
		env := &events.Envelope{
			EventID:       e.ID.String(),
			EventType:     e.EventType,
			OrgID:         e.OrgID.String(),
			AggregateType: e.AggregateType,
			AggregateID:   e.AggregateID.String(),
			OccurredAt:    e.OccurredAt,
			Payload:       e.Payload,
		}
		if err := r.bus.Publish(env); err != nil {
			r.log.Errorf("relay publish %s (%s): %v", e.EventType, e.ID, err)
			_ = r.store.MarkAttempt(ctx, e.ID)
			continue
		}
		if err := r.store.MarkPublished(ctx, e.ID); err != nil {
			r.log.Errorf("relay mark published %s: %v", e.ID, err)
		}
	}
}
