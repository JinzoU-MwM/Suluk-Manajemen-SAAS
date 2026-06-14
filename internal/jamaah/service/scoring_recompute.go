package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// noTouchDays is the days-since-last-touch value used when a lead has no notes
// or follow-ups yet — large enough that ComputeLeadScore awards no freshness
// bonus.
const noTouchDays = 9999

// RecomputeScore recalculates and persists one registration's lead score.
//
// paidAmount/totalAmount are passed when invoice balances are on hand (the lazy
// refresh in ListCRM): the precise payment-progress signal. From the event and
// in-process paths they're 0/0 — totalAmount then falls back to the
// snapshotted contract value and paidAmount stays 0, so stage base carries the
// progression signal until the next balance-backed refresh.
//
// Best-effort by contract: callers log on error and never fail the originating
// mutation because of scoring.
func (s *JamaahService) RecomputeScore(ctx context.Context, orgID, jamaahID, packageID uuid.UUID, paidAmount, totalAmount int64) (int, string, error) {
	base, err := s.repo.GetScoringBase(ctx, orgID, jamaahID, packageID)
	if err != nil {
		return 0, "", err
	}
	docsTotal, docsReceived, err := s.repo.CountDocuments(ctx, orgID, jamaahID)
	if err != nil {
		return 0, "", err
	}
	lastTouch, err := s.repo.LastTouchAt(ctx, orgID, jamaahID)
	if err != nil {
		return 0, "", err
	}

	if totalAmount == 0 {
		if t := base.PriceSnapshot - base.DiscountAmount; t > 0 {
			totalAmount = t
		}
	}

	now := time.Now()
	daysSinceTouch := noTouchDays
	if lastTouch != nil {
		daysSinceTouch = int(now.Sub(*lastTouch).Hours() / 24)
	}
	daysInStage := 0
	if base.StageEnteredAt != nil {
		daysInStage = int(now.Sub(*base.StageEnteredAt).Hours() / 24)
	}

	score, temp := ComputeLeadScore(ScoreSignals{
		Stage:              base.Stage,
		PaidAmount:         paidAmount,
		TotalAmount:        totalAmount,
		DocsTotal:          docsTotal,
		DocsReceived:       docsReceived,
		DaysSinceLastTouch: daysSinceTouch,
		DaysInStage:        daysInStage,
	})
	if err := s.repo.UpdateLeadScore(ctx, orgID, jamaahID, packageID, score, temp); err != nil {
		return 0, "", err
	}
	return score, temp, nil
}

// RecomputeJamaahActive rescoring every still-open registration of one jamaah.
// Used by in-process triggers (note/follow-up/document changes) that know the
// jamaah but not which package(s) to touch.
func (s *JamaahService) RecomputeJamaahActive(ctx context.Context, orgID, jamaahID uuid.UUID) error {
	pkgs, err := s.repo.ListActivePackagesForJamaah(ctx, orgID, jamaahID)
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		if _, _, err := s.RecomputeScore(ctx, orgID, jamaahID, pkg, 0, 0); err != nil {
			return err
		}
	}
	return nil
}

// RecomputeOrgActive rescoring every still-open registration in an org. Invoked
// from the payment.received consumer (which carries no jamaah_id) and the admin
// recompute endpoint. Per-registration failures are logged and skipped so one
// bad row can't abort the sweep.
func (s *JamaahService) RecomputeOrgActive(ctx context.Context, orgID uuid.UUID) error {
	refs, err := s.repo.ListActiveRegistrationIDsForRecompute(ctx, orgID)
	if err != nil {
		return err
	}
	for _, ref := range refs {
		if _, _, err := s.RecomputeScore(ctx, orgID, ref.JamaahID, ref.PackageID, 0, 0); err != nil && s.log != nil {
			s.log.Warnw("recompute lead score failed",
				"org_id", orgID, "jamaah_id", ref.JamaahID, "package_id", ref.PackageID, "err", err)
		}
	}
	return nil
}

// recompute is the fire-and-forget wrapper for in-process triggers: it never
// returns an error to the caller (mutations must not fail on scoring), only
// logs. Runs inline — recompute is a couple of cheap indexed queries.
func (s *JamaahService) recompute(ctx context.Context, orgID, jamaahID, packageID uuid.UUID) {
	if _, _, err := s.RecomputeScore(ctx, orgID, jamaahID, packageID, 0, 0); err != nil && s.log != nil {
		s.log.Warnw("recompute lead score failed",
			"org_id", orgID, "jamaah_id", jamaahID, "package_id", packageID, "err", err)
	}
}

// recomputeJamaah is the fire-and-forget counterpart for jamaah-scoped triggers.
func (s *JamaahService) recomputeJamaah(ctx context.Context, orgID, jamaahID uuid.UUID) {
	if err := s.RecomputeJamaahActive(ctx, orgID, jamaahID); err != nil && s.log != nil {
		s.log.Warnw("recompute lead score (jamaah) failed",
			"org_id", orgID, "jamaah_id", jamaahID, "err", err)
	}
}
