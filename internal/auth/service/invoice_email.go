package service

import (
	"fmt"
	"strings"
	"time"
)

type invoiceData struct {
	OrgName       string
	CustomerName  string
	CustomerEmail string
	PlanName      string
	Period        string
	Amount        int64
	PaymentMethod string
	OrderID       string
	StartsAt      time.Time
	ExpiresAt     time.Time
}

var idMonths = []string{"", "Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

func idDate(t time.Time) string {
	return fmt.Sprintf("%d %s %d", t.Day(), idMonths[int(t.Month())], t.Year())
}

// rupiah formats an integer amount as "Rp 5.990.000".
func rupiah(n int64) string {
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

func periodLabel(p string) string {
	switch strings.ToLower(p) {
	case "yearly", "annual":
		return "Tahunan"
	default:
		return "Bulanan"
	}
}

func invoiceNumber(orderID string, t time.Time) string {
	short := strings.ToUpper(strings.ReplaceAll(orderID, "-", ""))
	if len(short) > 6 {
		short = short[:6]
	}
	return fmt.Sprintf("SULUK-INV-%s-%s", t.Format("20060102"), short)
}

// buildInvoiceEmail returns the subject and HTML body of the combined
// payment-confirmation + invoice/receipt email (Indonesian, Suluk-branded).
func buildInvoiceEmail(d invoiceData) (subject, html string) {
	invNo := invoiceNumber(d.OrderID, d.StartsAt)
	subject = "Pembayaran Berhasil — Invoice " + invNo

	customer := d.CustomerName
	if customer == "" {
		customer = d.CustomerEmail
	}
	method := d.PaymentMethod
	if method == "" {
		method = "-"
	} else {
		method = strings.ToUpper(method)
	}

	html = `<!doctype html><html><body style="margin:0;background:#f6f8f7;font-family:Inter,Arial,sans-serif;color:#10211c;">
  <div style="max-width:560px;margin:0 auto;padding:24px;">
    <div style="background:#0F3D2E;border-radius:16px 16px 0 0;padding:28px 32px;color:#fff;">
      <div style="font-family:Georgia,'Playfair Display',serif;font-size:22px;font-weight:700;">Suluk</div>
      <div style="color:#C99A2E;font-size:12px;letter-spacing:.08em;text-transform:uppercase;font-weight:700;margin-top:2px;">ERP for Travel</div>
    </div>
    <div style="background:#fff;border:1px solid #e6ebe8;border-top:0;border-radius:0 0 16px 16px;padding:32px;">
      <div style="display:inline-block;background:#E8F4EF;color:#1B7F5A;font-size:13px;font-weight:700;padding:6px 12px;border-radius:999px;">✓ Pembayaran berhasil</div>
      <h1 style="font-size:20px;margin:18px 0 4px;">Terima kasih, ` + htmlEscape(customer) + `</h1>
      <p style="color:#6b7d77;font-size:14px;margin:0 0 22px;">Langganan <strong>` + htmlEscape(d.PlanName) + `</strong> Anda telah aktif. Berikut invoice pembayaran Anda.</p>

      <table style="width:100%;font-size:13px;color:#6b7d77;margin-bottom:18px;">
        <tr><td>No. Invoice</td><td style="text-align:right;color:#10211c;font-weight:600;">` + invNo + `</td></tr>
        <tr><td>Tanggal</td><td style="text-align:right;color:#10211c;font-weight:600;">` + idDate(d.StartsAt) + `</td></tr>
        <tr><td>Ditagihkan ke</td><td style="text-align:right;color:#10211c;font-weight:600;">` + htmlEscape(orDash(d.OrgName)) + `</td></tr>
        <tr><td>Metode</td><td style="text-align:right;color:#10211c;font-weight:600;">` + htmlEscape(method) + `</td></tr>
      </table>

      <table style="width:100%;border-collapse:collapse;font-size:14px;">
        <thead>
          <tr style="border-bottom:1px solid #e6ebe8;color:#97a6a1;font-size:11.5px;text-transform:uppercase;">
            <th style="text-align:left;padding:8px 0;font-weight:600;">Deskripsi</th>
            <th style="text-align:right;padding:8px 0;font-weight:600;">Jumlah</th>
          </tr>
        </thead>
        <tbody>
          <tr style="border-bottom:1px solid #eef2f0;">
            <td style="padding:12px 0;">Paket ` + htmlEscape(d.PlanName) + ` — ` + periodLabel(d.Period) + `<br><span style="color:#97a6a1;font-size:12px;">Berlaku ` + idDate(d.StartsAt) + ` – ` + idDate(d.ExpiresAt) + `</span></td>
            <td style="padding:12px 0;text-align:right;font-variant-numeric:tabular-nums;">` + rupiah(d.Amount) + `</td>
          </tr>
        </tbody>
      </table>

      <table style="width:100%;font-size:15px;margin-top:14px;">
        <tr><td style="font-weight:700;">Total Dibayar</td><td style="text-align:right;font-weight:800;color:#1B7F5A;font-variant-numeric:tabular-nums;">` + rupiah(d.Amount) + `</td></tr>
      </table>

      <p style="color:#97a6a1;font-size:12px;margin-top:26px;line-height:1.6;">Invoice ini dibuat otomatis dan sah tanpa tanda tangan. Simpan email ini sebagai bukti pembayaran. Jika ada pertanyaan, balas email ini.</p>
    </div>
    <p style="text-align:center;color:#97a6a1;font-size:11px;margin-top:18px;">© Suluk — ERP for Travel</p>
  </div>
</body></html>`
	return subject, html
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}
