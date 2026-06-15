package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/jamaah/model"
	"github.com/jamaah-in/v2/internal/shared/events"
)

// ErrDepartureGate is returned when a departure transition is blocked by a
// precondition (illegal state change or unmet readiness requirement).
var ErrDepartureGate = errors.New("departure transition not allowed")

// SetDeparture links a group to its package + departure date (draft only).
func (s *JamaahService) SetDeparture(ctx context.Context, groupID, orgID uuid.UUID, req model.SetDepartureRequest) (*model.Group, error) {
	var pkgID *uuid.UUID
	if req.PackageID != "" {
		id, err := uuid.Parse(req.PackageID)
		if err != nil {
			return nil, fmt.Errorf("invalid package_id")
		}
		pkgID = &id
	}
	var depDate *time.Time
	if req.DepartureDate != "" {
		d, err := parseDate(req.DepartureDate)
		if err != nil {
			return nil, fmt.Errorf("departure_date: %w", err)
		}
		depDate = d
	}
	if err := s.repo.SetDeparture(ctx, groupID, orgID, pkgID, depDate); err != nil {
		return nil, err
	}
	return s.repo.GetGroup(ctx, groupID, orgID)
}

// TransitionDeparture advances a kloter through its status workflow with gates:
// going "siap" requires a package, a departure date, and at least one member.
func (s *JamaahService) TransitionDeparture(ctx context.Context, groupID, orgID uuid.UUID, status string) (*model.Group, error) {
	g, err := s.repo.GetGroup(ctx, groupID, orgID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, fmt.Errorf("group not found")
	}
	from := model.DepartureStatus(g.DepartureStatus)
	if from == "" {
		from = model.DepartureDraft
	}
	to := model.DepartureStatus(status)
	if !model.CanTransitionDeparture(from, to) {
		return nil, fmt.Errorf("%w: %s → %s tidak valid", ErrDepartureGate, from, to)
	}

	var manifestFinalizedAt, departedAt *time.Time
	now := time.Now()
	var eventType string

	switch to {
	case model.DepartureSiap:
		if g.PackageID == nil || g.DepartureDate == nil {
			return nil, fmt.Errorf("%w: paket dan tanggal keberangkatan wajib diisi sebelum manifest difinalkan", ErrDepartureGate)
		}
		if g.MemberCount < 1 {
			return nil, fmt.Errorf("%w: kloter belum punya jamaah", ErrDepartureGate)
		}
		manifestFinalizedAt = &now
		eventType = events.EventGroupReady
	case model.DepartureBerangkat:
		departedAt = &now
		eventType = events.EventGroupDeparted
	}

	payload := departureEventPayload(groupID, g.PackageID, status, g.MemberCount)
	if err := s.repo.TransitionDeparture(ctx, groupID, orgID, status, manifestFinalizedAt, departedAt, eventType, payload); err != nil {
		return nil, err
	}
	return s.repo.GetGroup(ctx, groupID, orgID)
}

// GetDepartureManifest returns the kloter + its members for the boarding view.
func (s *JamaahService) GetDepartureManifest(ctx context.Context, groupID, orgID uuid.UUID) (*model.DepartureManifest, error) {
	g, err := s.repo.GetGroup(ctx, groupID, orgID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, fmt.Errorf("group not found")
	}
	members, err := s.repo.ListGroupMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return &model.DepartureManifest{Group: *g, Members: members}, nil
}

func departureEventPayload(groupID uuid.UUID, pkgID *uuid.UUID, status string, memberCount int) []byte {
	m := map[string]any{"group_id": groupID.String(), "status": status, "member_count": memberCount}
	if pkgID != nil {
		m["package_id"] = pkgID.String()
	}
	b, _ := json.Marshal(m)
	return b
}
