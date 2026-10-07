package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/wzantopulos/snipe-po/config"
	"github.com/wzantopulos/snipe-po/db"
)

func GeneratePDF(po *db.PurchaseOrder, items []db.LineItem) (string, error) {
	cfg := config.Get()
	if cfg == nil {
		return "", fmt.Errorf("config not loaded")
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(190, 10, "PURCHASE ORDER")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 10)
	headerY := pdf.GetY()

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(95, 5, "FROM:", 0, 0, "")
	pdf.Cell(95, 5, "TO:", 0, 1, "")
	pdf.SetY(headerY)

	pdf.SetFont("Arial", "", 9)
	pdf.Cell(95, 4, cfg.Company.Name, 0, 0, "")
	pdf.Cell(95, 4, po.Supplier, 0, 1, "")
	pdf.Cell(95, 4, cfg.Company.Address, 0, 0, "")
	pdf.Cell(95, 4, po.SupplierContact, 0, 1, "")
	pdf.Cell(95, 4, cfg.Company.CityStateZip, 0, 0, "")
	pdf.Cell(95, 4, "", 0, 1, "")
	pdf.Cell(95, 4, cfg.Company.Phone, 0, 0, "")
	pdf.Cell(95, 4, "", 0, 1, "")
	pdf.Cell(95, 4, cfg.Company.Email, 0, 1, "")

	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(25, 6, "PO Number:", 0, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(50, 6, po.PONumber, 0, 0, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(25, 6, "Date:", 0, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, po.Date, 0, 1, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(25, 6, "Department:", 0, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(50, 6, po.Department, 0, 0, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(25, 6, "GL Code:", 0, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, po.GLCode, 0, 1, "")

	pdf.Ln(10)

	headers := []string{"Qty", "Description", "Model", "MAC/Serial", "Unit Price", "Total"}
	colWidths := []float64{15, 50, 35, 40, 25, 25}

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(220, 220, 220)
	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 8)
	for _, item := range items {
		description := item.Description
		if len(description) > 40 {
			description = description[:37] + "..."
		}

		model := item.Model
		if len(model) > 25 {
			model = model[:22] + "..."
		}

		identifier := item.MAC
		if identifier == "" {
			identifier = item.Serial
		}
		if len(identifier) > 20 {
			identifier = identifier[:17] + "..."
		}

		pdf.CellFormat(colWidths[0], 7, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[1], 7, description, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[2], 7, model, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[3], 7, identifier, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[4], 7, fmt.Sprintf("$%.2f", item.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[5], 7, fmt.Sprintf("$%.2f", item.Total), "1", 1, "R", false, 0, "")
	}

	pdf.Ln(5)

	totalsX := 140.0
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(totalsX-140, 6, "", 0, 0, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(35, 6, "Subtotal:", 0, 0, "R")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(25, 6, fmt.Sprintf("$%.2f", po.Subtotal), 0, 1, "R")

	pdf.Cell(totalsX-140, 6, "", 0, 0, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(35, 6, "Shipping:", 0, 0, "R")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(25, 6, fmt.Sprintf("$%.2f", po.ShippingCost), 0, 1, "R")

	pdf.Cell(totalsX-140, 6, "", 0, 0, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(35, 6, "Tax:", 0, 0, "R")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(25, 6, fmt.Sprintf("$%.2f", po.TaxCost), 0, 1, "R")

	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(240, 240, 240)
	pdf.Cell(totalsX-140, 8, "", 0, 0, "")
	pdf.Cell(35, 8, "Grand Total:", 0, 0, "R")
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(25, 8, fmt.Sprintf("$%.2f", po.GrandTotal), "1", 1, "R", true, 0, "")

	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(190, 6, "Payment Information", 0, 1, "")
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(30, 5, "Terms:", 0, 0, "")
	pdf.Cell(50, 5, po.Terms, 0, 0, "")
	pdf.Cell(35, 5, "Payment Type:", 0, 0, "")
	pdf.Cell(75, 5, po.PaymentType, 0, 1, "")

	pdf.Ln(15)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(95, 6, "Approved By:", 0, 1, "")
	pdf.Cell(95, 6, "", "B", 1, "")
	pdf.SetFont("Arial", "", 8)
	pdf.Cell(95, 4, "Signature", 0, 0, "")
	pdf.Cell(95, 4, "Date", 0, 1, "")

	pdf.SetY(-30)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 10, fmt.Sprintf("Generated on %s", time.Now().Format("01/02/2006 at 3:04 PM")), 0, 1, "C")

	exePath, err := os.Executable()
	pdfDir := filepath.Join(exePath, "..", "pdfs")
	if err != nil {
		pdfDir = "pdfs"
	}
	os.MkdirAll(pdfDir, 0755)

	safePONumber := strings.ReplaceAll(po.PONumber, "/", "-")
	pdfPath := filepath.Join(pdfDir, fmt.Sprintf("%s.pdf", safePONumber))

	err = pdf.OutputFileAndClose(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to save PDF: %w", err)
	}

	return pdfPath, nil
}
