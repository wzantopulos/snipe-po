package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/wzantopulos/snipe-po/email"
	"github.com/wzantopulos/snipe-po/config"
	"github.com/wzantopulos/snipe-po/db"
	"github.com/wzantopulos/snipe-po/pdf"
)

func jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, message string, status int) {
	jsonResponse(w, map[string]string{"error": message}, status)
}

// GET /api/pos - List all POs
func listPOs(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	pos, err := db.GetAllPOs(status)
	if err != nil {
		errorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, pos, http.StatusOK)
}

// GET /api/pos/:id - Get single PO
func getPO(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	po, err := db.GetPO(vars["id"])
	if err != nil {
		errorResponse(w, "PO not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, po, http.StatusOK)
}

type CreatePORequest struct {
	Supplier      string `json:"supplier"`
	Date         string `json:"date"`
	Department   string `json:"department"`
	GLCode       string `json:"gl_code"`
	Terms        string `json:"terms"`
	PaymentType  string `json:"payment_type"`
	ShippingCost float64 `json:"shipping_cost"`
	TaxCost      float64 `json:"tax_cost"`
	ApproverEmail string `json:"approver_email"`
	APEmail      string `json:"ap_email"`
	LineItems    []struct {
		Description string  `json:"description"`
		Model       string  `json:"model"`
		Serial      string  `json:"serial"`
		Quantity    int     `json:"quantity"`
		UnitPrice   float64 `json:"unit_price"`
	} `json:"line_items"`
}

// POST /api/pos - Create new PO
func createPO(w http.ResponseWriter, r *http.Request) {
	var req CreatePORequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Supplier == "" {
		errorResponse(w, "Supplier is required", http.StatusBadRequest)
		return
	}

	poNumber, err := db.GetNextPONumber()
	if err != nil {
		errorResponse(w, "Failed to generate PO number", http.StatusInternalServerError)
		return
	}

	var subtotal float64
	for _, item := range req.LineItems {
		subtotal += float64(item.Quantity) * item.UnitPrice
	}
	grandTotal := subtotal + req.ShippingCost + req.TaxCost

	date := req.Date
	if date == "" {
		date = time.Now().Format("01/02/2006")
	}

	cfg := config.Get()
	terms := req.Terms
	if terms == "" {
		terms = cfg.PO.DefaultTerms
	}

	paymentType := req.PaymentType
	if paymentType == "" {
		paymentType = cfg.PO.DefaultPaymentType
	}

	now := time.Now()
	po := &db.PurchaseOrder{
		ID:              uuid.New().String(),
		PONumber:        poNumber,
		CreatedAt:       now,
		UpdatedAt:       now,
		Status:          "draft",
		Date:            date,
		Supplier:        req.Supplier,
		Terms:           terms,
		PaymentType:     paymentType,
		Department:      req.Department,
		GLCode:          req.GLCode,
		Subtotal:        subtotal,
		ShippingCost:    req.ShippingCost,
		TaxCost:         req.TaxCost,
		GrandTotal:      grandTotal,
		ApproverEmail:   req.ApproverEmail,
		APEmail:         req.APEmail,
	}

	if err := db.CreatePO(po); err != nil {
		errorResponse(w, "Failed to create PO: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, item := range req.LineItems {
		qty := item.Quantity
		if qty <= 0 {
			qty = 1
		}
		unitPrice := item.UnitPrice
		if unitPrice < 0 {
			unitPrice = 0
		}
		dbItem := &db.LineItem{
			POID:        po.ID,
			Description: item.Description,
			Model:       item.Model,
			Serial:      item.Serial,
			Quantity:    qty,
			UnitPrice:   unitPrice,
			Total:       float64(qty) * unitPrice,
			Department:  req.Department,
			GLCode:      req.GLCode,
		}
		if err := db.CreateLineItem(dbItem); err != nil {
			errorResponse(w, "Failed to create line item: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	pdfPath, err := pdf.GeneratePDF(po, nil)
	if err == nil && pdfPath != "" {
		po.PDFPath = pdfPath
		db.UpdatePO(po)
	}

	db.AddHistory(po.ID, "created", fmt.Sprintf("PO created with %d line items", len(req.LineItems)))

	jsonResponse(w, po, http.StatusCreated)
}

// POST /api/pos/:id/send - Send PO for approval
func sendPO(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	po, err := db.GetPO(vars["id"])
	if err != nil {
		errorResponse(w, "PO not found", http.StatusNotFound)
		return
	}

	if po.Status != "draft" && po.Status != "pending_approval" {
		errorResponse(w, "Only draft or pending POs can be sent for approval", http.StatusBadRequest)
		return
	}

	po.Status = "pending_approval"
	po.UpdatedAt = time.Now()

	if err := db.UpdatePO(po); err != nil {
		errorResponse(w, "Failed to update PO", http.StatusInternalServerError)
		return
	}

	db.AddHistory(po.ID, "sent", "PO sent for approval")

	// Send approval email with HTML and Approve/Reject buttons
	go func() {
		cfg := config.Get()
		appURL := cfg.AppURL
		if appURL == "" {
			appURL = "http://localhost:8080"
		}
		subject := fmt.Sprintf("Purchase Order %s Requires Approval", po.PONumber)
		textBody := fmt.Sprintf("A new purchase order requires your approval.\n\nPO Number: %s\nSupplier: %s\nTotal: $%.2f\n\nPlease review and approve or reject at %s/view?id=%s",
			po.PONumber, po.Supplier, po.GrandTotal, appURL, po.ID)
		htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><style>
body{font-family:Arial,sans-serif;margin:20px;color:#333}
.header{background:#2563eb;color:white;padding:20px;border-radius:8px 8px 0 0}
.content{background:#f8fafc;padding:20px;border:1px solid #e2e8f0;border-top:none}
.info{margin:15px 0}
.label{font-weight:bold;color:#64748b}
.total{font-size:1.5em;color:#2563eb;font-weight:bold;margin:15px 0}
.buttons{margin:25px 0;text-align:center}
.btn{display:inline-block;padding:12px 30px;margin:0 10px;border-radius:6px;text-decoration:none;font-weight:bold}
.btn-approve{background:#16a34a;color:white}
.btn-reject{background:#dc2626;color:white}
.footer{font-size:12px;color:#64748b;margin-top:20px}
</style></head>
<body>
<div class="header">
<h2>Purchase Order Requires Approval</h2>
</div>
<div class="content">
<div class="info"><span class="label">PO Number:</span> %s</div>
<div class="info"><span class="label">Supplier:</span> %s</div>
<div class="info"><span class="label">Date:</span> %s</div>
<div class="info"><span class="label">Department:</span> %s</div>
<div class="info"><span class="label">Terms:</span> %s</div>
<div class="total">Total: $%.2f</div>
<div class="buttons">
<a href="%s/api/pos/%s/approve" class="btn btn-approve">✓ Approve</a>
<a href="%s/view?id=%s" class="btn btn-reject">✗ Reject</a>
</div>
<div class="footer">
You can also view and manage this PO in the web interface.
</div>
</div>
</body>
</html>`,
				po.PONumber, po.Supplier, po.Date, po.Department, po.Terms, po.GrandTotal,
				appURL, po.ID, appURL, po.ID)
		if err := email.SendEmailHTML([]string{po.ApproverEmail}, subject, htmlBody, textBody, po.PDFPath); err != nil {
			fmt.Printf("Failed to send approval email: %v\n", err)
		} else {
			fmt.Printf("Approval email sent to %s\n", po.ApproverEmail)
		}
	}()

	http.Redirect(w, r, "/view?id="+po.ID, http.StatusSeeOther)
}

// POST /api/pos/:id/approve - Approve PO
func approvePO(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	po, err := db.GetPO(vars["id"])
	if err != nil {
		errorResponse(w, "PO not found", http.StatusNotFound)
		return
	}

	if po.Status != "pending_approval" {
		errorResponse(w, "Only POs pending approval can be approved", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	note := r.Form.Get("note")

	po.Status = "approved"
	po.UpdatedAt = time.Now()
	po.ApprovalNote = note
	now := time.Now()
	po.ApprovedAt = &now

	if err := db.UpdatePO(po); err != nil {
		errorResponse(w, "Failed to update PO", http.StatusInternalServerError)
		return
	}

	db.AddHistory(po.ID, "approved", "PO approved")
	if note != "" {
		db.AddHistory(po.ID, "note", note)
	}

	http.Redirect(w, r, "/view?id="+po.ID, http.StatusSeeOther)
}

// POST /api/pos/:id/reject - Reject PO
func deletePO(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	po, err := db.GetPO(vars["id"])
	if err != nil {
		http.Error(w, "PO not found", http.StatusNotFound)
		return
	}

	if err := db.DeletePO(po.ID); err != nil {
		http.Error(w, "Failed to delete PO", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func rejectPO(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	po, err := db.GetPO(vars["id"])
	if err != nil {
		errorResponse(w, "PO not found", http.StatusNotFound)
		return
	}

	if po.Status != "pending_approval" {
		errorResponse(w, "Only POs pending approval can be rejected", http.StatusBadRequest)
		return
	}

	po.Status = "rejected"
	po.UpdatedAt = time.Now()

	if err := db.UpdatePO(po); err != nil {
		errorResponse(w, "Failed to update PO", http.StatusInternalServerError)
		return
	}

	db.AddHistory(po.ID, "rejected", "PO rejected")
	http.Redirect(w, r, "/view?id="+po.ID, http.StatusSeeOther)
}

// POST /api/pos/:id/send-to-ap - Send to AP
func sendToAP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	po, err := db.GetPO(vars["id"])
	if err != nil {
		errorResponse(w, "PO not found", http.StatusNotFound)
		return
	}

	if po.Status != "approved" {
		errorResponse(w, "Only approved POs can be sent to AP", http.StatusBadRequest)
		return
	}

	po.Status = "sent_to_ap"
	po.UpdatedAt = time.Now()

	if err := db.UpdatePO(po); err != nil {
		errorResponse(w, "Failed to update PO", http.StatusInternalServerError)
		return
	}

	db.AddHistory(po.ID, "sent_to_ap", "PO sent to Accounts Payable")
	http.Redirect(w, r, "/view?id="+po.ID, http.StatusSeeOther)
}

// POST /api/pos/:id/mark-paid - Upload packing slip
func uploadPackingSlip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	po, err := db.GetPO(vars["id"])
	if err != nil {
		errorResponse(w, "PO not found", http.StatusNotFound)
		return
	}

	if po.Status != "sent_to_ap" && po.Status != "approved" {
		errorResponse(w, "Only approved POs can have a packing slip uploaded", http.StatusBadRequest)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		errorResponse(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("packing_slip")
	if err != nil {
		errorResponse(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(handler.Filename), ".pdf") {
		errorResponse(w, "Only PDF files are allowed", http.StatusBadRequest)
		return
	}

	// Create packing slips directory
	packingSlipDir := "packing_slips"
	if exePath, err := os.Executable(); err == nil {
		packingSlipDir = filepath.Join(filepath.Dir(exePath), "packing_slips")
	}
	os.MkdirAll(packingSlipDir, 0755)

	// Save file
	safePONumber := strings.ReplaceAll(po.PONumber, "/", "-")
	packingSlipPath := filepath.Join(packingSlipDir, fmt.Sprintf("%s_packing_slip.pdf", safePONumber))

	out, err := os.Create(packingSlipPath)
	if err != nil {
		errorResponse(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		errorResponse(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Update PO
	po.PackingSlipPath = packingSlipPath
	po.UpdatedAt = time.Now()
	if err := db.UpdatePO(po); err != nil {
		errorResponse(w, "Failed to update PO", http.StatusInternalServerError)
		return
	}

	db.AddHistory(po.ID, "packing_slip", "Packing slip uploaded")

	jsonResponse(w, map[string]string{"packing_slip_path": packingSlipPath}, http.StatusOK)
}

// GET /api/pos/:id/pdf - Download PDF
func getPDF(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	po, err := db.GetPO(vars["id"])
	if err != nil {
		errorResponse(w, "PO not found", http.StatusNotFound)
		return
	}

	if po.PDFPath == "" {
		errorResponse(w, "No PDF generated for this PO", http.StatusNotFound)
		return
	}

	// Open the file
	f, err := os.Open(po.PDFPath)
	if err != nil {
		errorResponse(w, "PDF file not found on disk", http.StatusNotFound)
		return
	}
	defer f.Close()

	// Get file info
	fi, err := f.Stat()
	if err != nil {
		errorResponse(w, "Cannot read PDF file", http.StatusNotFound)
		return
	}

	// Set headers
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(po.PDFPath)))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fi.Size()))

	// Copy file to response
	io.Copy(w, f)
}

// GET /api/pos/:id/packing-slip - Download Packing Slip
func getPackingSlip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	po, err := db.GetPO(vars["id"])
	if err != nil {
		errorResponse(w, "PO not found", http.StatusNotFound)
		return
	}

	if po.PackingSlipPath == "" {
		errorResponse(w, "No packing slip uploaded for this PO", http.StatusNotFound)
		return
	}

	// Open the file
	f, err := os.Open(po.PackingSlipPath)
	if err != nil {
		errorResponse(w, "Packing slip file not found on disk", http.StatusNotFound)
		return
	}
	defer f.Close()

	// Get file info
	fi, err := f.Stat()
	if err != nil {
		errorResponse(w, "Cannot read packing slip file", http.StatusNotFound)
		return
	}

	// Set headers
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(po.PackingSlipPath)))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fi.Size()))

	// Copy file to response
	io.Copy(w, f)
}

// GET /api/suppliers - Fetch suppliers from Snipe-IT
func listSuppliers(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	if cfg.SnipeIT.APIKey == "" {
		jsonResponse(w, []map[string]string{}, http.StatusOK)
		return
	}

	// Use internal docker network URL for Snipe-IT API calls
	snipeURL := "http://snipeit:80"
	url := snipeURL + "/api/v1/suppliers"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("suppliers: failed to create request: %v\n", err)
		errorResponse(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+cfg.SnipeIT.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("suppliers: failed to fetch from Snipe-IT: %v\n", err)
		jsonResponse(w, []map[string]string{}, http.StatusOK)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("suppliers: Snipe-IT returned status %d\n", resp.StatusCode)
		jsonResponse(w, []map[string]string{}, http.StatusOK)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("suppliers: failed to read response: %v\n", err)
		jsonResponse(w, []map[string]string{}, http.StatusOK)
		return
	}

	var result struct {
		Rows []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("suppliers: failed to parse response: %v\n", err)
		jsonResponse(w, []map[string]string{}, http.StatusOK)
		return
	}

	suppliers := make([]map[string]string, 0, len(result.Rows))
	for _, m := range result.Rows {
		suppliers = append(suppliers, map[string]string{"id": fmt.Sprintf("%d", m.ID), "name": m.Name})
	}
	fmt.Printf("suppliers: returning %d manufacturers\n", len(suppliers))
	jsonResponse(w, suppliers, http.StatusOK)
}

// Web handlers

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	pos, err := db.GetAllPOs(status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"POs":    pos,
		"Filter": status,
	}

	tmpl := template.Must(template.New("dashboard").Parse(dashboardTemplate))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	// Use the configured Snipe-IT URL as the API URL
	snipeURL := ""
	if cfg != nil {
		snipeURL = strings.TrimRight(cfg.SnipeIT.URL, "/")
	}
	// Derive web UI URL by stripping /api/v1 suffix if present
	webURL := snipeURL
	if strings.HasSuffix(webURL, "/api/v1") {
		webURL = strings.TrimSuffix(webURL, "/api/v1")
	}
	data := map[string]interface{}{
		"Today":         time.Now().Format("2006-01-02"),
		"SnipeITURL":    snipeURL,
		"SnipeITWebURL": webURL,
	}
	tmpl := template.Must(template.New("create").Parse(createTemplate))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	po, err := db.GetPO(id)
	if err != nil {
		http.Error(w, "PO not found", http.StatusNotFound)
		return
	}

	items, _ := db.GetLineItems(id)
	history, _ := db.GetHistory(id)

	data := map[string]interface{}{
		"PO":        po,
		"LineItems": items,
		"History":   history,
	}

	tmpl := template.Must(template.New("view").Parse(viewTemplate))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// API route helpers

func redirectToView(w http.ResponseWriter, r *http.Request, id string) {
	http.Redirect(w, r, "/view?id="+id, http.StatusSeeOther)
}

func safeStr(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func parseTimeStr(s string) time.Time {
	t, _ := time.Parse("01/02/2006", s)
	return t
}

func formatMoney(f float64) string {
	return fmt.Sprintf("$%.2f", f)
}

func statusClass(status string) string {
	return strings.ReplaceAll(status, "_", "-")
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int64) *int {
	v := int(i)
	return &v
}

func floatPtr(f float64) *float64 {
	return &f
}
