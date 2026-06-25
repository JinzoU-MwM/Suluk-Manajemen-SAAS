package service

import (
	"fmt"
	"time"

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
