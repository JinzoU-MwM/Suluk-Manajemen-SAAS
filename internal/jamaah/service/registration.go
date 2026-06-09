package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/jamaah/model"
)

func (s *JamaahService) GenerateRegistrationLink(ctx context.Context, orgID, userID string, req model.GenerateLinkRequest) (*model.RegistrationLink, error) {
	tokenBytes := make([]byte, 16)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	expiresIn := req.ExpiresInDays
	if expiresIn <= 0 {
		expiresIn = 30
	}

	link := &model.RegistrationLink{
		OrgID:     orgID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(expiresIn) * 24 * time.Hour),
		IsActive:  true,
		CreatedBy: &userID,
	}
	if req.GroupID != "" {
		link.GroupID = &req.GroupID
	}
	if req.PackageID != "" {
		link.PackageID = &req.PackageID
	}

	if err := s.repo.CreateRegistrationLink(ctx, link); err != nil {
		return nil, err
	}
	return link, nil
}

func (s *JamaahService) GetActiveLink(ctx context.Context, orgID, groupID string) (*model.RegistrationLink, error) {
	return s.repo.GetActiveRegistrationLink(ctx, orgID, groupID)
}

func (s *JamaahService) RevokeLink(ctx context.Context, orgID, groupID string) error {
	return s.repo.DeactivateRegistrationLink(ctx, orgID, groupID)
}

func (s *JamaahService) GetPublicRegistrationInfo(ctx context.Context, token string) (*model.PublicRegistrationInfo, error) {
	link, err := s.repo.GetRegistrationLinkByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	info := &model.PublicRegistrationInfo{
		ExpiresAt: link.ExpiresAt,
		IsExpired: time.Now().After(link.ExpiresAt),
	}
	if info.IsExpired {
		return nil, fmt.Errorf("registration link has expired")
	}
	return info, nil
}

func (s *JamaahService) SubmitPublicRegistration(ctx context.Context, token string, req model.PublicRegistrationSubmit, filePaths map[string]string) (*model.PendingRegistration, error) {
	link, err := s.repo.GetRegistrationLinkByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if time.Now().After(link.ExpiresAt) {
		return nil, fmt.Errorf("registration link has expired")
	}

	p := &model.PendingRegistration{
		OrgID:              link.OrgID,
		RegistrationLinkID: &link.ID,
		PhoneNumber:        req.PhoneNumber,
		Name:               req.Name,
		Email:              req.Email,
		KtpFileURL:         filePaths["ktp"],
		PassportFileURL:    filePaths["passport"],
		VisaFileURL:        filePaths["visa"],
		Status:             "pending",
	}
	if err := s.repo.CreatePendingRegistration(ctx, p); err != nil {
		return nil, err
	}
	if s.notifier != nil {
		name := req.Name
		if name == "" {
			name = req.PhoneNumber
		}
		s.notifier.Send(ctx, link.OrgID, "", "info", "Pendaftaran baru",
			fmt.Sprintf("Calon jamaah baru mendaftar: %s", name))
	}
	return p, nil
}

func (s *JamaahService) ListPending(ctx context.Context, orgID, groupID string) ([]model.PendingRegistration, error) {
	return s.repo.ListPendingRegistrations(ctx, orgID, groupID)
}

func (s *JamaahService) ApprovePending(ctx context.Context, orgID string, pendingID, reviewerID uuid.UUID) error {
	pr, err := s.repo.GetPendingRegistration(ctx, pendingID.String(), orgID)
	if err != nil {
		return err
	}
	if pr.Status != "pending" {
		return fmt.Errorf("registration already %s", pr.Status)
	}

	profile := &model.JamaahProfile{
		OrgID:       uuid.MustParse(orgID),
		Nama:        pr.Name,
		NoHP:        pr.PhoneNumber,
		Email:       pr.Email,
		NoIdentitas: &pr.KtpFileURL,
		NoPaspor:    &pr.PassportFileURL,
	}

	if err := s.repo.CreateProfile(ctx, profile); err != nil {
		return err
	}

	return s.repo.ApprovePendingRegistration(ctx, uuid.MustParse(pendingID.String()), uuid.MustParse(orgID), reviewerID, profile.ID)
}

func (s *JamaahService) RejectPending(ctx context.Context, orgID string, pendingID, reviewerID uuid.UUID) error {
	return s.repo.RejectPendingRegistration(ctx, uuid.MustParse(pendingID.String()), uuid.MustParse(orgID), reviewerID)
}
