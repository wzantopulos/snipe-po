package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wzantopulos/snipe-po/db"
)

var (
	listFlags struct {
		Status string
		Format string
	}
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all purchase orders",
	Long:  `Show all purchase orders with their current status.`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVar(&listFlags.Status, "status", "", "Filter by status (draft, pending_approval, approved, sent_to_ap, paid, rejected)")
	listCmd.Flags().StringVarP(&listFlags.Format, "format", "f", "table", "Output format: table or json")
}

func runList(cmd *cobra.Command, args []string) error {
	pos, err := db.GetAllPOs(listFlags.Status)
	if err != nil {
		return fmt.Errorf("failed to fetch POs: %w", err)
	}

	if len(pos) == 0 {
		fmt.Println("No purchase orders found.")
		return nil
	}

	if listFlags.Format == "json" {
		return printJSON(pos)
	}

	return printTable(pos)
}

func printTable(pos []db.PurchaseOrder) error {
	fmt.Printf("%-15s %-12s %-20s %-15s %-12s\n", "PO NUMBER", "STATUS", "SUPPLIER", "DATE", "GRAND TOTAL")
	fmt.Println(strings.Repeat("-", 74))

	for _, po := range pos {
		statusColor := po.Status
		switch po.Status {
		case "draft":
			statusColor = "draft"
		case "pending_approval":
			statusColor = "pending"
		case "approved":
			statusColor = "approved"
		case "sent_to_ap":
			statusColor = "sent_to_ap"
		case "paid":
			statusColor = "paid"
		case "rejected":
			statusColor = "rejected"
		}

		fmt.Printf("%-15s %-12s %-20s %-15s $%-11.2f\n",
			po.PONumber,
			statusColor,
			truncate(po.Supplier, 19),
			po.Date,
			po.GrandTotal,
		)
	}

	fmt.Println()
	fmt.Printf("Total: %d purchase order(s)\n", len(pos))
	return nil
}

func printJSON(pos []db.PurchaseOrder) error {
	type poJSON struct {
		ID          string    `json:"id"`
		PONumber   string    `json:"po_number"`
		Status      string    `json:"status"`
		Supplier    string    `json:"supplier"`
		Date        string    `json:"date"`
		Department  string    `json:"department"`
		GrandTotal  float64   `json:"grand_total"`
		CreatedAt   time.Time `json:"created_at"`
		ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	}

	var jsonPOs []poJSON
	for _, po := range pos {
		jsonPOs = append(jsonPOs, poJSON{
			ID:         po.ID,
			PONumber:   po.PONumber,
			Status:     po.Status,
			Supplier:   po.Supplier,
			Date:       po.Date,
			Department: po.Department,
			GrandTotal: po.GrandTotal,
			CreatedAt:  po.CreatedAt,
			ApprovedAt: po.ApprovedAt,
		})
	}

	data, err := json.MarshalIndent(jsonPOs, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
