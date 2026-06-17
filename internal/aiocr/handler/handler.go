package handler

import (
	"errors"
	"io"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/aiocr/model"
	"github.com/jamaah-in/v2/internal/aiocr/service"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	"github.com/jamaah-in/v2/internal/shared/response"
)

type AIOCRHandler struct {
	svc *service.AIOCRService
}

func NewAIOCRHandler(svc *service.AIOCRService) *AIOCRHandler {
	return &AIOCRHandler{svc: svc}
}

func (h *AIOCRHandler) CreateScanJob(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)

	var req model.CreateScanJobRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if len(req.Files) == 0 {
		return response.BadRequest(c, "at least one file is required")
	}

	job, err := h.svc.CreateScanJob(c.Context(), claims.OrgID, claims.UserID, req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Created(c, job)
}

// ProcessDocuments OCRs uploaded files synchronously and returns normalized
// records immediately. Matches the scanner frontend's contract:
// multipart/form-data in (field "files" + optional ?cache_mode=), and a
// synchronous {data, validation_warnings, file_results} response out.
func (h *AIOCRHandler) ProcessDocuments(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return response.BadRequest(c, "invalid multipart form")
	}
	fileHeaders := form.File["files"]
	if len(fileHeaders) == 0 {
		return response.BadRequest(c, "at least one file is required")
	}

	const maxFileSize = 20 * 1024 * 1024 // 20 MB per file
	files := make([]service.SyncFile, 0, len(fileHeaders))
	for _, fh := range fileHeaders {
		if fh.Size > maxFileSize {
			return response.BadRequest(c, "file terlalu besar (maks 20MB): "+fh.Filename)
		}
		f, err := fh.Open()
		if err != nil {
			return response.BadRequest(c, "gagal membuka file: "+fh.Filename)
		}
		data, readErr := io.ReadAll(f)
		f.Close()
		if readErr != nil {
			return response.Internal(c, readErr)
		}
		files = append(files, service.SyncFile{
			FileName:    fh.Filename,
			ContentType: fh.Header.Get("Content-Type"),
			Data:        data,
		})
	}

	result, err := h.svc.ProcessDocumentsSync(c.Context(), files, c.Query("cache_mode", "default"))
	if err != nil {
		if errors.Is(err, service.ErrOCRUnavailable) {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"success": false, "error": err.Error()})
		}
		return response.Internal(c, err)
	}
	return response.OK(c, result)
}

func (h *AIOCRHandler) GetScanJob(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid scan job id")
	}

	job, err := h.svc.GetScanJob(c.Context(), id, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "scan job not found")
	}
	return response.OK(c, job)
}

func (h *AIOCRHandler) ListScanJobs(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	status := c.Query("status")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("page_size", "20"))

	jobs, total, err := h.svc.ListScanJobs(c.Context(), claims.OrgID, status, page, limit)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Paginated(c, jobs, int64(total), page, limit)
}

func (h *AIOCRHandler) GetScanResult(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid scan result id")
	}

	result, err := h.svc.GetScanResult(c.Context(), id, claims.OrgID)
	if err != nil {
		return response.NotFound(c, "scan result not found")
	}
	return response.OK(c, result)
}

func (h *AIOCRHandler) GetScanResultsByJob(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	jobID, err := uuid.Parse(c.Params("jobId"))
	if err != nil {
		return response.BadRequest(c, "invalid job id")
	}

	results, err := h.svc.GetScanResultsByJob(c.Context(), claims.OrgID, jobID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, results)
}

func (h *AIOCRHandler) GetCacheStats(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	stats, err := h.svc.GetCacheStats(c.Context(), claims.OrgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, stats)
}

func (h *AIOCRHandler) ClearCache(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	if err := h.svc.ClearCache(c.Context(), claims.OrgID); err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, fiber.Map{"message": "cache cleared"})
}

func (h *AIOCRHandler) NormalizeToSiskopatuh(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.NormalizeRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	result, err := h.svc.NormalizeToSiskopatuh(c.Context(), claims.OrgID, req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, result)
}

func (h *AIOCRHandler) ExportSiskopatuhExcel(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	var req model.ExportSiskopatuhRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	excelData, err := h.svc.ExportSiskopatuhExcel(c.Context(), claims.OrgID, req)
	if err != nil {
		return response.Internal(c, err)
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=siskopatuh_export.xlsx")
	return c.Send(excelData)
}

func (h *AIOCRHandler) CreateExportTemplate(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)

	var req model.CreateExportTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	if req.ColumnMapping == nil {
		return response.BadRequest(c, "column_mapping is required")
	}

	t, err := h.svc.CreateExportTemplate(c.Context(), claims.OrgID, claims.UserID, req)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.Created(c, t)
}

func (h *AIOCRHandler) ListExportTemplates(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	templates, err := h.svc.ListExportTemplates(c.Context(), claims.OrgID)
	if err != nil {
		return response.Internal(c, err)
	}
	return response.OK(c, templates)
}

func (h *AIOCRHandler) DeleteExportTemplate(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid template id")
	}
	if err := h.svc.DeleteExportTemplate(c.Context(), id, claims.OrgID); err != nil {
		return response.NotFound(c, "export template not found")
	}
	return response.OK(c, fiber.Map{"message": "export template deleted"})
}
