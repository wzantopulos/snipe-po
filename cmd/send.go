package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wzantopulos/snipe-po/db"
	"github.com/wzantopulos/snipe-po/email"
)

var (
	sendFlags struct {
		ID string
		To string
	}
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a purchase order for approval",
	Long:  `Email a purchase order PDF as an attachment to the approver.`,
	RunE:  runSend,
}

func init() {
	rootCmd.AddCommand(sendCmd)

	sendCmd.Flags().StringVarP(&sendFlags.ID, "id", "", "", "PO ID or number")
	sendCmd.Flags().StringVarP(&sendFlags.To, "to", "", "", "Email recipient(s), comma-separated")
}

func runSend(cmd *cobra.Command, args []string) error {
	if sendFlags.ID == "" {
		return fmt.Errorf("PO ID or number is required (--id)")
	}

	po, err := db.GetPO(sendFlags.ID)
	if err != nil {
		return fmt.Errorf("PO not found: %w", err)
	}

	if po.PDFPath == "" {
		return fmt.Errorf("PO has no PDF generated")
	}

	if _, err := os.Stat(po.PDFPath); os.IsNotExist(err) {
		return fmt.Errorf("PDF file not found: %s", po.PDFPath)
	}

	to := sendFlags.To
	if to == "" {
		to = po.ApproverEmail
	}

	if to == "" {
		return fmt.Errorf("no recipient specified (--to or --approver when creating PO)")
	}

	recipients := strings.Split(to, ",")
	for i := range recipients {
		recipients[i] = strings.TrimSpace(recipients[i])
	}

	subject := fmt.Sprintf("Purchase Order %s - Approval Required", po.PONumber)
	body := fmt.Sprintf(`Please review and approve Purchase Order %s.

Supplier: %s
Date: %s
Grand Total: $%.2f

Please find the attached PDF for details.

Thank you,
Purchase Order System`, po.PONumber, po.Supplier, po.Date, po.GrandTotal)

	if err := email.SendEmail(recipients, subject, body, po.PDFPath); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	po.Status = "pending_approval"
	db.UpdatePO(po)
	db.AddHistory(po.ID, "sent", fmt.Sprintf("Sent to %s", strings.Join(recipients, ", ")))

	fmt.Printf("Purchase Order %s sent to %s\n", po.PONumber, to)

	return nil
}
