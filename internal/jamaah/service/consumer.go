package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/shared/events"
)

// StartConsumer subscribes the lead-scoring consumer to the event bus. It shares
// the single suluk.> stream with the accounting poster but uses its own durable
// ("jamaah-scoring") so the two consume independently.
//
// Only payment.received matters here: a payment shifts a lead's payment-progress
// signal. The event carries no jamaah_id (its aggregate is the invoice), so
// rather than resolve invoice→jamaah cross-service we recompute every active
// registration for the event's org — scoring is idempotent and cheap, so a full
// org sweep is simpler and needs no dedup. Every other event type is ACKed
// untouched.
func (s *JamaahService) StartConsumer(ctx context.Context, bus *events.Bus) error {
	_, err := bus.Subscribe(ctx, "jamaah-scoring", func(ctx context.Context, env *events.Envelope) error {
		if env.EventType != events.EventPaymentReceived {
			return nil // not ours — ACK and move on
		}
		orgID, err := uuid.Parse(env.OrgID)
		if err != nil {
			// Unparseable org id can never succeed on redelivery — ACK to avoid a
			// poison-message loop.
			if s.log != nil {
				s.log.Warnw("jamaah-scoring: bad org_id in event", "event_id", env.EventID, "org_id", env.OrgID)
			}
			return nil
		}
		if err := s.RecomputeOrgActive(ctx, orgID); err != nil {
			return err // NAK → redeliver
		}
		if s.log != nil {
			s.log.Infow("lead scores recomputed from payment", "event_id", env.EventID, "org_id", env.OrgID)
		}
		return nil
	})
	return err
}
