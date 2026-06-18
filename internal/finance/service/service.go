package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jamaah-in/v2/internal/finance/model"
	"github.com/jamaah-in/v2/internal/finance/repository"
	"github.com/jamaah-in/v2/internal/shared/httpclient"
)

type FinanceService struct {
	repo        *repository.FinanceRepo
	invoiceAddr string
	vendorAddr  string
	packageAddr string
	jamaahAddr  string
	httpc       *httpclient.Client
}

func NewFinanceService(repo *repository.FinanceRepo, invoiceAddr, vendorAddr, packageAddr, jamaahAddr string) *FinanceService {
	return &FinanceService{
		repo:        repo,
		invoiceAddr: invoiceAddr,
		vendorAddr:  vendorAddr,
		packageAddr: packageAddr,
		jamaahAddr:  jamaahAddr,
		httpc:       httpclient.New(),
	}
}

func (s *FinanceService) CreateExpense(ctx context.Context, orgID uuid.UUID, req model.CreateExpenseRequest) (*model.TripExpense, error) {
	if req.Currency == "" {
		req.Currency = "IDR"
	}
	if req.ExchangeRate == 0 {
		req.ExchangeRate = 1.0
	}
	if req.Status == "" {
		req.Status = "belum_bayar"
	}

	expenseDate, err := repository.ParseDate(req.ExpenseDate)
	if err != nil {
		return nil, err
	}
	if expenseDate == nil {
		return nil, fmt.Errorf("expense_date is required")
	}

	var dueDate *time.Time
	if req.DueDate != "" {
		dueDate, err = repository.ParseDate(req.DueDate)
		if err != nil {
			return nil, err
		}
	}

	e := &model.TripExpense{
		ID:           uuid.New(),
		OrgID:        orgID,
		PackageID:    req.PackageID,
		Category:     req.Category,
		Description:  req.Description,
		VendorName:   strPtr(req.VendorName),
		Amount:       req.Amount,
		Currency:     req.Currency,
		ExchangeRate: req.ExchangeRate,
		// AmountIDR is a DB GENERATED column; CreateExpense RETURNINGs it.
		ExpenseDate: *expenseDate,
		DueDate:     dueDate,
		Status:      req.Status,
	}

	if err := s.repo.CreateExpense(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *FinanceService) GetExpense(ctx context.Context, id, orgID uuid.UUID) (*model.TripExpense, error) {
	return s.repo.GetExpenseByID(ctx, id, orgID)
}

func (s *FinanceService) UpdateExpense(ctx context.Context, id, orgID uuid.UUID, req model.UpdateExpenseRequest) (*model.TripExpense, error) {
	e, err := s.repo.GetExpenseByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	if req.Category != nil {
		e.Category = *req.Category
	}
	if req.Description != nil {
		e.Description = *req.Description
	}
	if req.VendorName != nil {
		e.VendorName = req.VendorName
	}
	if req.Amount != nil {
		e.Amount = *req.Amount
	}
	if req.Currency != nil {
		e.Currency = *req.Currency
	}
	if req.ExchangeRate != nil {
		e.ExchangeRate = *req.ExchangeRate
	}
	if req.ExpenseDate != nil {
		t, err := repository.ParseDate(*req.ExpenseDate)
		if err != nil {
			return nil, err
		}
		if t != nil {
			e.ExpenseDate = *t
		}
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			e.DueDate = nil
		} else {
			t, err := repository.ParseDate(*req.DueDate)
			if err != nil {
				return nil, err
			}
			e.DueDate = t
		}
	}
	if req.Status != nil {
		e.Status = *req.Status
	}

	// AmountIDR is a DB GENERATED column and is repopulated by the re-read below.
	if err := s.repo.UpdateExpense(ctx, e); err != nil {
		return nil, err
	}
	return s.repo.GetExpenseByID(ctx, id, orgID)
}

func (s *FinanceService) DeleteExpense(ctx context.Context, id, orgID uuid.UUID) error {
	return s.repo.DeleteExpense(ctx, id, orgID)
}

func (s *FinanceService) ListExpenses(ctx context.Context, orgID uuid.UUID, category, status string, page, limit int) ([]model.TripExpense, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.repo.ListExpenses(ctx, orgID, category, status, offset, limit)
}

func (s *FinanceService) ListExpensesByPackage(ctx context.Context, orgID, packageID uuid.UUID) ([]model.TripExpense, error) {
	return s.repo.ListExpensesByPackage(ctx, orgID, packageID)
}

func (s *FinanceService) GetSummary(ctx context.Context, orgID uuid.UUID, packageID *uuid.UUID) (*model.ExpenseSummary, error) {
	return s.repo.GetSummary(ctx, orgID, packageID)
}

func (s *FinanceService) GetOverdueExpenses(ctx context.Context, orgID uuid.UUID) ([]model.TripExpense, error) {
	return s.repo.GetOverdueExpenses(ctx, orgID)
}

// --- P&L Aggregation ---

type packageSnapshot struct {
	ID             uuid.UUID              `json:"id"`
	Name           string                 `json:"name"`
	TotalSeats     int                    `json:"total_seats"`
	ReservedSeats  int                    `json:"reserved_seats"`
	PricingTiers   []packagePricingTier   `json:"pricing_tiers"`
	CostComponents []packageCostComponent `json:"cost_components"`
}

type packageOverviewResponse struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	DepartureDate *string   `json:"departure_date,omitempty"`
	TotalSeats    int       `json:"total_seats"`
	ReservedSeats int       `json:"reserved_seats"`
}

type invoiceSummaryResponse struct {
	TotalInvoices    int64 `json:"total_invoices"`
	TotalAmount      int64 `json:"total_amount"`
	TotalPaid        int64 `json:"total_paid"`
	TotalRemaining   int64 `json:"total_remaining"`
	OutstandingCount int64 `json:"outstanding_count"`
	OverdueCount     int64 `json:"overdue_count"`
}

func (s *FinanceService) GetOwnerDashboard(ctx context.Context, orgID uuid.UUID, authToken string) (*model.OwnerDashboard, error) {
	dash := &model.OwnerDashboard{}

	// --- Fan-out: fetch invoice summary, vendor summary, packages, alerts, vendor due concurrently ---
	type invoiceResult struct {
		sum invoiceSummaryResponse
		err error
	}
	type vendorSumResult struct {
		outstanding int64
		err         error
	}
	type packageListResult struct {
		list  []packageOverviewResponse
		total int
		err   error
	}
	type alertsResult struct {
		passportSoon    int
		overdueFollowUp int
		incompleteDocs  int
		err             error
	}
	type vendorDueResult struct {
		count int
		err   error
	}

	invCh := make(chan invoiceResult, 1)
	vendSumCh := make(chan vendorSumResult, 1)
	pkgCh := make(chan packageListResult, 1)
	alertsCh := make(chan alertsResult, 1)
	vendDueCh := make(chan vendorDueResult, 1)

	go func() {
		data, err := s.fetchFromService(ctx, s.invoiceAddr, "/api/v1/invoices/summary", authToken)
		if err != nil {
			invCh <- invoiceResult{err: err}
			return
		}
		var r invoiceSummaryResponse
		json.Unmarshal(data, &r)
		invCh <- invoiceResult{sum: r}
	}()

	go func() {
		data, err := s.fetchFromService(ctx, s.vendorAddr, "/api/v1/vendors/bills/summary", authToken)
		if err != nil {
			vendSumCh <- vendorSumResult{err: err}
			return
		}
		var v struct {
			TotalOutstandingIDR int64 `json:"total_outstanding_idr"`
		}
		json.Unmarshal(data, &v)
		vendSumCh <- vendorSumResult{outstanding: v.TotalOutstandingIDR}
	}()

	go func() {
		list, total, err := s.fetchOpenPackages(ctx, authToken)
		if err != nil {
			pkgCh <- packageListResult{err: err}
			return
		}
		pkgCh <- packageListResult{list: list, total: total}
	}()

	go func() {
		data, err := s.fetchFromService(ctx, s.jamaahAddr, "/api/v1/jamaah/dashboard/alerts", authToken)
		if err != nil {
			alertsCh <- alertsResult{err: err}
			return
		}
		var al struct {
			PassportExpiring90 []any `json:"passport_expiring_90"`
			PassportExpiring30 []any `json:"passport_expiring_30"`
			OverdueFollowUps   []any `json:"overdue_follow_ups"`
			IncompleteDocs     []any `json:"incomplete_docs"`
		}
		json.Unmarshal(data, &al)
		alertsCh <- alertsResult{
			passportSoon:    len(al.PassportExpiring30) + len(al.PassportExpiring90),
			overdueFollowUp: len(al.OverdueFollowUps),
			incompleteDocs:  len(al.IncompleteDocs),
		}
	}()

	go func() {
		data, err := s.fetchFromService(ctx, s.vendorAddr, "/api/v1/vendors/bills/due-soon?days=7", authToken)
		if err != nil {
			vendDueCh <- vendorDueResult{err: err}
			return
		}
		var bills []any
		json.Unmarshal(data, &bills)
		vendDueCh <- vendorDueResult{count: len(bills)}
	}()

	// Collect fan-out results. Any failed source is recorded in `degraded` so the
	// response signals partial data instead of silently reporting zeros.
	var degraded []string

	invRes := <-invCh
	if invRes.err == nil {
		dash.Summary.TotalRevenue = invRes.sum.TotalPaid
		dash.Summary.TotalPiutang = invRes.sum.TotalRemaining
		dash.Summary.OverdueInvoices = invRes.sum.OverdueCount
	} else {
		degraded = append(degraded, "invoices")
	}

	vendSumRes := <-vendSumCh
	if vendSumRes.err == nil {
		dash.Summary.TotalDebt = vendSumRes.outstanding
	} else {
		degraded = append(degraded, "vendor_debt")
	}

	pkgRes := <-pkgCh
	if pkgRes.err != nil {
		degraded = append(degraded, "packages")
	}
	if pkgRes.err == nil {
		dash.Summary.TotalPackages = pkgRes.total

		// Prefetch revenue for ALL packages in ONE call (was an HTTP call +
		// SQL aggregate per package — an O(packages) cross-service N+1).
		type pkgRev struct{ TotalAmount, TotalPaid, TotalRemaining int64 }
		revByPkg := map[string]pkgRev{}
		if revAllData, err := s.fetchFromService(ctx, s.invoiceAddr,
			"/api/v1/invoices/revenue/by-package", authToken); err == nil {
			var all []struct {
				PackageID      string `json:"package_id"`
				TotalAmount    int64  `json:"total_amount"`
				TotalPaid      int64  `json:"total_paid"`
				TotalRemaining int64  `json:"total_remaining"`
			}
			if json.Unmarshal(revAllData, &all) == nil {
				for _, rv := range all {
					revByPkg[rv.PackageID] = pkgRev{rv.TotalAmount, rv.TotalPaid, rv.TotalRemaining}
				}
			}
		}

		ordered := make([]model.PackageOverview, len(pkgRes.list))
		for i, pkg := range pkgRes.list {
			ov := model.PackageOverview{
				ID:            pkg.ID,
				Name:          pkg.Name,
				Status:        pkg.Status,
				TotalSeats:    pkg.TotalSeats,
				ReservedSeats: pkg.ReservedSeats,
			}
			if pkg.DepartureDate != nil {
				t := *pkg.DepartureDate
				if len(t) >= 10 {
					ov.DepartureDate = &t
				}
			}
			if rev, ok := revByPkg[pkg.ID.String()]; ok {
				ov.Revenue = rev.TotalAmount
				ov.Paid = rev.TotalPaid
				ov.Remaining = rev.TotalRemaining
				if rev.TotalAmount > 0 {
					ov.PaymentPct = float64(rev.TotalPaid) * 100 / float64(rev.TotalAmount)
				}
			}
			ordered[i] = ov
		}
		dash.ActivePackages = ordered
	}

	alertRes := <-alertsCh
	if alertRes.err == nil {
		dash.Alerts.PassportExpiringSoon = alertRes.passportSoon
		dash.Alerts.OverdueFollowUps = alertRes.overdueFollowUp
		dash.Alerts.IncompleteDocuments = alertRes.incompleteDocs
	} else {
		degraded = append(degraded, "jamaah_alerts")
	}
	dash.Alerts.OverduePayments = int(dash.Summary.OverdueInvoices)

	vendDueRes := <-vendDueCh
	if vendDueRes.err == nil {
		dash.Alerts.VendorBillsDueSoon = vendDueRes.count
	} else {
		degraded = append(degraded, "vendor_bills_due")
	}

	// Gross profit = revenue collected - local operating expenses
	expenseData, err := s.repo.GetSummary(ctx, orgID, nil)
	if err == nil {
		dash.Summary.GrossProfitMonth = dash.Summary.TotalRevenue - expenseData.TotalAmountIDR
	}

	// Populate RevenueChart: last 6 months from invoice service
	revenueChartData, err := s.fetchFromService(ctx, s.invoiceAddr, "/api/v1/invoices/revenue/monthly?months=6", authToken)
	if err == nil {
		var monthly []model.MonthlyRevenue
		if json.Unmarshal(revenueChartData, &monthly) == nil {
			dash.RevenueChart = monthly
		}
	} else {
		degraded = append(degraded, "revenue_chart")
	}

	if len(degraded) > 0 {
		dash.Partial = true
		dash.DegradedSources = degraded
	}

	return dash, nil
}

type packagePricingTier struct {
	RoomType string `json:"room_type"`
	Price    int64  `json:"price"`
}

type packageCostComponent struct {
	Category        string `json:"category"`
	AmountPerPerson int64  `json:"amount_per_person"`
	Quantity        int    `json:"quantity"`
	TotalAmount     int64  `json:"total_amount"`
}

type vendorBillList struct {
	Items []vendorBillListItem
}

type vendorBillListItem struct {
	Description string  `json:"description"`
	AmountIDR   int64   `json:"amount_idr"`
	VendorType  *string `json:"vendor_type"`
}

func (s *FinanceService) GetPnL(ctx context.Context, orgID, packageID uuid.UUID, authToken string) (*model.PackagePnL, error) {
	pnl := &model.PackagePnL{
		PackageID: packageID,
	}

	pkg, err := s.fetchPackageSnapshot(ctx, packageID, authToken)
	if err != nil {
		return nil, fmt.Errorf("get package snapshot: %w", err)
	}
	pnl.PackageName = pkg.Name
	pnl.TotalSeats = pkg.TotalSeats
	pnl.ReservedSeats = pkg.ReservedSeats

	expenseSummary, err := s.repo.GetSummary(ctx, orgID, &packageID)
	if err != nil {
		return nil, fmt.Errorf("get expense summary: %w", err)
	}
	pnl.OperatingExpenses = expenseSummary
	pnl.TotalOpExpenses = expenseSummary.TotalAmountIDR
	expenses, err := s.repo.ListExpensesByPackage(ctx, orgID, packageID)
	if err != nil {
		return nil, fmt.Errorf("list expenses by package: %w", err)
	}

	// Track failed upstream sources so the P&L can signal partial data instead of
	// silently reporting zeros as if they were real (mirrors the owner dashboard).
	var degraded []string

	invoiceData, err := s.fetchFromService(ctx, s.invoiceAddr, fmt.Sprintf("/api/v1/invoices/package/%s/revenue", packageID), authToken)
	if err == nil {
		var rev model.RevenueSummary
		if err := json.Unmarshal(invoiceData, &rev); err == nil {
			pnl.Revenue = &rev
			pnl.TotalRevenue = rev.TotalBilled
			pnl.RevenueCollected = rev.TotalPaid
			pnl.RevenueOutstanding = rev.TotalRemaining
		} else {
			degraded = append(degraded, "invoices")
		}
	} else {
		degraded = append(degraded, "invoices")
	}

	vendorSummaryData, err := s.fetchFromService(ctx, s.vendorAddr, fmt.Sprintf("/api/v1/vendors/bills/package/%s", packageID), authToken)
	if err == nil {
		var vcs model.VendorCostSummary
		if err := json.Unmarshal(vendorSummaryData, &vcs); err == nil {
			pnl.VendorCosts = &vcs
			pnl.TotalVendorCosts = vcs.TotalAmountIDR
			pnl.VendorPaidOut = vcs.TotalPaidIDR
			pnl.VendorOutstanding = vcs.TotalOutstandingIDR
		} else {
			degraded = append(degraded, "vendor_costs")
		}
	} else {
		degraded = append(degraded, "vendor_costs")
	}

	projectedRevenue, lowestPrice := projectedRevenue(pkg)
	hppPerPerson := sumProjectedPerPerson(pkg)
	projectedCategories, projectedExpense := projectedCategoryTotals(pkg)
	pnl.Projected = &model.ProjectedPnL{
		LowestPrice:              lowestPrice,
		HppPerPerson:             hppPerPerson,
		ProjectedMarginPerPerson: lowestPrice - hppPerPerson,
		Revenue:                  projectedRevenue,
		Expense:                  projectedExpense,
		Profit:                   projectedRevenue - projectedExpense,
		MarginPercent:            marginPercent(projectedRevenue-projectedExpense, projectedRevenue),
	}

	actualCategories := expenseCategoryTotals(expenses)

	vendorBills, err := s.fetchVendorBillList(ctx, packageID, authToken)
	if err == nil {
		for _, bill := range vendorBills {
			key := normalizeVendorBillCategory(bill)
			actualCategories[key] += bill.AmountIDR
		}
	}

	// Profit uses the EXACT operating-expense + vendor-cost summaries, NOT the
	// sampled (<=500) vendor bill breakdown in actualCategories — otherwise a failed
	// or capped vendor bill-list call would silently drop vendor cost and overstate
	// profit. actualCategories is used only for the per-category cost breakdown.
	totalActualCost := pnl.TotalOpExpenses + pnl.TotalVendorCosts
	pnl.Actual = &model.ActualPnL{
		Revenue:       pnl.TotalRevenue,
		Expense:       totalActualCost,
		Profit:        pnl.TotalRevenue - totalActualCost,
		MarginPercent: marginPercent(pnl.TotalRevenue-totalActualCost, pnl.TotalRevenue),
	}

	pnl.CostBreakdown = buildCostBreakdown(projectedCategories, actualCategories)
	if pnl.Revenue != nil {
		pnl.GrossProfit = pnl.TotalRevenue - totalActualCost
		pnl.NetProfit = pnl.GrossProfit
		pnl.CashFlow = pnl.RevenueCollected - pnl.VendorPaidOut - pnl.TotalOpExpenses
	}

	pnl.DataNotes = []string{
		"equipment: reflects vendor procurement cost (perlengkapan bills); distributed inventory HPP requires the inventory module (Phase 3.3)",
		"cost_breakdown: vendor bill category detail is sampled up to 500 bills; summary totals are always exact",
	}

	if len(degraded) > 0 {
		pnl.Partial = true
		pnl.DegradedSources = degraded
	}

	return pnl, nil
}

func (s *FinanceService) fetchPackageSnapshot(ctx context.Context, packageID uuid.UUID, authToken string) (*packageSnapshot, error) {
	data, err := s.fetchFromService(ctx, s.packageAddr, fmt.Sprintf("/api/v1/packages/%s", packageID), authToken)
	if err != nil {
		return nil, err
	}
	var pkg packageSnapshot
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (s *FinanceService) fetchVendorBillList(ctx context.Context, packageID uuid.UUID, authToken string) ([]vendorBillListItem, error) {
	data, err := s.fetchFromService(ctx, s.vendorAddr, fmt.Sprintf("/api/v1/vendors/bills?package_id=%s&page=1&page_size=500", packageID), authToken)
	if err != nil {
		return nil, err
	}
	var bills []vendorBillListItem
	if err := json.Unmarshal(data, &bills); err == nil {
		return bills, nil
	}
	var wrapped vendorBillList
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return nil, err
	}
	return wrapped.Items, nil
}

func projectedRevenue(pkg *packageSnapshot) (int64, int64) {
	lowest := int64(0)
	for i, tier := range pkg.PricingTiers {
		if i == 0 || tier.Price < lowest {
			lowest = tier.Price
		}
	}
	return lowest * int64(pkg.TotalSeats), lowest
}

func projectedCategoryTotals(pkg *packageSnapshot) (map[string]int64, int64) {
	out := make(map[string]int64)
	var total int64
	for _, component := range pkg.CostComponents {
		key := normalizeProjectedCategory(component.Category)
		perPersonAmount := component.TotalAmount
		if perPersonAmount == 0 {
			perPersonAmount = component.AmountPerPerson * int64(component.Quantity)
		}
		tripAmount := perPersonAmount * int64(pkg.TotalSeats)
		out[key] += tripAmount
		total += tripAmount
	}
	return out, total
}

func sumProjectedPerPerson(pkg *packageSnapshot) int64 {
	var total int64
	for _, component := range pkg.CostComponents {
		perPersonAmount := component.TotalAmount
		if perPersonAmount == 0 {
			perPersonAmount = component.AmountPerPerson * int64(component.Quantity)
		}
		total += perPersonAmount
	}
	return total
}

func expenseCategoryTotals(expenses []model.TripExpense) map[string]int64 {
	out := make(map[string]int64)
	for _, expense := range expenses {
		out[normalizeExpenseCategory(expense)] += expense.AmountIDR
	}
	return out
}

func sumCategoryTotals(categories map[string]int64) int64 {
	var total int64
	for _, amount := range categories {
		total += amount
	}
	return total
}

func buildCostBreakdown(projected, actual map[string]int64) []model.CostBreakdown {
	order := []string{"flight", "hotel_makkah", "hotel_madinah", "visa", "transport", "guide", "equipment", "catering", "other"}
	rows := make([]model.CostBreakdown, 0, len(order))
	for _, category := range order {
		projectedAmount := projected[category]
		actualAmount := actual[category]
		if projectedAmount == 0 && actualAmount == 0 {
			continue
		}
		rows = append(rows, model.CostBreakdown{
			Category:        category,
			Label:           categoryLabel(category),
			ProjectedAmount: projectedAmount,
			ActualAmount:    actualAmount,
			VarianceAmount:  actualAmount - projectedAmount,
		})
	}
	return rows
}

func normalizeProjectedCategory(category string) string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "flight":
		return "flight"
	case "hotel_makkah":
		return "hotel_makkah"
	case "hotel_madinah":
		return "hotel_madinah"
	case "visa":
		return "visa"
	case "transport":
		return "transport"
	case "guide", "guides":
		return "guide"
	case "equipment":
		return "equipment"
	case "catering":
		return "catering"
	default:
		return "other"
	}
}

func normalizeExpenseCategory(expense model.TripExpense) string {
	switch strings.ToLower(strings.TrimSpace(expense.Category)) {
	case "flight":
		return "flight"
	case "hotel_makkah":
		return "hotel_makkah"
	case "hotel_madinah":
		return "hotel_madinah"
	case "accommodation":
		return normalizeHotelCategory(expense.Description)
	case "visa":
		return "visa"
	case "transport":
		return "transport"
	case "guide", "guides":
		return "guide"
	case "equipment":
		return "equipment"
	case "catering", "meals":
		return "catering"
	default:
		return "other"
	}
}

func normalizeVendorBillCategory(bill vendorBillListItem) string {
	if bill.VendorType != nil {
		switch strings.ToLower(strings.TrimSpace(*bill.VendorType)) {
		case "maskapai":
			return "flight"
		case "hotel":
			return normalizeHotelCategory(bill.Description)
		case "transport":
			return "transport"
		case "perlengkapan":
			return "equipment"
		case "katering":
			return "catering"
		default:
			return "other"
		}
	}
	return "other"
}

func normalizeHotelCategory(description string) string {
	desc := strings.ToLower(description)
	switch {
	case strings.Contains(desc, "madinah"):
		return "hotel_madinah"
	case strings.Contains(desc, "makkah"), strings.Contains(desc, "mekkah"), strings.Contains(desc, "mecca"):
		return "hotel_makkah"
	default:
		return "other"
	}
}

func categoryLabel(category string) string {
	switch category {
	case "flight":
		return "Tiket Pesawat"
	case "hotel_makkah":
		return "Hotel Makkah"
	case "hotel_madinah":
		return "Hotel Madinah"
	case "visa":
		return "Visa"
	case "transport":
		return "Transportasi"
	case "guide":
		return "Muthawwif / Guide"
	case "equipment":
		return "Perlengkapan"
	case "catering":
		return "Katering"
	default:
		return "Lain-lain"
	}
}

func marginPercent(profit, revenue int64) float64 {
	if revenue <= 0 {
		return 0
	}
	return float64(profit) * 100 / float64(revenue)
}

// fetchOpenPackages pages through all open packages instead of capping at a
// single page, so the owner dashboard's active_packages and total stay
// consistent for orgs with many open packages. fetchFromService already unwraps
// the {success, data} envelope, so `data` here is the package array itself (the
// total/meta is not returned); the count is derived from pagination.
func (s *FinanceService) fetchOpenPackages(ctx context.Context, authToken string) ([]packageOverviewResponse, int, error) {
	const pageSize = 100
	const maxPages = 50 // safety cap (5000 packages)
	var all []packageOverviewResponse
	for page := 1; page <= maxPages; page++ {
		path := fmt.Sprintf("/api/v1/packages?status=open&page=%d&page_size=%d", page, pageSize)
		data, err := s.fetchFromService(ctx, s.packageAddr, path, authToken)
		if err != nil {
			return nil, 0, err
		}
		var pageItems []packageOverviewResponse
		if err := json.Unmarshal(data, &pageItems); err != nil {
			return nil, 0, err
		}
		all = append(all, pageItems...)
		if len(pageItems) < pageSize {
			break
		}
	}
	return all, len(all), nil
}

func (s *FinanceService) fetchFromService(ctx context.Context, addr, path, authToken string) (json.RawMessage, error) {
	return s.httpc.GetRaw(ctx, addr, path, authToken)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
