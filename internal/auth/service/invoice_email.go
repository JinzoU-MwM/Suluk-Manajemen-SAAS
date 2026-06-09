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
	AppURL        string   // dashboard / public app base, e.g. https://suluk.site
	Features      []string // tier feature list ("Yang termasuk dalam Paket …")
}

// supportWA is the Suluk sales/support WhatsApp (same number the landing page
// uses for "Hubungi Sales").
const (
	supportWANumber  = "6285159980404"
	supportWADisplay = "+62 851-5998-0404"
)

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

// periodUnit renders the recurring-cycle phrase under the amount.
func periodUnit(p string) string {
	switch strings.ToLower(p) {
	case "yearly", "annual":
		return "per tahun · diperpanjang otomatis"
	default:
		return "per bulan · diperpanjang otomatis"
	}
}

func invoiceNumber(orderID string, t time.Time) string {
	short := strings.ToUpper(strings.ReplaceAll(orderID, "-", ""))
	if len(short) > 6 {
		short = short[:6]
	}
	return fmt.Sprintf("SULUK-INV-%s-%s", t.Format("20060102"), short)
}

// featuresGrid renders the "Yang termasuk" list as a Gmail-safe 2-column table
// (unicode check marks, not SVG — Gmail strips inline SVG).
func featuresGrid(features []string) string {
	if len(features) == 0 {
		return ""
	}
	const check = `<td valign="top" style="padding-right:9px;color:#1B7F5A;font-size:15px;font-weight:800;line-height:1.4;">&#10003;</td>`
	cell := func(text string) string {
		if text == "" {
			return `<td class="em-feat" width="50%">&nbsp;</td>`
		}
		return `<td class="em-feat" width="50%" valign="top" style="padding:0 10px 14px 0;">` +
			`<table role="presentation" cellpadding="0" cellspacing="0"><tr>` + check +
			`<td valign="top" style="font-size:13.5px;color:#3f5650;line-height:1.5;">` + htmlEscape(text) + `</td>` +
			`</tr></table></td>`
	}
	var b strings.Builder
	for i := 0; i < len(features); i += 2 {
		b.WriteString(`<tr>`)
		b.WriteString(cell(features[i]))
		if i+1 < len(features) {
			b.WriteString(cell(features[i+1]))
		} else {
			b.WriteString(`<td class="em-feat" width="50%">&nbsp;</td>`)
		}
		b.WriteString(`</tr>`)
	}
	return b.String()
}

// buildInvoiceEmail returns the subject and HTML body of the subscription
// payment-confirmation email (Indonesian, Suluk-branded). It mirrors the
// "Email Konfirmasi Langganan" design: deep-green header, mint success badge,
// amount block, billing-detail rows, dashboard CTA, included-features grid,
// help note, and footer. Built table-based with inline styles for email
// clients; check marks use the unicode glyph so they render in Gmail.
func buildInvoiceEmail(d invoiceData) (subject, html string) {
	invNo := invoiceNumber(d.OrderID, d.StartsAt)
	subject = "Pembayaran Berhasil — Invoice " + invNo

	customer := strings.TrimSpace(d.CustomerName)
	if customer == "" {
		customer = d.CustomerEmail
	}
	method := strings.TrimSpace(d.PaymentMethod)
	if method == "" {
		method = "-"
	} else {
		method = strings.ToUpper(method)
	}
	appURL := strings.TrimRight(strings.TrimSpace(d.AppURL), "/")
	if appURL == "" {
		appURL = "https://suluk.site"
	}
	waURL := "https://wa.me/" + supportWANumber

	html = `<!DOCTYPE html>
<html lang="id"><head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<meta name="x-apple-disable-message-reformatting" />
<title>Suluk — Konfirmasi Langganan</title>
<style>
  body { margin:0; padding:0; background:#eef2f0; -webkit-font-smoothing:antialiased; }
  a { text-decoration:none; }
  .em-body { font-family:'Inter',-apple-system,system-ui,Arial,sans-serif; }
  .em-display { font-family:'Playfair Display',Georgia,serif; }
  .tnum { font-variant-numeric:tabular-nums; }
  @media (max-width:620px){
    .em-pad { padding-left:24px !important; padding-right:24px !important; }
    .em-h1 { font-size:26px !important; }
    .em-amt { font-size:30px !important; }
    .em-col { display:block !important; width:100% !important; }
    .em-col-r { text-align:left !important; padding-top:10px !important; }
    .em-feat { display:block !important; width:100% !important; padding:0 0 12px 0 !important; }
  }
</style></head>
<body class="em-body">
  <div style="display:none;max-height:0;overflow:hidden;opacity:0;color:#eef2f0;font-size:1px;line-height:1px;">Pembayaran berhasil — langganan Paket ` + htmlEscape(d.PlanName) + ` Anda kini aktif. Terima kasih telah bergabung bersama Suluk.</div>
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#eef2f0;"><tr>
    <td align="center" style="padding:40px 16px;">

      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="width:600px;max-width:100%;background:#ffffff;border-radius:20px;overflow:hidden;box-shadow:0 18px 48px rgba(15,61,46,0.12);">

        <!-- HEADER -->
        <tr><td style="background:#0F3D2E;padding:26px 40px;" class="em-pad">
          <table role="presentation" width="100%" cellpadding="0" cellspacing="0"><tr>
            <td align="left" valign="middle">
              <div class="em-display" style="color:#ffffff;font-size:21px;font-weight:800;line-height:1;">Suluk</div>
              <div style="color:#C99A2E;font-size:9px;font-weight:700;letter-spacing:2.6px;margin-top:5px;">ERP FOR TRAVEL</div>
            </td>
            <td align="right" valign="middle">
              <span style="color:#7FB5A1;font-size:12px;font-weight:600;">Konfirmasi&nbsp;Langganan</span>
            </td>
          </tr></table>
        </td></tr>

        <!-- SUCCESS BADGE -->
        <tr><td align="center" style="padding:44px 40px 0;" class="em-pad">
          <div style="width:74px;height:74px;background:#E8F4EF;border-radius:50%;line-height:74px;text-align:center;color:#1B7F5A;font-size:40px;font-weight:700;">&#10003;</div>
        </td></tr>
        <tr><td align="center" style="padding:22px 40px 0;" class="em-pad">
          <div class="em-display em-h1" style="font-size:30px;font-weight:800;color:#10211c;letter-spacing:-0.5px;">Pembayaran Berhasil!</div>
          <p style="margin:14px 0 0;font-size:15.5px;line-height:1.6;color:#6b7d77;">Terima kasih, <strong style="color:#10211c;">` + htmlEscape(customer) + `</strong>. Langganan <strong style="color:#10211c;">Paket ` + htmlEscape(d.PlanName) + `</strong> Anda kini <strong style="color:#1B7F5A;">aktif</strong>. Seluruh modul Suluk siap digunakan untuk mengelola travel Anda.</p>
        </td></tr>

        <!-- AMOUNT -->
        <tr><td style="padding:28px 40px 0;" class="em-pad">
          <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#F2F8F5;border:1px solid #E8F4EF;border-radius:16px;"><tr>
            <td style="padding:22px 24px;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0"><tr>
                <td class="em-col" align="left" valign="middle">
                  <div style="font-size:11.5px;font-weight:700;letter-spacing:0.6px;text-transform:uppercase;color:#7d9a90;">Total Dibayar</div>
                  <div class="em-display em-amt tnum" style="font-size:34px;font-weight:800;color:#0F3D2E;line-height:1.1;margin-top:4px;">` + rupiah(d.Amount) + `</div>
                  <div style="font-size:13px;color:#6b7d77;margin-top:2px;">` + periodUnit(d.Period) + `</div>
                </td>
                <td class="em-col em-col-r" align="right" valign="middle">
                  <span style="display:inline-block;background:#E8F4EF;color:#0F3D2E;font-size:12.5px;font-weight:700;padding:7px 14px;border-radius:999px;">&#10003; Lunas</span>
                </td>
              </tr></table>
            </td>
          </tr></table>
        </td></tr>

        <!-- DETAIL ROWS -->
        <tr><td style="padding:26px 40px 0;" class="em-pad">
          <div style="font-size:12px;font-weight:700;letter-spacing:0.6px;text-transform:uppercase;color:#97a6a1;padding-bottom:6px;">Rincian Tagihan</div>
          <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="font-size:14px;">
            <tr><td style="padding:13px 0;border-bottom:1px solid #eef2f0;color:#6b7d77;">No. Invoice</td><td align="right" style="padding:13px 0;border-bottom:1px solid #eef2f0;color:#10211c;font-weight:700;" class="tnum">` + invNo + `</td></tr>
            <tr><td style="padding:13px 0;border-bottom:1px solid #eef2f0;color:#6b7d77;">Paket</td><td align="right" style="padding:13px 0;border-bottom:1px solid #eef2f0;color:#10211c;font-weight:700;">` + htmlEscape(d.PlanName) + ` — ` + periodLabel(d.Period) + `</td></tr>
            <tr><td style="padding:13px 0;border-bottom:1px solid #eef2f0;color:#6b7d77;">Periode Tagihan</td><td align="right" style="padding:13px 0;border-bottom:1px solid #eef2f0;color:#10211c;font-weight:700;" class="tnum">` + idDate(d.StartsAt) + ` – ` + idDate(d.ExpiresAt) + `</td></tr>
            <tr><td style="padding:13px 0;border-bottom:1px solid #eef2f0;color:#6b7d77;">Metode Pembayaran</td><td align="right" style="padding:13px 0;border-bottom:1px solid #eef2f0;color:#10211c;font-weight:700;">` + htmlEscape(method) + `</td></tr>
            <tr><td style="padding:13px 0;color:#6b7d77;">Tagihan Berikutnya</td><td align="right" style="padding:13px 0;color:#10211c;font-weight:700;" class="tnum">` + idDate(d.ExpiresAt) + `</td></tr>
          </table>
        </td></tr>

        <!-- CTA -->
        <tr><td align="center" style="padding:30px 40px 4px;" class="em-pad">
          <table role="presentation" cellpadding="0" cellspacing="0"><tr>
            <td align="center" style="border-radius:12px;background:#1B7F5A;box-shadow:0 8px 20px rgba(27,127,90,0.3);">
              <a href="` + appURL + `" style="display:inline-block;padding:15px 34px;font-size:15px;font-weight:700;color:#ffffff;">Buka Dashboard &rarr;</a>
            </td>
          </tr></table>
        </td></tr>
` + includedSection(d.PlanName, d.Features) + `
        <!-- HELP NOTE -->
        <tr><td style="padding:22px 40px 38px;" class="em-pad">
          <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#FBF7EC;border-radius:14px;"><tr>
            <td style="padding:16px 20px;font-size:13px;color:#8a6d2a;line-height:1.55;">Butuh bantuan menyiapkan akun? Tim kami siap membantu via WhatsApp di <a href="` + waURL + `" style="color:#C99A2E;font-weight:700;">` + supportWADisplay + `</a>.</td>
          </tr></table>
        </td></tr>

      </table>

      <!-- FOOTER -->
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="width:600px;max-width:100%;"><tr>
        <td align="center" style="padding:26px 40px 10px;" class="em-pad">
          <div style="font-size:12.5px;color:#97a6a1;line-height:1.7;">Email ini dikirim ke <a href="mailto:` + htmlEscape(d.CustomerEmail) + `" style="color:#6b7d77;font-weight:600;">` + htmlEscape(d.CustomerEmail) + `</a> terkait langganan Anda.<br />Suluk · ERP for Travel · suluk.site</div>
          <div style="padding-top:14px;font-size:12.5px;">
            <a href="` + appURL + `" style="color:#1B7F5A;font-weight:600;">Beranda</a>
            <span style="color:#cdd6d2;padding:0 8px;">·</span>
            <a href="` + waURL + `" style="color:#1B7F5A;font-weight:600;">Bantuan</a>
            <span style="color:#cdd6d2;padding:0 8px;">·</span>
            <a href="` + appURL + `" style="color:#1B7F5A;font-weight:600;">Kelola Langganan</a>
          </div>
          <div style="padding-top:16px;font-size:11.5px;color:#b6c2bc;">&copy; ` + fmt.Sprintf("%d", d.StartsAt.Year()) + ` Suluk — ERP for Travel</div>
        </td>
      </tr></table>

    </td>
  </tr></table>
</body></html>`
	return subject, html
}

// includedSection renders the "Yang termasuk dalam Paket …" block, or nothing
// when the tier has no feature list.
func includedSection(planName string, features []string) string {
	grid := featuresGrid(features)
	if grid == "" {
		return ""
	}
	return `
        <!-- INCLUDED -->
        <tr><td style="padding:30px 40px 0;" class="em-pad">
          <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="border-top:1px solid #eef2f0;"><tr><td style="padding-top:24px;">
            <div style="font-size:15px;font-weight:800;color:#10211c;padding-bottom:16px;">Yang termasuk dalam Paket ` + htmlEscape(planName) + `</div>
            <table role="presentation" width="100%" cellpadding="0" cellspacing="0">` + grid + `</table>
          </td></tr></table>
        </td></tr>
`
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}
