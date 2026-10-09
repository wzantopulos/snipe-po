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

	approverName := ""
	if po.ApproverEmail != "" {
		// Extract name from email (e.g., john.smith@company.com -> John Smith)
		parts := strings.Split(po.ApproverEmail, "@")
		if len(parts) > 0 {
			name := strings.ReplaceAll(parts[0], ".", " ")
			name = strings.ReplaceAll(name, "_", " ")
			// Title case each word
			words := strings.Fields(name)
			for i, word := range words {
				if len(word) > 0 {
					words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
				}
			}
			approverName = strings.Join(words, " ")
		}
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(190, 10, "PURCHASE ORDER")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 10)
	headerY := pdf.GetY()

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(95, 5, "FROM:")
	pdf.Cell(95, 5, "TO:")
	pdf.SetY(headerY)

	pdf.SetFont("Arial", "", 9)
	pdf.Cell(95, 4, cfg.Company.Name)
	pdf.Cell(95, 4, po.Supplier)
	pdf.Ln(4)
	pdf.Cell(95, 4, cfg.Company.Address)
	pdf.Cell(95, 4, po.SupplierContact)
	pdf.Ln(4)
	pdf.Cell(95, 4, cfg.Company.CityStateZip)
	pdf.Ln(4)
	pdf.Cell(95, 4, cfg.Company.Phone)
	pdf.Ln(4)
	pdf.Cell(95, 4, cfg.Company.Email)
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(25, 6, "PO Number:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(50, 6, po.PONumber)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(25, 6, "Date:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, po.Date)
	pdf.Ln(6)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(25, 6, "Department:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(50, 6, po.Department)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(25, 6, "GL Code:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, po.GLCode)
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
	pdf.SetFillColor(255, 255, 255)
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
	pdf.Cell(totalsX-140, 6, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(35, 6, "Subtotal:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(25, 6, fmt.Sprintf("$%.2f", po.Subtotal))
	pdf.Ln(6)

	pdf.Cell(totalsX-140, 6, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(35, 6, "Shipping:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(25, 6, fmt.Sprintf("$%.2f", po.ShippingCost))
	pdf.Ln(6)

	pdf.Cell(totalsX-140, 6, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(35, 6, "Tax:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(25, 6, fmt.Sprintf("$%.2f", po.TaxCost))
	pdf.Ln(6)

	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(240, 240, 240)
	pdf.Cell(totalsX-140, 8, "")
	pdf.Cell(35, 8, "Grand Total:")
	pdf.Cell(25, 8, fmt.Sprintf("$%.2f", po.GrandTotal))
	pdf.Ln(8)

	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(190, 6, "Payment Information")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(30, 5, "Terms:")
	pdf.Cell(50, 5, po.Terms)
	pdf.Cell(35, 5, "Payment Type:")
	pdf.Cell(75, 5, po.PaymentType)
	pdf.Ln(15)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(190, 6, "Approved By:")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 8)
	pdf.Cell(95, 4, "Signature: ________________________")
	pdf.Ln(5)
	if approverName != "" {
		pdf.Cell(95, 4, "Approved by: "+approverName)
	} else {
		pdf.Cell(95, 4, "Approved by: ")
	}
	pdf.Ln(5)
	if po.ApprovedAt != nil {
		pdf.Cell(95, 4, "Date: "+po.ApprovedAt.Format("01/02/2006"))
	}

	pdf.SetY(-30)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 10, fmt.Sprintf("Generated on %s", time.Now().Format("01/02/2006 at 3:04 PM")))

	pdfDir := "/data/pdfs"
	os.MkdirAll(pdfDir, 0755)

	safePONumber := strings.ReplaceAll(po.PONumber, "/", "-")
	pdfPath := filepath.Join(pdfDir, fmt.Sprintf("%s.pdf", safePONumber))

	err := pdf.OutputFileAndClose(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to save PDF: %w", err)
	}

	return pdfPath, nil
}
