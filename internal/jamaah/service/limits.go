package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jamaah-in/v2/internal/shared/plan"
)

// ErrPlanLimit is returned when a create would exceed the org's plan quota.
// Handlers map it to a 4xx (not a 500).
var ErrPlanLimit = errors.New("plan limit reached")

// planLimits is the subset of the auth-service subscription-status response we
// need to enforce per-tier caps on jamaah and groups.
type planLimits struct {
	Plan      string `json:"plan"`
	MaxJamaah int    `json:"max_jamaah"`
	MaxGroups int    `json:"max_groups"`
}

type limitCacheEntry struct {
	limits planLimits
	expiry time.Time
}

// fetchLimits asks auth-service for the caller org's plan limits. It FAILS OPEN
// (returns unlimited) when the token/address is missing or the call fails, so a
// transient auth-service outage never blocks core CRUD — the cap is best-effort
// enforcement layered on top of the (authoritative) frontend gating.
//
// Results are cached per org for a short TTL: plan limits change rarely
// (upgrade/expiry), so this removes the auth round-trip from every create — and
// avoids N round-trips during a bulk import.
func (s *JamaahService) fetchLimits(ctx context.Context, orgID uuid.UUID, authToken string) planLimits {
	unlimited := planLimits{MaxJamaah: plan.Unlimited, MaxGroups: plan.Unlimited}
	if s.authAddr == "" || authToken == "" {
		return unlimited
	}
	if v, ok := s.limitCache.Load(orgID); ok {
		if e := v.(limitCacheEntry); e.expiry.After(time.Now()) {
			return e.limits
		}
	}
	var out planLimits
	if err := s.httpc.GetJSON(ctx, s.authAddr, "/api/v1/subscription/status", authToken, &out); err != nil {
		return unlimited
	}
	s.limitCache.Store(orgID, limitCacheEntry{limits: out, expiry: time.Now().Add(45 * time.Second)})
	return out
}

// reserveSeat asks package-service to reserve one seat (capacity-checked). It
// returns a user-facing error when the package is full/unavailable so the
// registration is aborted. If package-service isn't configured it no-ops (dev).
func (s *JamaahService) reserveSeat(ctx context.Context, packageID uuid.UUID, roomType, authToken string) error {
	if s.packageAddr == "" || authToken == "" {
		return nil
	}
	path := "/api/v1/packages/" + packageID.String() + "/reserve"
	headers := map[string]string{"Authorization": authToken}
	if err := s.httpc.PostJSON(ctx, s.packageAddr, path, headers, map[string]string{"room_type": roomType}, nil); err != nil {
		return fmt.Errorf("%w: kuota paket/tipe kamar sudah penuh atau paket tidak tersedia", ErrPlanLimit)
	}
	return nil
}

// releaseSeat frees one reserved seat + room-type seat (best-effort; ignored).
func (s *JamaahService) releaseSeat(ctx context.Context, packageID uuid.UUID, roomType, authToken string) {
	if s.packageAddr == "" || authToken == "" {
		return
	}
	path := "/api/v1/packages/" + packageID.String() + "/release"
	headers := map[string]string{"Authorization": authToken}
	_ = s.httpc.PostJSON(ctx, s.packageAddr, path, headers, map[string]string{"room_type": roomType}, nil)
}
