package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/jamaah-in/v2/internal/aiocr/model"
	"github.com/jamaah-in/v2/internal/aiocr/repository"
)

type AIOCRService struct {
	repo   *repository.AIOCRRepo
	gemini *GeminiClient
	logger *zap.SugaredLogger
}

func NewAIOCRService(repo *repository.AIOCRRepo, gemini *GeminiClient, logger *zap.SugaredLogger) *AIOCRService {
	return &AIOCRService{
		repo:   repo,
		gemini: gemini,
		logger: logger,
	}
}

// Available reports whether OCR is configured (Gemini API key present).
func (s *AIOCRService) Available() bool {
	return s.gemini != nil
}

func (s *AIOCRService) CreateScanJob(ctx context.Context, orgID, userID uuid.UUID, req model.CreateScanJobRequest) (*model.ScanJob, error) {
	var packageID *uuid.UUID
	if req.PackageID != nil && *req.PackageID != "" {
		pid, err := uuid.Parse(*req.PackageID)
		if err != nil {
			return nil, err
		}
		packageID = &pid
	}

	job := &model.ScanJob{
		ID:             uuid.New(),
		OrgID:          orgID,
		UserID:         userID,
		PackageID:      packageID,
		Status:         string(model.ScanStatusPending),
		TotalFiles:     len(req.Files),
		ProcessedFiles: 0,
	}

	if err := s.repo.CreateScanJob(ctx, job); err != nil {
		return nil, err
	}

	for _, f := range req.Files {
		fileHash := f.FileHash
		if fileHash == "" {
			fileHash = hashString(f.FileURL + f.FileName)
		}

		sr := &model.ScanResult{
			ID:            uuid.New(),
			ScanJobID:     job.ID,
			OrgID:         orgID,
			FileName:      f.FileName,
			FileURL:       f.FileURL,
			FileSize:      nil,
			FileHash:      fileHash,
			Status:        string(model.ResultStatusPending),
			ModelUsed:     "gemini-2.0-flash",
			PromptVersion: "v1",
		}
		if f.FileSize > 0 {
			fs := f.FileSize
			sr.FileSize = &fs
		}

		if err := s.repo.CreateScanResult(ctx, sr); err != nil {
			continue
		}
	}

	return s.repo.GetScanJob(ctx, job.ID, orgID)
}

func (s *AIOCRService) GetScanJob(ctx context.Context, id, orgID uuid.UUID) (*model.ScanJob, error) {
	job, err := s.repo.GetScanJob(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	results, _ := s.repo.GetScanResultsByJob(ctx, orgID, id)
	job.Results = results
	return job, nil
}

func (s *AIOCRService) ListScanJobs(ctx context.Context, orgID uuid.UUID, status string, page, limit int) ([]model.ScanJob, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.repo.ListScanJobs(ctx, orgID, status, offset, limit)
}

func (s *AIOCRService) GetScanResult(ctx context.Context, id, orgID uuid.UUID) (*model.ScanResult, error) {
	return s.repo.GetScanResult(ctx, id, orgID)
}

func (s *AIOCRService) GetScanResultsByJob(ctx context.Context, orgID, jobID uuid.UUID) ([]model.ScanResult, error) {
	return s.repo.GetScanResultsByJob(ctx, orgID, jobID)
}

func (s *AIOCRService) NormalizeToSiskopatuh(ctx context.Context, orgID uuid.UUID, req model.NormalizeRequest) (any, error) {
	return normalizeToSiskopatuh(req.Data, req.DocType), nil
}

func (s *AIOCRService) ExportSiskopatuhExcel(ctx context.Context, orgID uuid.UUID, req model.ExportSiskopatuhRequest) ([]byte, error) {
	var packageID *uuid.UUID
	if req.PackageID != nil && *req.PackageID != "" {
		pid, err := uuid.Parse(*req.PackageID)
		if err != nil {
			return nil, fmt.Errorf("invalid package_id: %w", err)
		}
		packageID = &pid
	}

	results, err := s.repo.GetCompletedScanResults(ctx, orgID, packageID)
	if err != nil {
		return nil, fmt.Errorf("fetch scan results: %w", err)
	}

	return generateSiskopatuhExcel(results)
}

func (s *AIOCRService) CreateExportTemplate(ctx context.Context, orgID, userID uuid.UUID, req model.CreateExportTemplateRequest) (*model.ExportTemplate, error) {
	t := &model.ExportTemplate{
		ID:            uuid.New(),
		OrgID:         orgID,
		UserID:        userID,
		Name:          req.Name,
		FileURL:       req.FileURL,
		ColumnMapping: req.ColumnMapping,
		HeaderRow:     req.HeaderRow,
		DataStartRow:  req.DataStartRow,
		IsDefault:     req.IsDefault,
	}
	if t.HeaderRow == 0 {
		t.HeaderRow = 1
	}
	if t.DataStartRow == 0 {
		t.DataStartRow = 2
	}

	if err := s.repo.CreateExportTemplate(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *AIOCRService) ListExportTemplates(ctx context.Context, orgID uuid.UUID) ([]model.ExportTemplate, error) {
	return s.repo.ListExportTemplates(ctx, orgID)
}

func (s *AIOCRService) DeleteExportTemplate(ctx context.Context, id, orgID uuid.UUID) error {
	return s.repo.DeleteExportTemplate(ctx, id, orgID)
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
