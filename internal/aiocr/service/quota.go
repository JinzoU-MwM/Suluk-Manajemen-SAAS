package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/shared/plan"
)

// ErrScanQuotaExceeded is returned when an org has used its full monthly scan
// quota (base plan tier + purchased top-ups). Handlers map it to 402, not 500.
var ErrScanQuotaExceeded = errors.New("kuota scan bulanan sudah habis")

type quotaCacheEntry struct {
	used, limit int
	expiry      time.Time
}

// fetchQuota asks auth-service for the org's monthly scan usage/limit —
// auth-service is the single source of truth for plan tiers (see
// internal/shared/plan) and already folds in purchased top-ups and the
// expired/cancelled-subscription downgrade to Gratis limits. It FAILS OPEN
// (limit=plan.Unlimited) when authAddr/authToken is unavailable or the call
// errors, mirroring jamaah-service's fetchLimits: a transient auth-service
// outage must not take down document scanning entirely.
func (s *AIOCRService) fetchQuota(ctx context.Context, orgID uuid.UUID, authToken string) (used, limit int) {
	if s.authAddr == "" || authToken == "" {
		return 0, plan.Unlimited
	}
	if v, ok := s.quotaCache.Load(orgID); ok {
		if e := v.(quotaCacheEntry); e.expiry.After(time.Now()) {
			return e.used, e.limit
		}
	}
	var out struct {
		UsageCount int `json:"usage_count"`
		UsageLimit int `json:"usage_limit"`
	}
	if err := s.httpc.GetJSON(ctx, s.authAddr, "/api/v1/subscription/status", authToken, &out); err != nil {
		return 0, plan.Unlimited
	}
	s.quotaCache.Store(orgID, quotaCacheEntry{used: out.UsageCount, limit: out.UsageLimit, expiry: time.Now().Add(45 * time.Second)})
	return out.UsageCount, out.UsageLimit
}
