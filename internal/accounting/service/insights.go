package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/accounting/model"
)

// GenerateInsights builds a financial-insights report from the GL: a rule-based
// anomaly/highlight engine over the balance sheet + income statement, plus an
// optional AI narrative when a Gemini key is configured. The rule-based core is
// fully deterministic and works without AI.
func (s *Service) GenerateInsights(ctx context.Context, orgID uuid.UUID, asOf time.Time) (*model.InsightReport, error) {
	from := time.Date(asOf.Year(), asOf.Month(), 1, 0, 0, 0, 0, asOf.Location())

	bs, err := s.BalanceSheet(ctx, orgID, asOf)
	if err != nil {
		return nil, err
	}
	is, err := s.IncomeStatement(ctx, orgID, from, asOf)
	if err != nil {
		return nil, err
	}
	tb, err := s.TrialBalance(ctx, orgID, asOf)
	if err != nil {
		return nil, err
	}

	var tbDebit, tbCredit int64
	for _, r := range tb {
		tbDebit += r.Debit
		tbCredit += r.Credit
	}
	cash := lineAmount(bs.Assets, "1101") + lineAmount(bs.Assets, "1102")
	ar := lineAmount(bs.Assets, "1201")

	m := model.InsightMetrics{
		TotalAssets:          bs.TotalAssets,
		TotalLiabilities:     bs.TotalLiabilities,
		TotalEquity:          bs.TotalEquity,
		Cash:                 cash,
		Receivables:          ar,
		Revenue:              is.TotalRevenue,
		Expense:              is.TotalExpense,
		NetIncome:            is.NetIncome,
		BalanceSheetBalanced: bs.Balanced,
		LedgerBalanced:       tbDebit == tbCredit,
	}

	rep := &model.InsightReport{
		AsOf:        asOf.Format("2006-01-02"),
		PeriodFrom:  from.Format("2006-01-02"),
		PeriodTo:    asOf.Format("2006-01-02"),
		Metrics:     m,
		Anomalies:   []model.Insight{},
		Highlights:  []model.Insight{},
		AIAvailable: s.ai.Available(),
	}

	// ── Rule-based anomalies ──
	if !m.LedgerBalanced {
		rep.Anomalies = append(rep.Anomalies, model.Insight{Severity: "critical", Title: "Buku besar tidak seimbang",
			Detail: fmt.Sprintf("Total debit %s ≠ total kredit %s. Periksa jurnal manual.", rp(tbDebit), rp(tbCredit))})
	}
	if !m.BalanceSheetBalanced {
		rep.Anomalies = append(rep.Anomalies, model.Insight{Severity: "critical", Title: "Neraca tidak seimbang",
			Detail: fmt.Sprintf("Aset %s ≠ Kewajiban+Ekuitas %s.", rp(m.TotalAssets), rp(m.TotalLiabilities+m.TotalEquity))})
	}
	if cash < 0 {
		rep.Anomalies = append(rep.Anomalies, model.Insight{Severity: "warning", Title: "Saldo kas/bank negatif",
			Detail: fmt.Sprintf("Saldo kas+bank %s. Kemungkinan pencatatan pengeluaran melebihi penerimaan.", rp(cash))})
	}
	if m.Revenue > 0 && m.Expense > m.Revenue {
		rep.Anomalies = append(rep.Anomalies, model.Insight{Severity: "warning", Title: "Beban melebihi pendapatan",
			Detail: fmt.Sprintf("Beban %s > pendapatan %s bulan ini.", rp(m.Expense), rp(m.Revenue))})
	} else if m.NetIncome < 0 {
		rep.Anomalies = append(rep.Anomalies, model.Insight{Severity: "warning", Title: "Rugi periode berjalan",
			Detail: fmt.Sprintf("Laba/rugi bulan ini %s.", rp(m.NetIncome))})
	}
	if m.Revenue > 0 && ar > m.Revenue {
		rep.Anomalies = append(rep.Anomalies, model.Insight{Severity: "info", Title: "Piutang jemaah besar",
			Detail: fmt.Sprintf("Piutang %s melebihi pendapatan bulan ini %s — tingkatkan penagihan.", rp(ar), rp(m.Revenue))})
	}

	// ── Highlights ──
	if m.NetIncome > 0 {
		rep.Highlights = append(rep.Highlights, model.Insight{Severity: "good", Title: "Laba bulan ini",
			Detail: fmt.Sprintf("Laba bersih %s dari pendapatan %s.", rp(m.NetIncome), rp(m.Revenue))})
	}
	rep.Highlights = append(rep.Highlights, model.Insight{Severity: "info", Title: "Posisi kas",
		Detail: fmt.Sprintf("Kas+bank %s; piutang jemaah %s.", rp(cash), rp(ar))})
	if top := topExpense(is.Expenses); top != nil {
		rep.Highlights = append(rep.Highlights, model.Insight{Severity: "info", Title: "Beban terbesar",
			Detail: fmt.Sprintf("%s sebesar %s.", top.Name, rp(top.Amount))})
	}

	// ── Optional AI narrative (memoized) ──
	// The prompt is fully derived from the numbers above, so an unchanged GL hits
	// the cache and skips the Gemini call; changed numbers (or a new day) miss and
	// regenerate. This keeps us well under free-tier rate limits.
	if s.ai.Available() {
		prompt := insightPrompt(rep)
		now := time.Now()
		key := narrativeKey(orgID, prompt)
		if cached, ok := s.narr.get(key, now); ok {
			rep.AINarrative = cached
			rep.AICached = true
		} else if narrative, aerr := s.ai.GenerateText(ctx, prompt); aerr == nil {
			rep.AINarrative = strings.TrimSpace(narrative)
			s.narr.put(key, rep.AINarrative, now)
		} else if s.log != nil {
			s.log.Warnw("copilot AI narrative failed", "org_id", orgID, "err", aerr)
		}
	}

	return rep, nil
}

func lineAmount(lines []model.StatementLine, code string) int64 {
	for _, l := range lines {
		if l.Code == code {
			return l.Amount
		}
	}
	return 0
}

func topExpense(lines []model.StatementLine) *model.StatementLine {
	var top *model.StatementLine
	for i := range lines {
		if top == nil || lines[i].Amount > top.Amount {
			top = &lines[i]
		}
	}
	return top
}

// rp formats an IDR amount (whole rupiah) for human-readable details.
func rp(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := fmt.Sprintf("%d", n)
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, c)
	}
	res := "Rp " + string(out)
	if neg {
		res = "-" + res
	}
	return res
}

func insightPrompt(r *model.InsightReport) string {
	var b strings.Builder
	b.WriteString("Anda asisten keuangan untuk biro travel umroh/haji. Berdasarkan ringkasan keuangan berikut, ")
	b.WriteString("berikan 2-3 kalimat analisis singkat dalam Bahasa Indonesia (bukan poin, langsung paragraf) ")
	b.WriteString("yang menyoroti kondisi keuangan dan satu rekomendasi praktis. Jangan mengulang angka mentah secara berlebihan.\n\n")
	m := r.Metrics
	fmt.Fprintf(&b, "Periode: %s s/d %s\n", r.PeriodFrom, r.PeriodTo)
	fmt.Fprintf(&b, "Total aset: %s; kewajiban: %s; ekuitas: %s\n", rp(m.TotalAssets), rp(m.TotalLiabilities), rp(m.TotalEquity))
	fmt.Fprintf(&b, "Kas+bank: %s; piutang jemaah: %s\n", rp(m.Cash), rp(m.Receivables))
	fmt.Fprintf(&b, "Pendapatan: %s; beban: %s; laba/rugi: %s\n", rp(m.Revenue), rp(m.Expense), rp(m.NetIncome))
	if len(r.Anomalies) > 0 {
		b.WriteString("Temuan: ")
		for i, a := range r.Anomalies {
			if i > 0 {
				b.WriteString("; ")
			}
			b.WriteString(a.Title)
		}
		b.WriteString("\n")
	}
	return b.String()
}
