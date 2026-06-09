package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"

	"github.com/jamaah-in/v2/internal/shared/plan"
	"github.com/jamaah-in/v2/internal/shared/sign"
)

var pdfMonths = []string{"", "Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

func pdfDate(t time.Time) string {
	return fmt.Sprintf("%d %s %d", t.Day(), pdfMonths[int(t.Month())], t.Year())
}

// rupiahDots formats 2990000 -> "Rp 2.990.000" (ASCII only for core PDF fonts).
func rupiahDots(n int64) string {
	s := fmt.Sprintf("%d", n)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, c)
	}
	r := "Rp " + string(out)
	if neg {
		r = "-" + r
	}
	return r
}

func pdfInvoiceNumber(orderID string, t time.Time) string {
	short := strings.ToUpper(strings.ReplaceAll(orderID, "-", ""))
	if len(short) > 6 {
		short = short[:6]
	}
	return fmt.Sprintf("SULUK-INV-%s-%s", t.Format("20060102"), short)
}

func pdfPeriodLabel(p string) string {
	switch strings.ToLower(p) {
	case "yearly", "annual":
		return "Tahunan"
	default:
		return "Bulanan"
	}
}

type billingInfo struct {
	OrgName   string `json:"org_name"`
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
}

// SubscriptionInvoicePDF verifies the HMAC signature, loads the paid order, and
// renders an A4 subscription-invoice/receipt PDF mirroring the confirmation
// email. The link is public (embedded in email) so it is protected by the sig
// rather than a JWT. Returns the PDF bytes and a download filename.
func (s *InvoiceService) SubscriptionInvoicePDF(ctx context.Context, orderIDStr, sig string) ([]byte, string, error) {
	if !sign.Valid(orderIDStr, s.internalKey, sig) {
		return nil, "", fmt.Errorf("invalid or missing signature")
	}
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return nil, "", fmt.Errorf("invalid order id")
	}
	order, err := s.repo.GetPaymentOrderByID(ctx, orderID)
	if err != nil {
		return nil, "", err
	}

	// Best-effort billing names from auth-service.
	var bi billingInfo
	if s.authAddr != "" {
		_ = s.httpc.PostJSON(ctx, s.authAddr, "/api/v1/internal/billing-info",
			map[string]string{"X-Internal-Key": s.internalKey},
			map[string]string{"org_id": order.OrgID.String(), "user_id": order.UserID.String()},
			&bi)
	}

	start := order.CreatedAt
	if order.CompletedAt != nil {
		start = *order.CompletedAt
	}
	end := start.AddDate(0, 1, 0)
	if p := strings.ToLower(order.PlanType); p == "yearly" || p == "annual" {
		end = start.AddDate(1, 0, 0)
	}
	tier := plan.Get(order.Plan)
	method := "-"
	if order.PaymentMethod != nil && *order.PaymentMethod != "" {
		method = strings.ToUpper(*order.PaymentMethod)
	}
	status := strings.ToUpper(order.Status)
	if order.Status == "paid" {
		status = "LUNAS"
	}
	invNo := pdfInvoiceNumber(order.ID.String(), start)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(18, 18, 18)
	pdf.AddPage()

	green := func() { pdf.SetTextColor(15, 61, 46) }
	ink := func() { pdf.SetTextColor(16, 33, 28) }
	soft := func() { pdf.SetTextColor(107, 125, 119) }

	// Header band
	pdf.SetFillColor(15, 61, 46)
	pdf.Rect(0, 0, 210, 30, "F")
	pdf.SetXY(18, 9)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 20)
	pdf.Cell(100, 8, "Suluk")
	pdf.SetXY(18, 18)
	pdf.SetTextColor(201, 154, 46)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.Cell(100, 4, "ERP FOR TRAVEL")
	pdf.SetXY(120, 12)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.CellFormat(72, 6, "INVOICE", "", 0, "R", false, 0, "")

	pdf.SetY(40)

	// Meta block
	row := func(label, val string) {
		pdf.SetFont("Helvetica", "", 10)
		soft()
		pdf.CellFormat(45, 7, label, "", 0, "L", false, 0, "")
		ink()
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(0, 7, val, "", 1, "L", false, 0, "")
	}
	row("No. Invoice", invNo)
	row("Tanggal", pdfDate(start))
	row("Status", status)
	pdf.Ln(4)

	// Billed-to
	soft()
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(0, 6, "DITAGIHKAN KEPADA", "", 1, "L", false, 0, "")
	ink()
	pdf.SetFont("Helvetica", "B", 11)
	billTo := bi.OrgName
	if billTo == "" {
		billTo = bi.UserName
	}
	if billTo == "" {
		billTo = "-"
	}
	pdf.CellFormat(0, 6, billTo, "", 1, "L", false, 0, "")
	if bi.UserName != "" || bi.UserEmail != "" {
		soft()
		pdf.SetFont("Helvetica", "", 10)
		line := strings.TrimSpace(bi.UserName)
		if bi.UserEmail != "" {
			if line != "" {
				line += "  "
			}
			line += bi.UserEmail
		}
		pdf.CellFormat(0, 6, line, "", 1, "L", false, 0, "")
	}
	pdf.Ln(6)

	// Line-items table
	pdf.SetFillColor(232, 244, 239)
	green()
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(130, 9, "  Deskripsi", "", 0, "L", true, 0, "")
	pdf.CellFormat(44, 9, "Jumlah  ", "", 1, "R", true, 0, "")

	ink()
	pdf.SetFont("Helvetica", "", 10)
	desc := fmt.Sprintf("  Langganan Paket %s - %s", tier.Name, pdfPeriodLabel(order.PlanType))
	pdf.CellFormat(130, 9, desc, "B", 0, "L", false, 0, "")
	pdf.CellFormat(44, 9, rupiahDots(order.Amount)+"  ", "B", 1, "R", false, 0, "")

	soft()
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(130, 7, fmt.Sprintf("  Periode: %s - %s", pdfDate(start), pdfDate(end)), "B", 0, "L", false, 0, "")
	pdf.CellFormat(44, 7, "", "B", 1, "R", false, 0, "")
	pdf.CellFormat(130, 7, fmt.Sprintf("  Metode Pembayaran: %s", method), "B", 0, "L", false, 0, "")
	pdf.CellFormat(44, 7, "", "B", 1, "R", false, 0, "")

	// Total
	pdf.Ln(2)
	ink()
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(130, 10, "  Total Dibayar", "", 0, "L", false, 0, "")
	green()
	pdf.CellFormat(44, 10, rupiahDots(order.Amount)+"  ", "", 1, "R", false, 0, "")

	// Footer note
	pdf.Ln(16)
	soft()
	pdf.SetFont("Helvetica", "I", 9)
	pdf.MultiCell(0, 5, "Invoice ini dibuat otomatis dan sah tanpa tanda tangan. Simpan sebagai bukti pembayaran langganan Anda.\nSuluk - ERP for Travel - suluk.site", "", "L", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), "Invoice-" + invNo + ".pdf", nil
}
