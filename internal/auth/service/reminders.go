package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jamaah-in/v2/internal/auth/model"
	"github.com/jamaah-in/v2/internal/shared/plan"
)

// reminderThresholds are the days-before-expiry marks, ascending (most-urgent first).
var reminderThresholds = []int{1, 3, 7}

// dueReminder decides which renewal reminder is due for a sub daysLeft from
// expiry, given thresholds already sent this cycle. It returns the single
// most-urgent unsent threshold to EMAIL, the full set of currently-applicable
// thresholds to MARK (so a larger window entered late never fires out of order),
// and ok=false when nothing new is due.
func dueReminder(daysLeft float64, sent []int) (emailThreshold int, markThresholds []int, ok bool) {
	sentSet := make(map[int]bool, len(sent))
	for _, t := range sent {
		sentSet[t] = true
	}
	for _, t := range reminderThresholds { // ascending
		if daysLeft <= float64(t) {
			markThresholds = append(markThresholds, t)
			if emailThreshold == 0 && !sentSet[t] {
				emailThreshold = t // smallest unsent applicable
			}
		}
	}
	if emailThreshold == 0 {
		return 0, nil, false
	}
	return emailThreshold, markThresholds, true
}

// renderRenewalEmail builds the Indonesian renewal-reminder subject + HTML body.
func renderRenewalEmail(planKey string, daysLeft int, expiresAt time.Time, renewURL string) (subject, html string) {
	name := plan.Get(planKey).Name
	subject = fmt.Sprintf("Paket %s Anda berakhir dalam %d hari", name, daysLeft)
	html = fmt.Sprintf(`<p>Halo,</p>
<p>Paket <strong>%s</strong> Anda akan berakhir dalam <strong>%d hari</strong> (pada %s).</p>
<p>Perpanjang sekarang biar fitur gak keputus dan akun gak turun ke paket Gratis.</p>
<p><a href="%s">Perpanjang paket</a></p>`,
		name, daysLeft, expiresAt.Format("02-01-2006"), renewURL)
	return subject, html
}

// runRenewalReminders sends due H-7/3/1 reminders for paid subs nearing expiry.
// Best-effort per sub; a threshold is marked only after a successful email (or
// when the owner opted out via notify_expiry), so transient failures retry.
func (s *AuthService) runRenewalReminders(ctx context.Context, publicURL string) {
	subs, err := s.repo.ListExpiringSubscriptions(ctx)
	if err != nil {
		log.Printf("renewal reminders: list expiring subs: %v", err)
		return
	}
	now := time.Now()
	for _, sub := range subs {
		daysLeft := sub.ExpiresAt.Sub(now).Hours() / 24
		sent, err := s.repo.SentReminderThresholds(ctx, sub.OrgID, sub.ExpiresAt)
		if err != nil {
			log.Printf("renewal reminders: sent thresholds (org %s): %v", sub.OrgID, err)
			continue
		}
		emailT, markTs, ok := dueReminder(daysLeft, sent)
		if !ok {
			continue
		}
		email, notifyExpiry, err := s.repo.GetOrgOwner(ctx, sub.OrgID)
		if err != nil {
			log.Printf("renewal reminders: owner (org %s): %v", sub.OrgID, err)
			continue
		}

		marked := false
		if !notifyExpiry {
			marked = true // opted out: mark so we don't re-evaluate hourly
		} else if email != "" && s.email != nil && s.email.Enabled() {
			subject, html := renderRenewalEmail(sub.Plan, emailT, sub.ExpiresAt, publicURL)
			if err := s.email.Send(ctx, email, subject, html); err != nil {
				log.Printf("renewal reminders: email (org %s): %v", sub.OrgID, err)
			} else {
				_ = s.CreateNotification(ctx, &model.Notification{
					OrgID:    sub.OrgID,
					Severity: "warning",
					Title:    "Paket akan berakhir",
					Message:  fmt.Sprintf("Paket %s Anda berakhir dalam %d hari. Perpanjang biar fitur gak keputus.", plan.Get(sub.Plan).Name, emailT),
				})
				marked = true
			}
		}
		// else: email unavailable/no address — leave unmarked, retry next tick.

		if marked {
			for _, t := range markTs {
				if err := s.repo.MarkReminderSent(ctx, sub.OrgID, sub.ExpiresAt, t); err != nil {
					log.Printf("renewal reminders: mark sent (org %s, t=%d): %v", sub.OrgID, t, err)
				}
			}
		}
	}
}

// StartRenewalReminderScheduler runs runRenewalReminders hourly (and once at
// startup), mirroring StartCleanupScheduler. publicURL is the renew-link base.
func (s *AuthService) StartRenewalReminderScheduler(ctx context.Context, publicURL string) {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.runRenewalReminders(ctx, publicURL)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
	s.runRenewalReminders(ctx, publicURL)
}
