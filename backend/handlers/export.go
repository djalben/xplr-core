package handlers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/models"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

// txTypeLabelExport maps transaction_type to a human-readable Russian label for export.
func txTypeLabelExport(txType string) string {
	switch txType {
	case "WALLET_TOPUP":
		return "Пополнение"
	case "CARD_ISSUE_FEE":
		return "Выпуск карты"
	case "CARD_TOPUP":
		return "Перевод на карту"
	case "CARD_REFUND":
		return "Возврат"
	case "WALLET_RECLAIM":
		return "Возврат (истёкшая)"
	case "CARD_CHARGE":
		return "Списание"
	case "REFERRAL_BONUS":
		return "Реф. бонус"
	case "COMMISSION":
		return "Комиссия"
	default:
		return txType
	}
}

// statusLabelExport maps status to Russian.
func statusLabelExport(status string) string {
	switch status {
	case "SUCCESS", "COMPLETED":
		return "Успешно"
	case "CAPTURED", "CAPTURE":
		return "Захвачено"
	case "APPROVED":
		return "Одобрено"
	case "PENDING":
		return "В обработке"
	case "FAILED", "DECLINED":
		return "Отклонено"
	case "REVERSED":
		return "Возврат"
	default:
		return status
	}
}

// ExportTransactionsHandler — GET /api/v1/user/transactions/export?format={pdf|excel}&start_date=&end_date=
func ExportTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	format := r.URL.Query().Get("format")
	if format != "pdf" && format != "excel" {
		http.Error(w, "Invalid format. Use ?format=pdf or ?format=excel", http.StatusBadRequest)
		return
	}

	// Build filters
	filters := make(map[string]interface{})
	if sd := r.URL.Query().Get("start_date"); sd != "" {
		filters["start_date"] = sd
	}
	if ed := r.URL.Query().Get("end_date"); ed != "" {
		filters["end_date"] = ed
	}
	filters["limit"] = 500

	txs, _, err := repository.GetUnifiedTransactions(userID, filters)
	if err != nil {
		log.Printf("[EXPORT] Error fetching transactions for user %d: %v", userID, err)
		http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
		return
	}

	if format == "excel" {
		generateExcel(w, txs)
	} else {
		generatePDF(w, txs)
	}
}

// ──────────────────────── EXCEL ────────────────────────

func generateExcel(w http.ResponseWriter, txs []models.Transaction) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Транзакции"
	idx, _ := f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(idx)

	// Column widths
	f.SetColWidth(sheet, "A", "A", 20)
	f.SetColWidth(sheet, "B", "B", 22)
	f.SetColWidth(sheet, "C", "C", 35)
	f.SetColWidth(sheet, "D", "D", 16)
	f.SetColWidth(sheet, "E", "E", 16)

	// Header style (bold, centered, blue background)
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"3B82F6"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "bottom", Color: "1D4ED8", Style: 2},
		},
	})

	// Amount style (right-aligned, number format)
	amountStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt:    4, // #,##0.00
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})

	// Date style
	dateStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	// Status styles
	successStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "16A34A"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	failStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "DC2626"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	pendingStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "D97706"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	// Headers
	headers := []string{"Дата", "Тип операции", "Описание", "Сумма ($)", "Статус"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// Data rows
	var totalSum float64
	for i, tx := range txs {
		row := i + 2
		amt, _ := tx.Amount.Float64()

		// Date
		dateCell, _ := excelize.CoordinatesToCellName(1, row)
		f.SetCellValue(sheet, dateCell, tx.ExecutedAt.Format("02.01.2006 15:04"))
		f.SetCellStyle(sheet, dateCell, dateCell, dateStyle)

		// Type
		typeCell, _ := excelize.CoordinatesToCellName(2, row)
		f.SetCellValue(sheet, typeCell, txTypeLabelExport(tx.TransactionType))

		// Description (card or wallet)
		descCell, _ := excelize.CoordinatesToCellName(3, row)
		desc := tx.Details
		if desc == "" {
			if tx.CardLast4Digits != "" {
				desc = "Карта •••• " + tx.CardLast4Digits
			} else {
				desc = "Кошелёк"
			}
		}
		f.SetCellValue(sheet, descCell, desc)

		// Amount
		amtCell, _ := excelize.CoordinatesToCellName(4, row)
		f.SetCellValue(sheet, amtCell, math.Round(amt*100)/100)
		f.SetCellStyle(sheet, amtCell, amtCell, amountStyle)
		totalSum += amt

		// Status
		statusCell, _ := excelize.CoordinatesToCellName(5, row)
		statusLabel := statusLabelExport(tx.Status)
		f.SetCellValue(sheet, statusCell, statusLabel)
		switch tx.Status {
		case "SUCCESS", "COMPLETED", "CAPTURED", "CAPTURE", "APPROVED":
			f.SetCellStyle(sheet, statusCell, statusCell, successStyle)
		case "FAILED", "DECLINED":
			f.SetCellStyle(sheet, statusCell, statusCell, failStyle)
		default:
			f.SetCellStyle(sheet, statusCell, statusCell, pendingStyle)
		}
	}

	// Total row
	totalRow := len(txs) + 2
	totalLabelCell, _ := excelize.CoordinatesToCellName(3, totalRow)
	totalAmtCell, _ := excelize.CoordinatesToCellName(4, totalRow)

	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})
	totalAmtStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11},
		NumFmt:    4,
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})

	f.SetCellValue(sheet, totalLabelCell, "Итого:")
	f.SetCellStyle(sheet, totalLabelCell, totalLabelCell, totalStyle)
	f.SetCellValue(sheet, totalAmtCell, math.Round(totalSum*100)/100)
	f.SetCellStyle(sheet, totalAmtCell, totalAmtCell, totalAmtStyle)

	// Auto-filter
	lastDataRow := len(txs) + 1
	if lastDataRow < 2 {
		lastDataRow = 2
	}
	lastCell, _ := excelize.CoordinatesToCellName(5, lastDataRow)
	f.AutoFilter(sheet, "A1:"+lastCell, nil)

	// Write to response
	now := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("XPLR_transactions_%s.xlsx", now)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	if err := f.Write(w); err != nil {
		log.Printf("[EXPORT-EXCEL] Error writing file: %v", err)
	}
}

// ──────────────────────── PDF ────────────────────────

func generatePDF(w http.ResponseWriter, txs []models.Transaction) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)

	// Use built-in Helvetica (ASCII only) + register UTF-8 font for Cyrillic
	// gofpdf ships with cp1252 encoding maps; for Cyrillic we use its built-in translation
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1251")

	pdf.AddPage()

	// ── Header ──
	pdf.SetFillColor(30, 41, 59) // slate-800
	pdf.Rect(0, 0, 297, 28, "F")
	pdf.SetFont("Helvetica", "B", 20)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(15, 6)
	pdf.CellFormat(60, 16, "XPLR", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetXY(200, 10)
	pdf.CellFormat(85, 8, tr("Отчёт по транзакциям"), "", 0, "R", false, 0, "")

	// Date range subtitle
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(148, 163, 184) // slate-400
	pdf.SetXY(15, 32)
	dateLabel := tr(fmt.Sprintf("Дата формирования: %s", time.Now().Format("02.01.2006 15:04")))
	pdf.CellFormat(0, 6, dateLabel, "", 1, "L", false, 0, "")

	// Count
	pdf.SetXY(15, 38)
	pdf.CellFormat(0, 6, tr(fmt.Sprintf("Всего операций: %d", len(txs))), "", 1, "L", false, 0, "")

	pdf.Ln(4)

	// ── Table Header ──
	colWidths := []float64{40, 40, 95, 35, 28, 30}
	headerLabels := []string{"Дата", "Тип", "Описание", "Сумма ($)", "Статус", "Карта"}

	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(59, 130, 246) // blue-500
	pdf.SetTextColor(255, 255, 255)
	for i, label := range headerLabels {
		pdf.CellFormat(colWidths[i], 8, tr(label), "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// ── Table Rows ──
	pdf.SetFont("Helvetica", "", 8)
	var totalSum float64
	for idx, tx := range txs {
		// Alternating row colors
		if idx%2 == 0 {
			pdf.SetFillColor(241, 245, 249) // slate-100
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		amt, _ := tx.Amount.Float64()
		totalSum += amt

		// Date
		pdf.SetTextColor(30, 41, 59)
		pdf.CellFormat(colWidths[0], 7, tx.ExecutedAt.Format("02.01.2006 15:04"), "1", 0, "C", true, 0, "")

		// Type
		pdf.CellFormat(colWidths[1], 7, tr(txTypeLabelExport(tx.TransactionType)), "1", 0, "C", true, 0, "")

		// Description
		desc := tx.Details
		if desc == "" {
			if tx.CardLast4Digits != "" {
				desc = "Карта •••• " + tx.CardLast4Digits
			} else {
				desc = "Кошелёк"
			}
		}
		// Truncate long descriptions
		if len(desc) > 55 {
			desc = desc[:52] + "..."
		}
		pdf.CellFormat(colWidths[2], 7, tr(desc), "1", 0, "L", true, 0, "")

		// Amount — color based on sign
		if amt >= 0 {
			pdf.SetTextColor(22, 163, 74) // green
		} else {
			pdf.SetTextColor(220, 38, 38) // red
		}
		amtStr := strconv.FormatFloat(math.Round(amt*100)/100, 'f', 2, 64)
		pdf.CellFormat(colWidths[3], 7, amtStr, "1", 0, "R", true, 0, "")

		// Status
		pdf.SetTextColor(30, 41, 59)
		pdf.CellFormat(colWidths[4], 7, tr(statusLabelExport(tx.Status)), "1", 0, "C", true, 0, "")

		// Card
		card := ""
		if tx.CardLast4Digits != "" {
			card = "*" + tx.CardLast4Digits
		}
		pdf.CellFormat(colWidths[5], 7, card, "1", 0, "C", true, 0, "")

		pdf.Ln(-1)
	}

	// ── Total Row ──
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(226, 232, 240) // slate-200
	pdf.SetTextColor(30, 41, 59)
	sumWidth := colWidths[0] + colWidths[1] + colWidths[2]
	pdf.CellFormat(sumWidth, 8, tr("Итого:"), "1", 0, "R", true, 0, "")
	totalStr := strconv.FormatFloat(math.Round(totalSum*100)/100, 'f', 2, 64)
	pdf.CellFormat(colWidths[3], 8, totalStr, "1", 0, "R", true, 0, "")
	remainWidth := colWidths[4] + colWidths[5]
	pdf.CellFormat(remainWidth, 8, "", "1", 0, "", true, 0, "")
	pdf.Ln(-1)

	// ── Footer ──
	pdf.Ln(8)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(148, 163, 184)
	pdf.CellFormat(0, 5, tr("XPLR — Управление виртуальными картами"), "", 1, "C", false, 0, "")

	// Write to response
	now := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("XPLR_transactions_%s.pdf", now)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	if err := pdf.Output(w); err != nil {
		log.Printf("[EXPORT-PDF] Error writing file: %v", err)
	}
}
