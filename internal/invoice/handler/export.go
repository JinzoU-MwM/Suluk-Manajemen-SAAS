package handler

import (
	"bytes"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"

	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
)

func (h *InvoiceHandler) ExportInvoicePDF(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid invoice id"})
	}
	inv, err := h.svc.GetInvoice(c.Context(), id, claims.OrgID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "invoice not found"})
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(190, 10, "KWITANSI")
	pdf.Ln(8)
	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(190, 5, "Suluk Travel Management")
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "B", 10)
	pdf.Cell(40, 7, "No. Invoice")
	pdf.Cell(5, 7, ":")
	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(0, 7, inv.InvoiceNumber)

	pdf.Ln(7)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.Cell(40, 7, "Tanggal")
	pdf.Cell(5, 7, ":")
	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(0, 7, inv.IssuedAt.Format("02 Jan 2006"))

	pdf.Ln(7)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.Cell(40, 7, "Status")
	pdf.Cell(5, 7, ":")
	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(0, 7, string(inv.Status))

	pdf.Ln(12)
	pdf.SetFillColor(230, 230, 230)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.Cell(90, 8, "Deskripsi")
	pdf.Cell(40, 8, "Jumlah (IDR)")
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(90, 7, fmt.Sprintf("Paket Jamaah (Room: %s)", inv.RoomType))
	pdf.Cell(40, 7, formatIDR(inv.TotalAmount))
	pdf.Ln(7)
	pdf.Cell(90, 7, "Diskon")
	pdf.Cell(40, 7, formatIDR(inv.DiscountAmount))
	pdf.Ln(7)
	pdf.Cell(90, 7, "Biaya Tambahan")
	pdf.Cell(40, 7, formatIDR(inv.SurchargeAmount))
	pdf.Ln(7)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.Cell(90, 7, "Total")
	pdf.Cell(40, 7, formatIDR(inv.TotalAmount-inv.DiscountAmount+inv.SurchargeAmount))

	pdf.Ln(10)
	pdf.Cell(90, 7, "Terbayar")
	pdf.Cell(40, 7, formatIDR(inv.AmountPaid))
	pdf.Ln(7)
	pdf.Cell(90, 7, "Sisa")
	pdf.Cell(40, 7, formatIDR(inv.AmountRemaining))

	pdf.Ln(20)
	pdf.SetFont("Helvetica", "I", 9)
	pdf.Cell(0, 5, "Dokumen ini dibuat otomatis oleh sistem Suluk Travel Management.")
	pdf.Ln(5)
	pdf.Cell(0, 5, fmt.Sprintf("Dicetak: %s", inv.IssuedAt.Format("02 Jan 2006 15:04")))

	var buf bytes.Buffer
	pdf.Output(&buf)

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="invoice_%s.pdf"`, inv.InvoiceNumber))
	return c.Send(buf.Bytes())
}

func formatIDR(n int64) string {
	if n < 0 {
		return fmt.Sprintf("-Rp %d", -n)
	}
	return fmt.Sprintf("Rp %d", n)
}
