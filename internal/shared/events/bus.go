package events

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Bus is a connected JetStream client. It is safe for concurrent publishing.
type Bus struct {
	nc  *nats.Conn
	js  nats.JetStreamContext
	log *zap.SugaredLogger
}

// Connect dials NATS, obtains a JetStream context, and ensures the SULUK_EVENTS
// stream exists (file-backed, 30-day retention, 10-minute publish-dedup window).
// Reconnect is infinite so a NATS restart never permanently breaks producers.
func Connect(addr string, log *zap.SugaredLogger) (*Bus, error) {
	nc, err := nats.Connect(addr,
		nats.Name("suluk"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("nats jetstream: %w", err)
	}

	b := &Bus{nc: nc, js: js, log: log}
	if err := b.ensureStream(); err != nil {
		nc.Close()
		return nil, err
	}
	return b, nil
}

func (b *Bus) ensureStream() error {
	cfg := &nats.StreamConfig{
		Name:       StreamName,
		Subjects:   []string{SubjectAll},
		Storage:    nats.FileStorage,
		Retention:  nats.LimitsPolicy,
		MaxAge:     30 * 24 * time.Hour,
		Duplicates: 10 * time.Minute, // Msg-Id dedup window
	}
	if _, err := b.js.StreamInfo(StreamName); err != nil {
		if errors.Is(err, nats.ErrStreamNotFound) {
			if _, aerr := b.js.AddStream(cfg); aerr != nil {
				return fmt.Errorf("add stream: %w", aerr)
			}
			return nil
		}
		return fmt.Errorf("stream info: %w", err)
	}
	// Stream exists — keep subjects/retention in sync (idempotent).
	if _, err := b.js.UpdateStream(cfg); err != nil {
		// Non-fatal: an older server may reject a no-op update; log and continue.
		if b.log != nil {
			b.log.Warnf("update stream %s: %v", StreamName, err)
		}
	}
	return nil
}

// Publish sends an envelope. The EventID becomes the JetStream Msg-Id, so a
// re-published outbox row (relay retry) is de-duplicated by the server.
func (b *Bus) Publish(env *Envelope) error {
	data, err := env.Marshal()
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}
	_, err = b.js.Publish(Subject(env.EventType), data, nats.MsgId(env.EventID))
	if err != nil {
		return fmt.Errorf("publish %s: %w", env.EventType, err)
	}
	return nil
}

// Handler processes one event. Returning nil ACKs the message; a non-nil error
// triggers NAK + redelivery (with backoff), until maxDeliver attempts are
// exhausted, after which the message is Term'd (dropped) to avoid head-of-line
// blocking the whole stream on one poison event.
type Handler func(ctx context.Context, env *Envelope) error

const (
	maxDeliver = 6
	nakBackoff = 5 * time.Second
)

// Subscribe creates a durable push subscription over `suluk.>` and invokes the
// handler for each message. MaxAckPending(1) keeps ordered, one-at-a-time
// delivery — correctness over throughput for journaling. The consumer dedups
// via processed_events, so redelivery after a NAK is safe.
func (b *Bus) Subscribe(ctx context.Context, durable string, handler Handler) (*nats.Subscription, error) {
	sub, err := b.js.Subscribe(SubjectAll, func(msg *nats.Msg) {
		env, perr := ParseEnvelope(msg.Data)
		if perr != nil {
			if b.log != nil {
				b.log.Errorf("events: bad envelope, terminating msg: %v", perr)
			}
			_ = msg.Term() // unparseable — never redeliver
			return
		}
		if herr := handler(ctx, env); herr != nil {
			delivered := uint64(1)
			if meta, merr := msg.Metadata(); merr == nil && meta != nil {
				delivered = meta.NumDelivered
			}
			if delivered >= maxDeliver {
				// Poison event: drop it so the stream isn't blocked. A daily
				// GL-vs-source reconciliation job surfaces the missing journal.
				if b.log != nil {
					b.log.Errorf("events: POISON %s (event_id=%s) dropped after %d attempts: %v", env.EventType, env.EventID, delivered, herr)
				}
				_ = msg.Term()
				return
			}
			if b.log != nil {
				b.log.Warnf("events: handler failed for %s (event_id=%s) attempt %d, retrying: %v", env.EventType, env.EventID, delivered, herr)
			}
			_ = msg.NakWithDelay(nakBackoff)
			return
		}
		_ = msg.Ack()
	},
		nats.Durable(durable),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.DeliverAll(),
		nats.MaxAckPending(1),
		nats.AckWait(30*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("subscribe: %w", err)
	}
	return sub, nil
}

// Close drains and closes the underlying connection.
func (b *Bus) Close() {
	if b.nc != nil {
		_ = b.nc.Drain()
	}
}
