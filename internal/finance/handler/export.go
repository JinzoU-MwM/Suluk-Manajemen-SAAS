package handler

import (
	"bytes"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/jamaah-in/v2/internal/finance/model"
	sharedMW "github.com/jamaah-in/v2/internal/shared/middleware"
	"github.com/jamaah-in/v2/internal/shared/response"
)

func (h *FinanceHandler) ExportPnL(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}
	packageID := c.Query("package_id", "")

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	_ = f.SetCellValue(sheet, "A1", "Laporan P&L — Suluk Travel Management")
	_ = f.SetCellValue(sheet, "A2", fmt.Sprintf("Tanggal: %s", time.Now().Format("02 Jan 2006 15:04")))

	// Headers
	headers := []string{"Kategori", "Deskripsi", "Jumlah (IDR)"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		_ = f.SetCellValue(sheet, cell, h)
	}

	_ = f.SetColWidth(sheet, "A", "A", 20)
	_ = f.SetColWidth(sheet, "B", "B", 45)
	_ = f.SetColWidth(sheet, "C", "C", 22)

	titleStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14}})
	_ = f.SetCellStyle(sheet, "A1", "A1", titleStyle)
	_ = f.MergeCell(sheet, "A1", "C1")

	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"1E40AF"}, Pattern: 1}})
	_ = f.SetCellStyle(sheet, "A4", "C4", headerStyle)

	// Get data — use the package-scoped query when filtering so we don't post-filter
	// after a 500-row org-wide cap (which can silently drop a package's later rows).
	var expenses []model.TripExpense
	if packageID != "" {
		pkgUUID, err := uuid.Parse(packageID)
		if err != nil {
			return response.BadRequest(c, "invalid package_id")
		}
		expenses, err = h.svc.ListExpensesByPackage(c.Context(), claims.OrgID, pkgUUID)
		if err != nil {
			return response.Internal(c, err)
		}
	} else {
		var err error
		expenses, _, err = h.svc.ListExpenses(c.Context(), claims.OrgID, "", "", 1, 500)
		if err != nil {
			return response.Internal(c, err)
		}
	}

	row := 5
	totalExp := int64(0)

	for _, expense := range expenses {
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), expense.Category)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), expense.Description)
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), expense.AmountIDR)
		totalExp += expense.AmountIDR
		row++
	}

	row++
	_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "TOTAL")
	_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "Biaya Operasional")
	_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), totalExp)
	totalStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 12}, Fill: excelize.Fill{Type: "pattern", Color: []string{"DBEAFE"}, Pattern: 1}})
	_ = f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), totalStyle)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return response.Internal(c, err)
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="pnl_report.xlsx"`)
	return c.Send(buf.Bytes())
}

func (h *FinanceHandler) ExportExpenses(c *fiber.Ctx) error {
	claims, ok := sharedMW.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "unauthorized")
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	_ = f.SetCellValue(sheet, "A1", "Laporan Biaya Operasional")
	_ = f.SetCellValue(sheet, "A2", fmt.Sprintf("Export: %s", time.Now().Format("02 Jan 2006 15:04")))

	headers := []string{"No", "Tanggal", "Kategori", "Deskripsi", "Jumlah"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		_ = f.SetCellValue(sheet, cell, h)
	}

	_ = f.SetColWidth(sheet, "A", "A", 6)
	_ = f.SetColWidth(sheet, "B", "B", 14)
	_ = f.SetColWidth(sheet, "C", "C", 16)
	_ = f.SetColWidth(sheet, "D", "D", 40)
	_ = f.SetColWidth(sheet, "E", "E", 20)

	titleStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 13}})
	_ = f.SetCellStyle(sheet, "A1", "A1", titleStyle)

	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"E2E8F0"}, Pattern: 1}})
	_ = f.SetCellStyle(sheet, "A4", "E4", headerStyle)

	expenses, _, err := h.svc.ListExpenses(c.Context(), claims.OrgID, "", "", 1, 500)
	if err != nil {
		return response.Internal(c, err)
	}

	row := 5
	total := int64(0)
	for _, exp := range expenses {
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), row-4)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), exp.CreatedAt.Format("02/01/2006"))
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), exp.Category)
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), exp.Description)
		_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", row), exp.AmountIDR)
		total += exp.AmountIDR
		row++
	}

	row++
	_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), "TOTAL")
	_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", row), total)
	boldStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	_ = f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("E%d", row), boldStyle)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return response.Internal(c, err)
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="expenses.xlsx"`)
	return c.Send(buf.Bytes())
}
