package handler

import (
	"bytes"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"

	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
)

func (h *PayrollHandler) ExportSlipPDF(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	// Get all slips and find the one matching the ID
	slips, err := h.svc.ListSalarySlips(c.Context(), claims.OrgID.String(), "")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	var target *struct {
		ID           string
		EmployeeName string
		Period       string
		BaseSalary   int64
		Allowance    int64
		Pph21Amount  int64
		BpjsAmount   int64
		NetSalary    int64
	}

	id := c.Params("id")
	for _, s := range slips {
		if s.ID == id {
			target = &struct {
				ID           string
				EmployeeName string
				Period       string
				BaseSalary   int64
				Allowance    int64
				Pph21Amount  int64
				BpjsAmount   int64
				NetSalary    int64
			}{s.ID, s.EmployeeName, s.Period, s.BaseSalary, s.Allowance, s.Pph21Amount, s.BpjsAmount, s.NetSalary}
			break
		}
	}
	if target == nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "slip not found"})
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 14)
	pdf.Cell(190, 8, "SLIP GAJI")
	pdf.Ln(6)
	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(190, 5, "Suluk Travel Management")
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "B", 10)
	pdf.Cell(40, 7, "Karyawan")
	pdf.Cell(5, 7, ":")
	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(0, 7, target.EmployeeName)
	pdf.Ln(7)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.Cell(40, 7, "Periode")
	pdf.Cell(5, 7, ":")
	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(0, 7, target.Period)
	pdf.Ln(12)

	pdf.SetFillColor(230, 230, 230)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.Cell(90, 8, "Komponen")
	pdf.Cell(40, 8, "Jumlah (IDR)")
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(90, 7, "Gaji Pokok")
	pdf.Cell(40, 7, formatIDR(target.BaseSalary))
	pdf.Ln(7)
	pdf.Cell(90, 7, "Tunjangan")
	pdf.Cell(40, 7, formatIDR(target.Allowance))
	pdf.Ln(7)
	gross := target.BaseSalary + target.Allowance
	pdf.SetFont("Helvetica", "B", 9)
	pdf.Cell(90, 7, "Gaji Kotor")
	pdf.Cell(40, 7, formatIDR(gross))
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(90, 7, "PPh 21")
	pdf.Cell(40, 7, "- "+formatIDR(target.Pph21Amount))
	pdf.Ln(7)
	pdf.Cell(90, 7, "BPJS")
	pdf.Cell(40, 7, "- "+formatIDR(target.BpjsAmount))
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetFillColor(220, 240, 255)
	pdf.Cell(90, 8, "GAJI BERSIH")
	pdf.Cell(40, 8, formatIDR(target.NetSalary))
	pdf.Ln(18)

	pdf.SetFont("Helvetica", "I", 9)
	pdf.Cell(0, 5, "Dokumen ini dibuat otomatis oleh sistem Suluk Travel Management.")

	var buf bytes.Buffer
	_ = pdf.Output(&buf)
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", `attachment; filename="slip_gaji.pdf"`)
	return c.Send(buf.Bytes())
}

func formatIDR(n int64) string {
	if n < 0 {
		return fmt.Sprintf("-Rp %d", -n)
	}
	return fmt.Sprintf("Rp %d", n)
}
