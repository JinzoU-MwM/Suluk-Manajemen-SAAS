package response

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// pkgLogger is set once at startup via SetLogger so Internal() can log the real
// error server-side while returning only a generic message to the client.
var pkgLogger *zap.SugaredLogger

// SetLogger wires the package logger used by Internal(). Call once per service at startup.
func SetLogger(l *zap.SugaredLogger) { pkgLogger = l }

// Internal logs the underlying error (server-side) and returns a generic 500 to
// the client, avoiding leakage of DB/internal details. Prefer this over
// InternalError(c, err.Error()).
func Internal(c *fiber.Ctx, err error) error {
	if pkgLogger != nil && err != nil {
		pkgLogger.Errorw("internal_error",
			"method", c.Method(),
			"path", c.Path(),
			"error", err.Error(),
		)
	}
	return c.Status(http.StatusInternalServerError).JSON(APIResponse{
		Success: false,
		Error:   "terjadi kesalahan pada server",
	})
}

type APIResponse struct {
	Success bool            `json:"success"`
	Data    any             `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
	Errors  []FieldError    `json:"errors,omitempty"`
	Meta    *PaginationMeta `json:"meta,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type PaginationMeta struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Pages    int   `json:"pages"`
}

func Success(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

func Created(c *fiber.Ctx, data any) error {
	return Success(c, http.StatusCreated, data)
}

func OK(c *fiber.Ctx, data any) error {
	return Success(c, http.StatusOK, data)
}

func Paginated(c *fiber.Ctx, data any, total int64, page, pageSize int) error {
	if pageSize < 1 {
		pageSize = 1
	}
	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}
	return c.Status(http.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    data,
		Meta: &PaginationMeta{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			Pages:    pages,
		},
	})
}

func BadRequest(c *fiber.Ctx, msg string) error {
	return c.Status(http.StatusBadRequest).JSON(APIResponse{
		Success: false,
		Error:   msg,
	})
}

func Unauthorized(c *fiber.Ctx, msg string) error {
	return c.Status(http.StatusUnauthorized).JSON(APIResponse{
		Success: false,
		Error:   msg,
	})
}

func Forbidden(c *fiber.Ctx, msg string) error {
	return c.Status(http.StatusForbidden).JSON(APIResponse{
		Success: false,
		Error:   msg,
	})
}

func NotFound(c *fiber.Ctx, msg string) error {
	return c.Status(http.StatusNotFound).JSON(APIResponse{
		Success: false,
		Error:   msg,
	})
}

func Conflict(c *fiber.Ctx, msg string) error {
	return c.Status(http.StatusConflict).JSON(APIResponse{
		Success: false,
		Error:   msg,
	})
}

func InternalError(c *fiber.Ctx, msg string) error {
	return c.Status(http.StatusInternalServerError).JSON(APIResponse{
		Success: false,
		Error:   msg,
	})
}

func ValidationError(c *fiber.Ctx, errors []FieldError) error {
	return c.Status(http.StatusUnprocessableEntity).JSON(APIResponse{
		Success: false,
		Error:   "Validation failed",
		Errors:  errors,
	})
}
