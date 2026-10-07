package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/wzantopulos/snipe-po/config"
	"github.com/wzantopulos/snipe-po/db"
	"github.com/wzantopulos/snipe-po/pdf"
	"github.com/wzantopulos/snipe-po/snipe"
)

var (
	createFlags struct {
		SnipeITURL   string
		SnipeITToken string
		AssetIDs     string
		Supplier     string
		Terms        string
		PaymentType  string
		Department   string
		GLCode       string
		Shipping     float64
		Tax          float64
		Approver     string
		APEmail      string
		Interactive  bool
	}
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new purchase order",
	Long:  `Create a new purchase order, optionally from Snipe-IT assets.`,
	RunE:  runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVar(&createFlags.SnipeITURL, "snipeit-url", "", "Snipe-IT URL")
	createCmd.Flags().StringVar(&createFlags.SnipeITToken, "snipeit-token", "", "Snipe-IT API token")
	createCmd.Flags().StringVar(&createFlags.AssetIDs, "asset-ids", "", "Comma-separated Snipe-IT asset IDs")
	createCmd.Flags().StringVar(&createFlags.Supplier, "supplier", "", "Supplier name")
	createCmd.Flags().StringVar(&createFlags.Terms, "terms", "", "Payment terms (e.g. Net 30)")
	createCmd.Flags().StringVar(&createFlags.PaymentType, "payment-type", "", "Payment type (Credit Card, ACH, Check)")
	createCmd.Flags().StringVar(&createFlags.Department, "department", "", "Department name")
	createCmd.Flags().StringVar(&createFlags.GLCode, "gl-code", "", "GL code")
	createCmd.Flags().Float64Var(&createFlags.Shipping, "shipping", 0, "Shipping cost")
	createCmd.Flags().Float64Var(&createFlags.Tax, "tax", 0, "Tax cost")
	createCmd.Flags().StringVar(&createFlags.Approver, "approver", "", "Approver email")
	createCmd.Flags().StringVar(&createFlags.APEmail, "ap-email", "", "AP email (after approval)")
	createCmd.Flags().BoolVarP(&createFlags.Interactive, "interactive", "i", false, "Interactive line item entry")
}

func runCreate(cmd *cobra.Command, args []string) error {
	cfg := config.Get()

	supplier := createFlags.Supplier
	if supplier == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Supplier: ")
		supplier, _ = reader.ReadString('\n')
		supplier = strings.TrimSpace(supplier)
	}

	terms := createFlags.Terms
	if terms == "" {
		terms = cfg.PO.DefaultTerms
	}

	paymentType := createFlags.PaymentType
	if paymentType == "" {
		paymentType = cfg.PO.DefaultPaymentType
	}

	department := createFlags.Department
	if department == "" && !createFlags.Interactive {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Department: ")
		department, _ = reader.ReadString('\n')
		department = strings.TrimSpace(department)
	}

	glCode := createFlags.GLCode
	if glCode == "" && !createFlags.Interactive {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("GL Code: ")
		glCode, _ = reader.ReadString('\n')
		glCode = strings.TrimSpace(glCode)
	}

	approver := createFlags.Approver
	if approver == "" && !createFlags.Interactive {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Approver Email: ")
		approver, _ = reader.ReadString('\n')
		approver = strings.TrimSpace(approver)
	}

	apEmail := createFlags.APEmail
	if apEmail == "" && !createFlags.Interactive {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("AP Email (optional): ")
		apEmail, _ = reader.ReadString('\n')
		apEmail = strings.TrimSpace(apEmail)
	}

	var items []db.LineItem

	if createFlags.SnipeITURL != "" && createFlags.SnipeITToken != "" && createFlags.AssetIDs != "" {
		cfg.SnipeIT.URL = createFlags.SnipeITURL
		cfg.SnipeIT.APIKey = createFlags.SnipeITToken

		client := snipe.GetClient()
		ids := strings.Split(createFlags.AssetIDs, ",")
		assets, err := client.GetAssets(ids)
		if err != nil {
			fmt.Printf("Warning: Failed to fetch assets from Snipe-IT: %v\n", err)
		} else {
			for _, asset := range assets {
				modelName := ""
				if asset.Model.Name != "" {
					modelName = asset.Model.Name
				}
				item := db.LineItem{
					AssetID:     strconv.FormatInt(asset.ID, 10),
					Description: asset.Name,
					Model:       modelName,
					MAC:         asset.MACAddress,
					Serial:      asset.Serial,
					Quantity:    1,
					UnitPrice:   0,
					Total:       0,
					Department:  department,
					GLCode:      glCode,
				}
				items = append(items, item)
			}
		}
	}

	if createFlags.Interactive {
		items = collectInteractiveItems(department, glCode)
	}

	if len(items) == 0 && !createFlags.Interactive {
		fmt.Println("No assets specified and not in interactive mode. Adding line items manually...")
		items = collectInteractiveItems(department, glCode)
	}

	var subtotal float64
	for i := range items {
		items[i].Total = float64(items[i].Quantity) * items[i].UnitPrice
		subtotal += items[i].Total
	}

	grandTotal := subtotal + createFlags.Shipping + createFlags.Tax

	poNumber, err := db.GetNextPONumber(cfg.PO.NumberPrefix)
	if err != nil {
		return fmt.Errorf("failed to generate PO number: %w", err)
	}

	now := time.Now()
	po := &db.PurchaseOrder{
		ID:              uuid.New().String(),
		PONumber:       poNumber,
		CreatedAt:       now,
		UpdatedAt:       now,
		Status:          "draft",
		Date:            now.Format("01/02/2006"),
		Supplier:        supplier,
		Terms:           terms,
		PaymentType:     paymentType,
		Department:      department,
		GLCode:          glCode,
		Subtotal:        subtotal,
		ShippingCost:    createFlags.Shipping,
		TaxCost:         createFlags.Tax,
		GrandTotal:      grandTotal,
		ApproverEmail:   approver,
		APEmail:         apEmail,
	}

	if err := db.CreatePO(po); err != nil {
		return fmt.Errorf("failed to create PO: %w", err)
	}

	for i := range items {
		items[i].POID = po.ID
		if err := db.CreateLineItem(&items[i]); err != nil {
			return fmt.Errorf("failed to create line item: %w", err)
		}
	}

	pdfPath, err := pdf.GeneratePDF(po, items)
	if err != nil {
		fmt.Printf("Warning: Failed to generate PDF: %v\n", err)
	} else {
		po.PDFPath = pdfPath
		db.UpdatePO(po)
	}

	db.AddHistory(po.ID, "created", fmt.Sprintf("PO created with %d line items", len(items)))

	fmt.Printf("Purchase Order Created: %s\n", poNumber)
	fmt.Printf("ID: %s\n", po.ID)
	fmt.Printf("Supplier: %s\n", supplier)
	fmt.Printf("Grand Total: $%.2f\n", grandTotal)
	if pdfPath != "" {
		fmt.Printf("PDF: %s\n", pdfPath)
	}

	return nil
}

func collectInteractiveItems(department, glCode string) []db.LineItem {
	var items []db.LineItem
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n--- Line Items (enter empty description to finish) ---")

	for {
		fmt.Print("Description: ")
		desc, _ := reader.ReadString('\n')
		desc = strings.TrimSpace(desc)
		if desc == "" {
			break
		}

		fmt.Print("Model: ")
		model, _ := reader.ReadString('\n')
		model = strings.TrimSpace(model)

		fmt.Print("MAC: ")
		mac, _ := reader.ReadString('\n')
		mac = strings.TrimSpace(mac)

		fmt.Print("Serial: ")
		serial, _ := reader.ReadString('\n')
		serial = strings.TrimSpace(serial)

		fmt.Print("Quantity: ")
		qtyStr, _ := reader.ReadString('\n')
		qty, _ := strconv.Atoi(strings.TrimSpace(qtyStr))
		if qty <= 0 {
			qty = 1
		}

		fmt.Print("Unit Price: ")
		priceStr, _ := reader.ReadString('\n')
		price, _ := strconv.ParseFloat(strings.TrimSpace(priceStr), 64)

		fmt.Print("Department (leave empty for default): ")
		itemDept, _ := reader.ReadString('\n')
		itemDept = strings.TrimSpace(itemDept)
		if itemDept == "" {
			itemDept = department
		}

		fmt.Print("GL Code (leave empty for default): ")
		itemGL, _ := reader.ReadString('\n')
		itemGL = strings.TrimSpace(itemGL)
		if itemGL == "" {
			itemGL = glCode
		}

		items = append(items, db.LineItem{
			Description: desc,
			Model:       model,
			MAC:         mac,
			Serial:      serial,
			Quantity:    qty,
			UnitPrice:   price,
			Total:       float64(qty) * price,
			Department:  itemDept,
			GLCode:      itemGL,
		})

		fmt.Println()
	}

	return items
}
