package handler

import (
	"bytes"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"

	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
)

func (h *FinanceHandler) ExportPnL(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)
	packageID := c.Query("package_id", "")

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	f.SetCellValue(sheet, "A1", "Laporan P&L — Suluk Travel Management")
	f.SetCellValue(sheet, "A2", fmt.Sprintf("Tanggal: %s", time.Now().Format("02 Jan 2006 15:04")))

	// Headers
	headers := []string{"Kategori", "Deskripsi", "Jumlah (IDR)"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		f.SetCellValue(sheet, cell, h)
	}

	f.SetColWidth(sheet, "A", "A", 20)
	f.SetColWidth(sheet, "B", "B", 45)
	f.SetColWidth(sheet, "C", "C", 22)

	titleStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14}})
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)
	f.MergeCell(sheet, "A1", "C1")

	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"1E40AF"}, Pattern: 1}})
	f.SetCellStyle(sheet, "A4", "C4", headerStyle)

	// Get data
	expenses, _, _ := h.svc.ListExpenses(c.Context(), claims.OrgID, "", "", 1, 500)

	row := 5
	totalExp := int64(0)

	for _, expense := range expenses {
		if packageID != "" && expense.PackageID.String() != packageID {
			continue
		}
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), expense.Category)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), expense.Description)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), expense.Amount)
		totalExp += expense.Amount
		row++
	}

	row++
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "TOTAL")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "Biaya Operasional")
	f.SetCellValue(sheet, fmt.Sprintf("C%d", row), totalExp)
	totalStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 12}, Fill: excelize.Fill{Type: "pattern", Color: []string{"DBEAFE"}, Pattern: 1}})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), totalStyle)

	var buf bytes.Buffer
	f.Write(&buf)
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="pnl_report.xlsx"`)
	return c.Send(buf.Bytes())
}

func (h *FinanceHandler) ExportExpenses(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*sharedAuth.Claims)

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	f.SetCellValue(sheet, "A1", "Laporan Biaya Operasional")
	f.SetCellValue(sheet, "A2", fmt.Sprintf("Export: %s", time.Now().Format("02 Jan 2006 15:04")))

	headers := []string{"No", "Tanggal", "Kategori", "Deskripsi", "Jumlah"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		f.SetCellValue(sheet, cell, h)
	}

	f.SetColWidth(sheet, "A", "A", 6)
	f.SetColWidth(sheet, "B", "B", 14)
	f.SetColWidth(sheet, "C", "C", 16)
	f.SetColWidth(sheet, "D", "D", 40)
	f.SetColWidth(sheet, "E", "E", 20)

	titleStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 13}})
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)

	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"E2E8F0"}, Pattern: 1}})
	f.SetCellStyle(sheet, "A4", "E4", headerStyle)

	expenses, _, _ := h.svc.ListExpenses(c.Context(), claims.OrgID, "", "", 1, 500)

	row := 5
	total := int64(0)
	for _, exp := range expenses {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), row-4)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), exp.CreatedAt.Format("02/01/2006"))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), exp.Category)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), exp.Description)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), exp.Amount)
		total += exp.Amount
		row++
	}

	row++
	f.SetCellValue(sheet, fmt.Sprintf("D%d", row), "TOTAL")
	f.SetCellValue(sheet, fmt.Sprintf("E%d", row), total)
	boldStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("E%d", row), boldStyle)

	var buf bytes.Buffer
	f.Write(&buf)
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="expenses.xlsx"`)
	return c.Send(buf.Bytes())
}
