package invoice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"

	"github.com/bhanuprakaash/job-scheduler/internal/blob"
	"github.com/bhanuprakaash/job-scheduler/internal/logger"
	"github.com/bhanuprakaash/job-scheduler/internal/store"
)

type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type InvoiceJobPayload struct {
	UserId      string        `json:"user_id"`
	InvoiceId   string        `json:"invoice_id"`
	Date        string        `json:"date"`
	Currency    string        `json:"currency"`
	Items       []InvoiceItem `json:"items"`
	TotalAmount float32       `json:"amount"`
}

type InvoiceJob struct {
	storageClient blob.StorageClient
}

func NewInvoiceJob(storageClient blob.StorageClient) *InvoiceJob {
	return &InvoiceJob{
		storageClient: storageClient,
	}
}

func (e *InvoiceJob) Handle(ctx context.Context, job store.Job) error {
	var payload InvoiceJobPayload
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		return fmt.Errorf("parse invoice payload: %w", err)
	}
	startTime := time.Now()

	objectName := fmt.Sprintf("secure/invoices/%s.pdf", payload.InvoiceId)

	// check: invoice exists
	exists, err := e.storageClient.Exists(ctx, objectName)
	if err != nil {
		return fmt.Errorf("error checking object: %w", err)
	}
	if exists {
		logger.Info("Invoice already exists, skipping", "invoice", payload.InvoiceId)
		return nil
	}

	// generate pdf
	var buf bytes.Buffer
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	generateInvoicePdf(pdf, payload)

	if err := pdf.Output(&buf); err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	// upload to storage client
	if err = e.storageClient.Upload(ctx, &buf, int64(buf.Len()), objectName, "application/pdf"); err != nil {
		return fmt.Errorf("upload pdf: %w", err)
	}

	duration := time.Since(startTime)

	logger.Info("Invoice Generated Successfully", "invoice id", payload.InvoiceId, "duration", duration)

	return nil
}

func generateInvoicePdf(pdf *fpdf.Fpdf, payload InvoiceJobPayload) {
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(50, 50, 150)
	pdf.Cell(0, 10, "DUNDER MUFFLIN")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(0, 5, "Dunder Mufflin Paper Company")
	pdf.Ln(5)
	pdf.Cell(0, 5, "Scranton, Pennsylvania - 18503")
	pdf.Ln(10)

	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(10)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 12)

	yBefore := pdf.GetY()

	pdf.Cell(95, 7, "BILL TO:")
	pdf.Ln(7)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(95, 6, fmt.Sprintf("User ID: %s", payload.UserId))
	pdf.Ln(6)
	pdf.Cell(95, 6, "Customer Support: dunder_mufflin@scheduler.com")

	pdf.SetY(yBefore)
	pdf.SetX(115)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 7, "INVOICE DETAILS:")
	pdf.Ln(7)
	pdf.SetX(115)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 6, fmt.Sprintf("Invoice ID: %s", payload.InvoiceId))
	pdf.Ln(6)
	pdf.SetX(115)
	pdf.Cell(0, 6, fmt.Sprintf("Date: %s", payload.Date))

	pdf.SetY(yBefore + 35)

	drawInvoiceTable(pdf, payload.Items, payload.Currency, float64(payload.TotalAmount))

	pdf.SetY(-30)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(0, 10, "Thank you for your business. For any queries, contact billing@scheduler.com", "", 0, "C", false, 0, "")
}

func drawInvoiceTable(pdf *fpdf.Fpdf, items []InvoiceItem, currency string, grandTotal float64) {
	w := []float64{90, 20, 40, 40}

	// Header
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(w[0], 10, "Description", "1", 0, "L", true, 0, "")
	pdf.CellFormat(w[1], 10, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(w[2], 10, fmt.Sprintf("Price (%s)", currency), "1", 0, "R", true, 0, "")
	pdf.CellFormat(w[3], 10, fmt.Sprintf("Total (%s)", currency), "1", 0, "R", true, 0, "")
	pdf.Ln(-1)

	// Body
	pdf.SetFont("Arial", "", 10)
	for _, item := range items {
		itemTotal := float64(item.Quantity) * item.UnitPrice
		pdf.CellFormat(w[0], 8, item.Description, "1", 0, "L", false, 0, "")
		pdf.CellFormat(w[1], 8, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(w[2], 8, fmt.Sprintf("%.2f", item.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(w[3], 8, fmt.Sprintf("%.2f", itemTotal), "1", 0, "R", false, 0, "")
		pdf.Ln(-1)

		if pdf.GetY() > 270 {
			pdf.AddPage()
		}
	}

	// Grand Total Row
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(w[0]+w[1]+w[2], 10, "Grand Total", "1", 0, "R", false, 0, "")
	pdf.CellFormat(w[3], 10, fmt.Sprintf("%.2f %s", grandTotal, currency), "1", 0, "R", false, 0, "")
}
