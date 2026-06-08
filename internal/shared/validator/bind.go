package validator

import (
	"errors"
	"reflect"
	"strings"

	gpv "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/jamaah-in/v2/internal/shared/response"
)

var validate = newValidator()

func newValidator() *gpv.Validate {
	v := gpv.New(gpv.WithRequiredStructEnabled())
	// Report the JSON field name (not the Go struct field name) in errors.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
	return v
}

// BindAndValidate parses the JSON body into dst and validates its `validate:` tags.
// Returns nil on success, or a slice of field errors suitable for response.ValidationError.
func BindAndValidate(c *fiber.Ctx, dst any) []response.FieldError {
	if err := c.BodyParser(dst); err != nil {
		return []response.FieldError{{Field: "body", Message: "format permintaan tidak valid"}}
	}
	return ValidateStruct(dst)
}

// ValidateStruct validates an already-populated struct against its `validate:` tags.
func ValidateStruct(dst any) []response.FieldError {
	if err := validate.Struct(dst); err != nil {
		var verrs gpv.ValidationErrors
		if errors.As(err, &verrs) {
			out := make([]response.FieldError, 0, len(verrs))
			for _, fe := range verrs {
				out = append(out, response.FieldError{Field: fe.Field(), Message: messageFor(fe)})
			}
			return out
		}
		return []response.FieldError{{Field: "body", Message: "validasi gagal"}}
	}
	return nil
}

func messageFor(fe gpv.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "oneof":
		return "harus salah satu dari: " + fe.Param()
	case "min":
		return "minimal " + fe.Param()
	case "max":
		return "maksimal " + fe.Param()
	case "gt":
		return "harus lebih besar dari " + fe.Param()
	case "gte":
		return "minimal " + fe.Param()
	case "len":
		return "panjang harus " + fe.Param()
	default:
		return "nilai tidak valid"
	}
}
