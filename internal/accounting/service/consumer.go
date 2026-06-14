package service

import (
	"context"

	"github.com/jamaah-in/v2/internal/shared/events"
)

// StartConsumer subscribes the accounting poster to the event bus. Each event is
// mapped to a balanced journal and posted idempotently; a handler error NAKs the
// message for redelivery (safe — posting dedups via processed_events).
func (s *Service) StartConsumer(ctx context.Context, bus *events.Bus) error {
	_, err := bus.Subscribe(ctx, "accounting-poster", func(ctx context.Context, env *events.Envelope) error {
		posted, err := s.PostFromEvent(ctx, env)
		if err != nil {
			return err // NAK → redeliver later
		}
		if posted && s.log != nil {
			s.log.Infow("journal posted from event",
				"event_id", env.EventID, "event_type", env.EventType, "org_id", env.OrgID)
		}
		return nil
	})
	return err
}
