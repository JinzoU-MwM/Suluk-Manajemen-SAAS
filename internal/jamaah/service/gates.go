package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jamaah-in/v2/internal/jamaah/model"
)

// ErrGate is returned when a pipeline transition is blocked by an unmet
// prerequisite (missing documents, missing mahram). Handlers map it to a 4xx.
var ErrGate = errors.New("transition gate")

// checkTransitionGate enforces document-completeness (and mahram) before a
// jamaah can advance to lunas or berangkat:
//   - lunas: KTP/identitas + paspor
//   - berangkat: paspor + visa, plus a mahram for female jamaah
//
// A requirement counts as met if either a document of that type has been
// received OR the corresponding profile field is filled. It fails open if the
// profile can't be read (never hard-block on an infra hiccup).
func (s *JamaahService) checkTransitionGate(ctx context.Context, orgID, jamaahID uuid.UUID, status string, reg *model.JamaahPackageRegistration) error {
	if status != string(model.StatusLunas) && status != string(model.StatusBerangkat) {
		return nil
	}
	profile, err := s.repo.GetProfileByID(ctx, jamaahID, orgID)
	if err != nil || profile == nil {
		return nil
	}
	docs, _ := s.repo.ListDocuments(ctx, orgID, jamaahID)
	hasDoc := func(t string) bool {
		for _, d := range docs {
			if d.DocType == t && d.Status != "belum_diterima" {
				return true
			}
		}
		return false
	}
	filled := func(p *string) bool { return p != nil && strings.TrimSpace(*p) != "" }

	hasKTP := hasDoc("ktp") || filled(profile.NoIdentitas)
	hasPaspor := hasDoc("paspor") || filled(profile.NoPaspor)
	hasVisa := hasDoc("visa") || strings.TrimSpace(profile.NoVisa) != ""

	var missing []string
	switch status {
	case string(model.StatusLunas):
		if !hasKTP {
			missing = append(missing, "KTP/identitas")
		}
		if !hasPaspor {
			missing = append(missing, "paspor")
		}
	case string(model.StatusBerangkat):
		if !hasPaspor {
			missing = append(missing, "paspor")
		}
		if !hasVisa {
			missing = append(missing, "visa")
		}
		if strings.EqualFold(strings.TrimSpace(profile.Gender), "P") && reg.MahramID == nil {
			missing = append(missing, "mahram (wajib untuk jamaah perempuan)")
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: lengkapi dulu sebelum status \"%s\": %s", ErrGate, status, strings.Join(missing, ", "))
	}
	return nil
}

// SetMahram links (or clears, when mahramID is nil) the mahram of a jamaah's
// registration.
func (s *JamaahService) SetMahram(ctx context.Context, orgID, jamaahID, packageID uuid.UUID, mahramID *uuid.UUID) (*model.JamaahPackageRegistration, error) {
	if _, err := s.repo.GetRegistration(ctx, orgID, jamaahID, packageID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateRegistrationMahram(ctx, orgID, jamaahID, packageID, mahramID); err != nil {
		return nil, err
	}
	return s.repo.GetRegistration(ctx, orgID, jamaahID, packageID)
}
