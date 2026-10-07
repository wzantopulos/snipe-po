package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wzantopulos/snipe-po/db"
	"github.com/wzantopulos/snipe-po/email"
)

var (
	approveFlags struct {
		ID   string
		Note string
	}
)

var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve a purchase order",
	Long:  `Mark a purchase order as approved and optionally forward to AP.`,
	RunE:  runApprove,
}

func init() {
	rootCmd.AddCommand(approveCmd)

	approveCmd.Flags().StringVarP(&approveFlags.ID, "id", "", "", "PO ID or number")
	approveCmd.Flags().StringVarP(&approveFlags.Note, "note", "", "", "Optional approval note")
}

func runApprove(cmd *cobra.Command, args []string) error {
	if approveFlags.ID == "" {
		return fmt.Errorf("PO ID or number is required (--id)")
	}

	po, err := db.GetPO(approveFlags.ID)
	if err != nil {
		return fmt.Errorf("PO not found: %w", err)
	}

	if po.Status != "draft" && po.Status != "pending_approval" {
		return fmt.Errorf("PO cannot be approved from status: %s", po.Status)
	}

	now := time.Now()
	po.Status = "approved"
	po.ApprovedAt = &now
	po.ApprovalNote = approveFlags.Note

	if err := db.UpdatePO(po); err != nil {
		return fmt.Errorf("failed to update PO: %w", err)
	}

	noteStr := ""
	if approveFlags.Note != "" {
		noteStr = fmt.Sprintf("Note: %s", approveFlags.Note)
	}
	db.AddHistory(po.ID, "approved", noteStr)

	fmt.Printf("Purchase Order %s approved\n", po.PONumber)

	if po.APEmail != "" && po.PDFPath != "" {
		if _, err := os.Stat(po.PDFPath); err == nil {
			recipients := strings.Split(po.APEmail, ",")
			for i := range recipients {
				recipients[i] = strings.TrimSpace(recipients[i])
			}

			subject := fmt.Sprintf("Approved Purchase Order %s - Forwarded to AP", po.PONumber)
			body := fmt.Sprintf(`Purchase Order %s has been approved and is being forwarded to Accounts Payable.

Supplier: %s
Date: %s
Grand Total: $%.2f
Approved: %s

%s

Please find the attached PDF for processing.

Thank you,
Purchase Order System`, po.PONumber, po.Supplier, po.Date, po.GrandTotal, now.Format("01/02/2006"), noteStr)

			if err := email.SendEmail(recipients, subject, body, po.PDFPath); err != nil {
				fmt.Printf("Warning: Failed to forward to AP: %v\n", err)
			} else {
				po.Status = "sent_to_ap"
				db.UpdatePO(po)
				db.AddHistory(po.ID, "forwarded_to_ap", fmt.Sprintf("Forwarded to %s", po.APEmail))
				fmt.Printf("Forwarded to AP: %s\n", po.APEmail)
			}
		}
	}

	return nil
}
