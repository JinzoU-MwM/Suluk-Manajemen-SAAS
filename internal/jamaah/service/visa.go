package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/jamaah/model"
	"github.com/jamaah-in/v2/internal/jamaah/repository"
	"github.com/jamaah-in/v2/internal/shared/events"
)

// ErrVisaGate is returned when a visa transition is blocked by a precondition
// (illegal state change or a missing prerequisite document).
var ErrVisaGate = errors.New("visa transition not allowed")

// GetVisa returns a jamaah's visa application, or ErrVisaNotFound.
func (s *JamaahService) GetVisa(ctx context.Context, orgID, jamaahID uuid.UUID) (*model.VisaApplication, error) {
	return s.repo.GetVisaByJamaah(ctx, orgID, jamaahID)
}

// GetVisaHistory returns the audit trail for a jamaah's visa.
func (s *JamaahService) GetVisaHistory(ctx context.Context, orgID, jamaahID uuid.UUID) ([]model.VisaHistory, error) {
	v, err := s.repo.GetVisaByJamaah(ctx, orgID, jamaahID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListVisaHistory(ctx, orgID, v.ID)
}

// UpsertVisa creates or edits the draft visa application's fields.
func (s *JamaahService) UpsertVisa(ctx context.Context, orgID, jamaahID uuid.UUID, req model.UpsertVisaRequest) (*model.VisaApplication, error) {
	v := &model.VisaApplication{OrgID: orgID, JamaahID: jamaahID, Provider: req.Provider, ReferenceNo: req.ReferenceNo, Notes: req.Notes}
	if req.PackageID != "" {
		pid, err := uuid.Parse(req.PackageID)
		if err != nil {
			return nil, fmt.Errorf("invalid package_id")
		}
		v.PackageID = &pid
	}
	if req.ExpiryDate != "" {
		d, err := parseDate(req.ExpiryDate)
		if err != nil {
			return nil, fmt.Errorf("expiry_date: %w", err)
		}
		v.ExpiryDate = d
	}
	return s.repo.UpsertVisaDraft(ctx, v)
}

func (s *JamaahService) ListVisas(ctx context.Context, orgID uuid.UUID, status, search string, page, limit int) ([]model.VisaApplication, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	return s.repo.ListVisas(ctx, orgID, status, search, (page-1)*limit, limit)
}

// TransitionVisa applies a validated status change with its gates. submit
// requires the passport document to be received; approve records the reference
// number + expiry and mirrors them onto the profile.
func (s *JamaahService) TransitionVisa(ctx context.Context, orgID, userID, jamaahID uuid.UUID, req model.VisaTransitionRequest) (*model.VisaApplication, error) {
	cur, err := s.repo.GetVisaByJamaah(ctx, orgID, jamaahID)
	if err != nil {
		return nil, err
	}
	from := model.VisaStatus(cur.Status)
	to := model.VisaStatus(req.Status)
	if !model.CanTransitionVisa(from, to) {
		return nil, fmt.Errorf("%w: %s → %s tidak valid", ErrVisaGate, from, to)
	}

	t := repository.VisaTransition{
		VisaID: cur.ID, OrgID: orgID, JamaahID: jamaahID,
		FromStatus: cur.Status, ToStatus: req.Status,
		Reason: req.Reason, ChangedBy: &userID,
	}
	now := time.Now()

	switch to {
	case model.VisaSubmitted:
		ok, derr := s.repo.HasReceivedDocument(ctx, orgID, jamaahID, "paspor")
		if derr != nil {
			return nil, derr
		}
		if !ok {
			return nil, fmt.Errorf("%w: dokumen paspor harus diterima sebelum visa diajukan", ErrVisaGate)
		}
		t.SubmittedAt = &now
		t.EventType = events.EventVisaSubmitted
	case model.VisaApproved:
		t.DecidedAt = &now
		t.ReferenceNo = req.ReferenceNo
		if req.ExpiryDate != "" {
			d, derr := parseDate(req.ExpiryDate)
			if derr != nil {
				return nil, fmt.Errorf("expiry_date: %w", derr)
			}
			t.ExpiryDate = d
		}
		t.EventType = events.EventVisaApproved
	case model.VisaRejected:
		t.DecidedAt = &now
		t.RejectReason = req.Reason
		t.EventType = events.EventVisaRejected
	case model.VisaExpired:
		t.EventType = events.EventVisaExpired
	}

	t.Payload = visaEventPayload(cur.JamaahID, req.Status, t.ReferenceNo)
	if err := s.repo.TransitionVisaTx(ctx, t); err != nil {
		return nil, err
	}

	// On approval, mirror the visa onto the profile (best-effort).
	if to == model.VisaApproved {
		if err := s.repo.UpdateProfileVisaFields(ctx, orgID, jamaahID, t.ReferenceNo, &now, t.ExpiryDate); err != nil && s.log != nil {
			s.log.Warnw("sync profile visa fields failed", "jamaah_id", jamaahID, "err", err)
		}
	}
	return s.repo.GetVisaByJamaah(ctx, orgID, jamaahID)
}

func visaEventPayload(jamaahID uuid.UUID, status, ref string) []byte {
	b, _ := json.Marshal(map[string]any{"jamaah_id": jamaahID.String(), "status": status, "reference_no": ref})
	return b
}

// reminderThresholds are the day-marks at which an expiry reminder fires (once
// each, deduped via lifecycle_reminders).
var reminderThresholds = []int{30, 14, 7}

// StartLifecycleReminders runs a daily scan that notifies on passport/visa
// expiry and auto-expires lapsed visas. Best-effort; logs and continues.
func (s *JamaahService) StartLifecycleReminders(ctx context.Context, interval time.Duration) {
	go func() {
		// Small initial delay so startup isn't noisy, then tick.
		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				s.runLifecycleScan(ctx)
				timer.Reset(interval)
			}
		}
	}()
}

func (s *JamaahService) runLifecycleScan(ctx context.Context) {
	// 1) auto-expire lapsed approved visas.
	if expired, err := s.repo.ScanVisaExpiredAll(ctx); err == nil {
		for _, v := range expired {
			t := repository.VisaTransition{
				VisaID: v.ID, OrgID: v.OrgID, JamaahID: v.JamaahID,
				FromStatus: v.Status, ToStatus: string(model.VisaExpired),
				Reason: "auto: visa kedaluwarsa", EventType: events.EventVisaExpired,
				Payload: visaEventPayload(v.JamaahID, string(model.VisaExpired), v.ReferenceNo),
			}
			if err := s.repo.TransitionVisaTx(ctx, t); err != nil && s.log != nil {
				s.log.Warnw("auto-expire visa failed", "visa_id", v.ID, "err", err)
			}
		}
	} else if s.log != nil {
		s.log.Warnw("scan expired visas failed", "err", err)
	}

	// 2) threshold reminders for visa + passport (each subject+milestone once).
	maxDays := reminderThresholds[0]
	s.remind(ctx, "visa", func() ([]repository.ExpiringSubject, error) { return s.repo.ScanVisaExpiringAll(ctx, maxDays) }, "Visa akan kedaluwarsa")
	s.remind(ctx, "passport", func() ([]repository.ExpiringSubject, error) { return s.repo.ScanPassportExpiringAll(ctx, maxDays) }, "Paspor akan kedaluwarsa")
}

func (s *JamaahService) remind(ctx context.Context, subjectType string, scan func() ([]repository.ExpiringSubject, error), title string) {
	subjects, err := scan()
	if err != nil {
		if s.log != nil {
			s.log.Warnw("expiry scan failed", "type", subjectType, "err", err)
		}
		return
	}
	for _, sub := range subjects {
		th := thresholdBucket(sub.DaysLeft)
		if th == 0 {
			continue
		}
		milestone := fmt.Sprintf("%s_%d", subjectType, th)
		fresh, err := s.repo.TryRecordReminder(ctx, sub.OrgID, sub.JamaahID, subjectType, milestone)
		if err != nil || !fresh {
			continue
		}
		if s.notifier != nil {
			msg := fmt.Sprintf("%s a.n. %s dalam %d hari.", title, sub.Name, sub.DaysLeft)
			s.notifier.Send(ctx, sub.OrgID.String(), "", "warning", title, msg)
		}
	}
}

// thresholdBucket maps a days-left value to the tightest threshold it has
// crossed (so a 5-days-left item fires the "7" milestone, not "30").
func thresholdBucket(daysLeft int) int {
	for i := len(reminderThresholds) - 1; i >= 0; i-- {
		if daysLeft <= reminderThresholds[i] {
			return reminderThresholds[i]
		}
	}
	return 0
}
