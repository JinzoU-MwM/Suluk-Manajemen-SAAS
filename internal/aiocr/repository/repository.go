package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jamaah-in/v2/internal/aiocr/model"
)

type AICacheEntry struct {
	ID             uuid.UUID `json:"id" db:"id"`
	InputHash      string    `json:"input_hash" db:"input_hash"`
	Model          string    `json:"model" db:"model"`
	PromptVersion  string    `json:"prompt_version" db:"prompt_version"`
	TaskType       string    `json:"task_type" db:"task_type"`
	ResultJSON     any       `json:"result_json" db:"result_json"`
	Hits           int       `json:"hits" db:"hits"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	LastAccessedAt time.Time `json:"last_accessed_at" db:"last_accessed_at"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
}

type CacheStatsResult struct {
	TotalEntries        int `json:"total_entries"`
	TotalHits           int `json:"total_hits"`
	ExpiredEntries      int `json:"expired_entries"`
	CacheHitsToday      int `json:"cache_hits_today"`
	ApiCallsToday       int `json:"api_calls_today"`
	TotalProcessingTimeMs int `json:"total_processing_time_ms"`
}

type AIOCRRepo struct {
	pool *pgxpool.Pool
}

func NewAIOCRRepo(pool *pgxpool.Pool) *AIOCRRepo {
	return &AIOCRRepo{pool: pool}
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
		json.Unmarshal(extractedData, &sr.ExtractedData)
	}
	if normalizedData != nil {
		json.Unmarshal(normalizedData, &sr.NormalizedData)
	}
	if validationErrors != nil {
		json.Unmarshal(validationErrors, &sr.ValidationErrors)
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
			json.Unmarshal(extractedData, &sr.ExtractedData)
		}
		if normalizedData != nil {
			json.Unmarshal(normalizedData, &sr.NormalizedData)
		}
		if validationErrors != nil {
			json.Unmarshal(validationErrors, &sr.ValidationErrors)
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
			json.Unmarshal(columnMapping, &t.ColumnMapping)
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

func (r *AIOCRRepo) GetPendingScanResults(ctx context.Context, limit int) ([]model.ScanResult, error) {
	// FOR UPDATE SKIP LOCKED prevents duplicate processing when multiple workers run.
	rows, err := r.pool.Query(ctx,
		`SELECT id, scan_job_id, org_id, file_name, file_url, file_size, file_hash, doc_type,
			extracted_data, normalized_data, validation_errors, cache_hit, model_used, prompt_version,
			status, error_message, processing_time_ms, created_at, updated_at
		FROM scan_results WHERE status = 'pending'
		ORDER BY created_at ASC LIMIT $1
		FOR UPDATE SKIP LOCKED`, limit)
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
			json.Unmarshal(extractedData, &sr.ExtractedData)
		}
		if normalizedData != nil {
			json.Unmarshal(normalizedData, &sr.NormalizedData)
		}
		if validationErrors != nil {
			json.Unmarshal(validationErrors, &sr.ValidationErrors)
		}
		results = append(results, sr)
	}
	return results, nil
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

func (r *AIOCRRepo) CheckCache(ctx context.Context, fileHash, model, promptVersion string) (*AICacheEntry, error) {
	entry := &AICacheEntry{}
	var resultJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, input_hash, model, prompt_version, task_type, result_json, hits, created_at, last_accessed_at, expires_at
		FROM ai_cache WHERE input_hash = $1 AND model = $2 AND prompt_version = $3 AND expires_at > NOW()`,
		fileHash, model, promptVersion).Scan(
		&entry.ID, &entry.InputHash, &entry.Model, &entry.PromptVersion, &entry.TaskType,
		&resultJSON, &entry.Hits, &entry.CreatedAt, &entry.LastAccessedAt, &entry.ExpiresAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if resultJSON != nil {
		var data any
		if json.Unmarshal(resultJSON, &data) == nil {
			entry.ResultJSON = data
		}
	}
	return entry, nil
}

func (r *AIOCRRepo) StoreInCache(ctx context.Context, inputHash, model, promptVersion, taskType string, resultJSON any, docType *string, ttl time.Duration) error {
	resultBytes, _ := json.Marshal(resultJSON)
	expiresAt := time.Now().Add(ttl)

	// Conflict on (input_hash, model, prompt_version) so different models don't overwrite each other.
	_, err := r.pool.Exec(ctx,
		`INSERT INTO ai_cache (input_hash, model, prompt_version, task_type, result_json, hits, expires_at)
		VALUES ($1, $2, $3, $4, $5, 1, $6)
		ON CONFLICT (input_hash, model, prompt_version) DO UPDATE SET
			hits = ai_cache.hits + 1,
			last_accessed_at = NOW(),
			expires_at = $6,
			result_json = $5`,
		inputHash, model, promptVersion, taskType, resultBytes, expiresAt)
	return err
}

func (r *AIOCRRepo) IncrementCacheHit(ctx context.Context, inputHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE ai_cache SET hits = hits + 1, last_accessed_at = NOW() WHERE input_hash = $1`, inputHash)
	return err
}

func (r *AIOCRRepo) GetCacheStats(ctx context.Context, orgID uuid.UUID) (*CacheStatsResult, error) {
	stats := &CacheStatsResult{}
	today := "created_at >= CURRENT_DATE"

	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM ai_cache`).Scan(&stats.TotalEntries); err != nil {
		return nil, fmt.Errorf("cache total_entries: %w", err)
	}
	if err := r.pool.QueryRow(ctx, `SELECT COALESCE(SUM(hits), 0) FROM ai_cache`).Scan(&stats.TotalHits); err != nil {
		return nil, fmt.Errorf("cache total_hits: %w", err)
	}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM ai_cache WHERE expires_at < NOW()`).Scan(&stats.ExpiredEntries); err != nil {
		return nil, fmt.Errorf("cache expired_entries: %w", err)
	}
	// "Today" queries: filter to current day for accuracy
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM scan_results WHERE cache_hit = TRUE AND org_id = $1 AND `+today, orgID,
	).Scan(&stats.CacheHitsToday); err != nil {
		return nil, fmt.Errorf("cache_hits_today: %w", err)
	}
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM scan_results WHERE cache_hit IS NOT TRUE AND org_id = $1 AND status = 'completed' AND `+today, orgID,
	).Scan(&stats.ApiCallsToday); err != nil {
		return nil, fmt.Errorf("api_calls_today: %w", err)
	}
	if err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(processing_time_ms), 0) FROM scan_results WHERE status = 'completed' AND org_id = $1`, orgID,
	).Scan(&stats.TotalProcessingTimeMs); err != nil {
		return nil, fmt.Errorf("total_processing_time: %w", err)
	}

	return stats, nil
}

func (r *AIOCRRepo) ClearCache(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM ai_cache WHERE expires_at < NOW()`)
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
			json.Unmarshal(extractedData, &sr.ExtractedData)
		}
		if normalizedData != nil {
			json.Unmarshal(normalizedData, &sr.NormalizedData)
		}
		if validationErrors != nil {
			json.Unmarshal(validationErrors, &sr.ValidationErrors)
		}
		results = append(results, sr)
	}
	return results, nil
}