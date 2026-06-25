package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jamaah-in/v2/internal/aiocr/model"
)

type AIOCRRepo struct {
	pool *pgxpool.Pool
}

func NewAIOCRRepo(pool *pgxpool.Pool) *AIOCRRepo {
	return &AIOCRRepo{pool: pool}
}

// IncrementScanUsage atomically adds n successful scans to the org's counter for
// the current calendar month (upsert on the (org, year, month) key). The
// increment runs in the DB, so it is safe under the concurrent OCR goroutines.
func (r *AIOCRRepo) IncrementScanUsage(ctx context.Context, orgID uuid.UUID, n int) error {
	if n <= 0 {
		return nil
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO scan_usage (org_id, year, month, documents_scanned)
		VALUES ($1, EXTRACT(YEAR FROM NOW())::int, EXTRACT(MONTH FROM NOW())::int, $2)
		ON CONFLICT (org_id, year, month)
		DO UPDATE SET documents_scanned = scan_usage.documents_scanned + EXCLUDED.documents_scanned,
		              updated_at = NOW()`,
		orgID, n)
	return err
}

// GetScanUsageThisMonth returns the org's scanned-document count for the current
// calendar month (0 when no row exists yet).
func (r *AIOCRRepo) GetScanUsageThisMonth(ctx context.Context, orgID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE((SELECT documents_scanned FROM scan_usage
			WHERE org_id = $1 AND year = EXTRACT(YEAR FROM NOW())::int
			  AND month = EXTRACT(MONTH FROM NOW())::int), 0)`,
		orgID).Scan(&n)
	return n, err
}

// CreditScanTopup records a purchased top-up for the org's current month. The
// order_id PK + DO NOTHING make it idempotent: a duplicate webhook credits once.
func (r *AIOCRRepo) CreditScanTopup(ctx context.Context, orderID, orgID uuid.UUID, scans int) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO scan_topups (order_id, org_id, year, month, scans)
		VALUES ($1, $2, EXTRACT(YEAR FROM NOW())::int, EXTRACT(MONTH FROM NOW())::int, $3)
		ON CONFLICT (order_id) DO NOTHING`,
		orderID, orgID, scans)
	return err
}

// GetPurchasedScansThisMonth sums the org's top-up credits for the current month
// (0 when none).
func (r *AIOCRRepo) GetPurchasedScansThisMonth(ctx context.Context, orgID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(scans), 0) FROM scan_topups
		WHERE org_id = $1 AND year = EXTRACT(YEAR FROM NOW())::int
		  AND month = EXTRACT(MONTH FROM NOW())::int`,
		orgID).Scan(&n)
	return n, err
}

var (
	ErrScanJobNotFound    = fmt.Errorf("scan job not found")
	ErrScanResultNotFound = fmt.Errorf("scan result not found")
	ErrTemplateNotFound   = fmt.Errorf("export template not found")
)

func (r *AIOCRRepo) CreateScanJob(ctx context.Context, job *model.ScanJob) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO scan_jobs (id, org_id, user_id, package_id, status, total_files, processed_files)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING created_at, updated_at`,
		job.ID, job.OrgID, job.UserID, job.PackageID, job.Status, job.TotalFiles, job.ProcessedFiles,
	).Scan(&job.CreatedAt, &job.UpdatedAt)
}

func (r *AIOCRRepo) GetScanJob(ctx context.Context, id, orgID uuid.UUID) (*model.ScanJob, error) {
	job := &model.ScanJob{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, org_id, user_id, package_id, status, total_files, processed_files, created_at, updated_at
		FROM scan_jobs WHERE id = $1 AND org_id = $2`, id, orgID).Scan(
		&job.ID, &job.OrgID, &job.UserID, &job.PackageID, &job.Status,
		&job.TotalFiles, &job.ProcessedFiles, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return nil, ErrScanJobNotFound
	}
	return job, nil
}

func (r *AIOCRRepo) UpdateScanJobStatus(ctx context.Context, id, orgID uuid.UUID, status string, processed int) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE scan_jobs SET status = $3, processed_files = $4, updated_at = NOW() WHERE id = $1 AND org_id = $2`,
		id, orgID, status, processed)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrScanJobNotFound
	}
	return nil
}

func (r *AIOCRRepo) ListScanJobs(ctx context.Context, orgID uuid.UUID, status string, offset, limit int) ([]model.ScanJob, int, error) {
	countQuery := `SELECT COUNT(*) FROM scan_jobs WHERE org_id = $1`
	query := `SELECT id, org_id, user_id, package_id, status, total_files, processed_files, created_at, updated_at
		FROM scan_jobs WHERE org_id = $1`
	args := []any{orgID}
	argIdx := 2

	if status != "" {
		countQuery += fmt.Sprintf(` AND status = $%d`, argIdx)
		query += fmt.Sprintf(` AND status = $%d`, argIdx)
		args = append(args, status)
		argIdx++
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	jobs := []model.ScanJob{}
	for rows.Next() {
		var j model.ScanJob
		if err := rows.Scan(&j.ID, &j.OrgID, &j.UserID, &j.PackageID, &j.Status,
			&j.TotalFiles, &j.ProcessedFiles, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, 0, err
		}
		jobs = append(jobs, j)
	}
	return jobs, total, nil
}

func (r *AIOCRRepo) CreateScanResult(ctx context.Context, sr *model.ScanResult) error {
	extractedData, _ := json.Marshal(sr.ExtractedData)
	normalizedData, _ := json.Marshal(sr.NormalizedData)
	validationErrors, _ := json.Marshal(sr.ValidationErrors)

	return r.pool.QueryRow(ctx,
		`INSERT INTO scan_results (id, scan_job_id, org_id, file_name, file_url, file_size, file_hash,
			doc_type, extracted_data, normalized_data, validation_errors, cache_hit, model_used,
			prompt_version, status, error_message, processing_time_ms)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17) RETURNING created_at, updated_at`,
		sr.ID, sr.ScanJobID, sr.OrgID, sr.FileName, sr.FileURL, sr.FileSize, sr.FileHash,
		sr.DocType, extractedData, normalizedData, validationErrors, sr.CacheHit, sr.ModelUsed,
		sr.PromptVersion, sr.Status, sr.ErrorMessage, sr.ProcessingTimeMs,
	).Scan(&sr.CreatedAt, &sr.UpdatedAt)
}

func (r *AIOCRRepo) GetScanResult(ctx context.Context, id, orgID uuid.UUID) (*model.ScanResult, error) {
	sr := &model.ScanResult{}
	var extractedData, normalizedData, validationErrors []byte

	err := r.pool.QueryRow(ctx,
		`SELECT id, scan_job_id, org_id, file_name, file_url, file_size, file_hash, doc_type,
			extracted_data, normalized_data, validation_errors, cache_hit, model_used, prompt_version,
			status, error_message, processing_time_ms, created_at, updated_at
		FROM scan_results WHERE id = $1 AND org_id = $2`, id, orgID).Scan(
		&sr.ID, &sr.ScanJobID, &sr.OrgID, &sr.FileName, &sr.FileURL, &sr.FileSize, &sr.FileHash,
		&sr.DocType, &extractedData, &normalizedData, &validationErrors, &sr.CacheHit, &sr.ModelUsed,
		&sr.PromptVersion, &sr.Status, &sr.ErrorMessage, &sr.ProcessingTimeMs, &sr.CreatedAt, &sr.UpdatedAt)
	if err != nil {
		return nil, ErrScanResultNotFound
	}

	if extractedData != nil {
		_ = json.Unmarshal(extractedData, &sr.ExtractedData)
	}
	if normalizedData != nil {
		_ = json.Unmarshal(normalizedData, &sr.NormalizedData)
	}
	if validationErrors != nil {
		_ = json.Unmarshal(validationErrors, &sr.ValidationErrors)
	}

	return sr, nil
}

func (r *AIOCRRepo) GetScanResultsByJob(ctx context.Context, orgID, jobID uuid.UUID) ([]model.ScanResult, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, scan_job_id, org_id, file_name, file_url, file_size, file_hash, doc_type,
			extracted_data, normalized_data, validation_errors, cache_hit, model_used, prompt_version,
			status, error_message, processing_time_ms, created_at, updated_at
		FROM scan_results WHERE org_id = $1 AND scan_job_id = $2 ORDER BY created_at`, orgID, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []model.ScanResult{}
	for rows.Next() {
		var sr model.ScanResult
		var extractedData, normalizedData, validationErrors []byte
		if err := rows.Scan(
			&sr.ID, &sr.ScanJobID, &sr.OrgID, &sr.FileName, &sr.FileURL, &sr.FileSize, &sr.FileHash,
			&sr.DocType, &extractedData, &normalizedData, &validationErrors, &sr.CacheHit, &sr.ModelUsed,
			&sr.PromptVersion, &sr.Status, &sr.ErrorMessage, &sr.ProcessingTimeMs, &sr.CreatedAt, &sr.UpdatedAt); err != nil {
			return nil, err
		}
		if extractedData != nil {
			_ = json.Unmarshal(extractedData, &sr.ExtractedData)
		}
		if normalizedData != nil {
			_ = json.Unmarshal(normalizedData, &sr.NormalizedData)
		}
		if validationErrors != nil {
			_ = json.Unmarshal(validationErrors, &sr.ValidationErrors)
		}
		results = append(results, sr)
	}
	return results, nil
}

func (r *AIOCRRepo) UpdateScanResultStatus(ctx context.Context, id, orgID uuid.UUID, status, errorMessage string) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE scan_results SET status = $3, error_message = $4, updated_at = NOW() WHERE id = $1 AND org_id = $2`,
		id, orgID, status, errorMessage)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrScanResultNotFound
	}
	return nil
}

func (r *AIOCRRepo) CreateExportTemplate(ctx context.Context, t *model.ExportTemplate) error {
	columnMapping, _ := json.Marshal(t.ColumnMapping)
	return r.pool.QueryRow(ctx,
		`INSERT INTO export_templates (id, org_id, user_id, name, file_url, column_mapping, header_row, data_start_row, is_default)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING created_at`,
		t.ID, t.OrgID, t.UserID, t.Name, t.FileURL, columnMapping, t.HeaderRow, t.DataStartRow, t.IsDefault,
	).Scan(&t.CreatedAt)
}

func (r *AIOCRRepo) ListExportTemplates(ctx context.Context, orgID uuid.UUID) ([]model.ExportTemplate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, org_id, user_id, name, file_url, column_mapping, header_row, data_start_row, is_default, created_at
		FROM export_templates WHERE org_id = $1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := []model.ExportTemplate{}
	for rows.Next() {
		var t model.ExportTemplate
		var columnMapping []byte
		if err := rows.Scan(&t.ID, &t.OrgID, &t.UserID, &t.Name, &t.FileURL, &columnMapping,
			&t.HeaderRow, &t.DataStartRow, &t.IsDefault, &t.CreatedAt); err != nil {
			return nil, err
		}
		if columnMapping != nil {
			_ = json.Unmarshal(columnMapping, &t.ColumnMapping)
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (r *AIOCRRepo) DeleteExportTemplate(ctx context.Context, id, orgID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM export_templates WHERE id = $1 AND org_id = $2`, id, orgID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrTemplateNotFound
	}
	return nil
}

func (r *AIOCRRepo) UpdateScanResultFailed(ctx context.Context, id, orgID uuid.UUID, errorMessage string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE scan_results SET status = 'failed', error_message = $3, updated_at = NOW() WHERE id = $1 AND org_id = $2`,
		id, orgID, errorMessage)
	return err
}

func (r *AIOCRRepo) UpdateScanResultCompleted(ctx context.Context, id, orgID uuid.UUID, docType *string,
	extractedData, normalizedData, validationErrors any, cacheHit *bool, processingTimeMs int) error {
	extractedJSON, _ := json.Marshal(extractedData)
	normalizedJSON, _ := json.Marshal(normalizedData)
	valErrJSON, _ := json.Marshal(validationErrors)

	_, err := r.pool.Exec(ctx,
		`UPDATE scan_results SET
			status = 'completed', doc_type = $3, extracted_data = $4,
			normalized_data = $5, validation_errors = $6, cache_hit = COALESCE($7, cache_hit),
			processing_time_ms = $8, updated_at = NOW()
		WHERE id = $1 AND org_id = $2`,
		id, orgID, docType, extractedJSON, normalizedJSON, valErrJSON, cacheHit, processingTimeMs)
	return err
}

func (r *AIOCRRepo) GetCompletedScanResults(ctx context.Context, orgID uuid.UUID, packageID *uuid.UUID) ([]model.ScanResult, error) {
	query := `SELECT sr.id, sr.scan_job_id, sr.org_id, sr.file_name, sr.file_url, sr.file_size, sr.file_hash,
			sr.doc_type, sr.extracted_data, sr.normalized_data, sr.validation_errors, sr.cache_hit,
			sr.model_used, sr.prompt_version, sr.status, sr.error_message, sr.processing_time_ms,
			sr.created_at, sr.updated_at
		FROM scan_results sr
		JOIN scan_jobs sj ON sj.id = sr.scan_job_id
		WHERE sr.org_id = $1 AND sr.status = 'completed'`
	args := []any{orgID}

	if packageID != nil {
		query += ` AND sj.package_id = $2`
		args = append(args, *packageID)
	}

	query += ` ORDER BY sr.created_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []model.ScanResult{}
	for rows.Next() {
		var sr model.ScanResult
		var extractedData, normalizedData, validationErrors []byte
		if err := rows.Scan(
			&sr.ID, &sr.ScanJobID, &sr.OrgID, &sr.FileName, &sr.FileURL, &sr.FileSize, &sr.FileHash,
			&sr.DocType, &extractedData, &normalizedData, &validationErrors, &sr.CacheHit, &sr.ModelUsed,
			&sr.PromptVersion, &sr.Status, &sr.ErrorMessage, &sr.ProcessingTimeMs, &sr.CreatedAt, &sr.UpdatedAt); err != nil {
			return nil, err
		}
		if extractedData != nil {
			_ = json.Unmarshal(extractedData, &sr.ExtractedData)
		}
		if normalizedData != nil {
			_ = json.Unmarshal(normalizedData, &sr.NormalizedData)
		}
		if validationErrors != nil {
			_ = json.Unmarshal(validationErrors, &sr.ValidationErrors)
		}
		results = append(results, sr)
	}
	return results, nil
}
