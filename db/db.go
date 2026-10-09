package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db   *sql.DB
	once sync.Once
)

type PurchaseOrder struct {
	ID              string    `json:"id"`
	PONumber       string    `json:"po_number"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Status          string    `json:"status"`
	Date            string    `json:"date"`
	Supplier        string    `json:"supplier"`
	SupplierContact string    `json:"supplier_contact"`
	Terms           string    `json:"terms"`
	PaymentType     string    `json:"payment_type"`
	Department      string    `json:"department"`
	GLCode          string    `json:"gl_code"`
	Subtotal        float64   `json:"subtotal"`
	ShippingCost    float64   `json:"shipping_cost"`
	TaxCost         float64   `json:"tax_cost"`
	GrandTotal      float64   `json:"grand_total"`
	ApproverEmail   string    `json:"approver_email"`
	APEmail         string    `json:"ap_email"`
	ApprovalNote    string    `json:"approval_note"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	PDFPath         string    `json:"pdf_path"`
	PackingSlipPath string    `json:"packing_slip_path"`
}

type LineItem struct {
	ID          int64   `json:"id"`
	POID        string  `json:"po_id"`
	AssetID     string  `json:"asset_id"`
	Description string  `json:"description"`
	Model       string  `json:"model"`
	MAC         string  `json:"mac"`
	Serial      string  `json:"serial"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
	Department  string  `json:"department"`
	GLCode      string  `json:"gl_code"`
}

func Init() error {
	var initErr error
	once.Do(func() {
		initErr = initDB()
	})
	return initErr
}

func initDB() error {
	dbPath := "/data/po.db"

	sqliteDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	db = sqliteDB

	schema := `
	CREATE TABLE IF NOT EXISTS purchase_orders (
		id TEXT PRIMARY KEY,
		po_number TEXT UNIQUE NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		status TEXT NOT NULL DEFAULT 'draft',
		date TEXT NOT NULL,
		supplier TEXT NOT NULL,
		supplier_contact TEXT DEFAULT '',
		terms TEXT NOT NULL,
		payment_type TEXT NOT NULL,
		department TEXT DEFAULT '',
		gl_code TEXT DEFAULT '',
		subtotal REAL NOT NULL DEFAULT 0,
		shipping_cost REAL NOT NULL DEFAULT 0,
		tax_cost REAL NOT NULL DEFAULT 0,
		grand_total REAL NOT NULL DEFAULT 0,
		approver_email TEXT DEFAULT '',
		ap_email TEXT DEFAULT '',
		approval_note TEXT DEFAULT '',
		approved_at DATETIME,
		pdf_path TEXT DEFAULT '',
		packing_slip_path TEXT DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS line_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		po_id TEXT NOT NULL,
		asset_id TEXT DEFAULT '',
		description TEXT NOT NULL,
		model TEXT DEFAULT '',
		mac TEXT DEFAULT '',
		serial TEXT DEFAULT '',
		quantity INTEGER NOT NULL DEFAULT 1,
		unit_price REAL NOT NULL DEFAULT 0,
		total REAL NOT NULL DEFAULT 0,
		department TEXT DEFAULT '',
		gl_code TEXT DEFAULT '',
		FOREIGN KEY (po_id) REFERENCES purchase_orders(id)
	);

	CREATE TABLE IF NOT EXISTS po_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		po_id TEXT NOT NULL,
		action TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		details TEXT DEFAULT '',
		FOREIGN KEY (po_id) REFERENCES purchase_orders(id)
	);

	CREATE TABLE IF NOT EXISTS po_sequence (
		date_key TEXT PRIMARY KEY,
		last_sequence INTEGER NOT NULL DEFAULT 0
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Migration: add packing_slip_path column if it doesn't exist (ignore errors)
	_, _ = db.Exec(`ALTER TABLE purchase_orders ADD COLUMN packing_slip_path TEXT DEFAULT ''`)

	return nil
}

func GetDB() *sql.DB {
	return db
}

func CreatePO(po *PurchaseOrder) error {
	_, err := db.Exec(`
		INSERT INTO purchase_orders (id, po_number, created_at, updated_at, status, date, supplier, supplier_contact, terms, payment_type, department, gl_code, subtotal, shipping_cost, tax_cost, grand_total, approver_email, ap_email, pdf_path, packing_slip_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		po.ID, po.PONumber, po.CreatedAt, po.UpdatedAt, po.Status, po.Date, po.Supplier, po.SupplierContact, po.Terms, po.PaymentType, po.Department, po.GLCode, po.Subtotal, po.ShippingCost, po.TaxCost, po.GrandTotal, po.ApproverEmail, po.APEmail, po.PDFPath, po.PackingSlipPath)
	return err
}

func GetPO(idOrNumber string) (*PurchaseOrder, error) {
	var po PurchaseOrder
	var approvedAt sql.NullTime

	err := db.QueryRow(`
		SELECT id, po_number, created_at, updated_at, status, date, supplier, supplier_contact, terms, payment_type, department, gl_code, subtotal, shipping_cost, tax_cost, grand_total, approver_email, ap_email, approval_note, approved_at, pdf_path, packing_slip_path
		FROM purchase_orders WHERE id = ? OR po_number = ?`,
		idOrNumber, idOrNumber).Scan(
		&po.ID, &po.PONumber, &po.CreatedAt, &po.UpdatedAt, &po.Status, &po.Date, &po.Supplier, &po.SupplierContact, &po.Terms, &po.PaymentType, &po.Department, &po.GLCode, &po.Subtotal, &po.ShippingCost, &po.TaxCost, &po.GrandTotal, &po.ApproverEmail, &po.APEmail, &po.ApprovalNote, &approvedAt, &po.PDFPath, &po.PackingSlipPath)

	if err != nil {
		return nil, err
	}
	if approvedAt.Valid {
		po.ApprovedAt = &approvedAt.Time
	}
	return &po, nil
}

func GetAllPOs(status string) ([]PurchaseOrder, error) {
	query := `SELECT id, po_number, created_at, updated_at, status, date, supplier, supplier_contact, terms, payment_type, department, gl_code, subtotal, shipping_cost, tax_cost, grand_total, approver_email, ap_email, approval_note, approved_at, pdf_path, packing_slip_path FROM purchase_orders`
	var args []interface{}

	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pos []PurchaseOrder
	for rows.Next() {
		var po PurchaseOrder
		var approvedAt sql.NullTime

		err := rows.Scan(&po.ID, &po.PONumber, &po.CreatedAt, &po.UpdatedAt, &po.Status, &po.Date, &po.Supplier, &po.SupplierContact, &po.Terms, &po.PaymentType, &po.Department, &po.GLCode, &po.Subtotal, &po.ShippingCost, &po.TaxCost, &po.GrandTotal, &po.ApproverEmail, &po.APEmail, &po.ApprovalNote, &approvedAt, &po.PDFPath, &po.PackingSlipPath)
		if err != nil {
			return nil, err
		}
		if approvedAt.Valid {
			po.ApprovedAt = &approvedAt.Time
		}
		pos = append(pos, po)
	}
	return pos, nil
}

func UpdatePO(po *PurchaseOrder) error {
	_, err := db.Exec(`
		UPDATE purchase_orders SET 
			updated_at = ?, status = ?, date = ?, supplier = ?, supplier_contact = ?, terms = ?, payment_type = ?, department = ?, gl_code = ?, subtotal = ?, shipping_cost = ?, tax_cost = ?, grand_total = ?, approver_email = ?, ap_email = ?, approval_note = ?, approved_at = ?, pdf_path = ?, packing_slip_path = ?
		WHERE id = ?`,
		po.UpdatedAt, po.Status, po.Date, po.Supplier, po.SupplierContact, po.Terms, po.PaymentType, po.Department, po.GLCode, po.Subtotal, po.ShippingCost, po.TaxCost, po.GrandTotal, po.ApproverEmail, po.APEmail, po.ApprovalNote, po.ApprovedAt, po.PDFPath, po.PackingSlipPath, po.ID)
	return err
}

func DeletePO(id string) error {
	_, err := db.Exec(`DELETE FROM purchase_orders WHERE id = ?`, id)
	return err
}

func CreateLineItem(item *LineItem) error {
	result, err := db.Exec(`
		INSERT INTO line_items (po_id, asset_id, description, model, mac, serial, quantity, unit_price, total, department, gl_code)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.POID, item.AssetID, item.Description, item.Model, item.MAC, item.Serial, item.Quantity, item.UnitPrice, item.Total, item.Department, item.GLCode)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	item.ID = id
	return nil
}

func GetLineItems(poID string) ([]LineItem, error) {
	rows, err := db.Query(`
		SELECT id, po_id, asset_id, description, model, mac, serial, quantity, unit_price, total, department, gl_code
		FROM line_items WHERE po_id = ?`, poID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []LineItem
	for rows.Next() {
		var item LineItem
		err := rows.Scan(&item.ID, &item.POID, &item.AssetID, &item.Description, &item.Model, &item.MAC, &item.Serial, &item.Quantity, &item.UnitPrice, &item.Total, &item.Department, &item.GLCode)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func AddHistory(poID, action, details string) error {
	_, err := db.Exec(`
		INSERT INTO po_history (po_id, action, timestamp, details) VALUES (?, ?, ?, ?)`,
		poID, action, time.Now(), details)
	return err
}

func GetHistory(poID string) ([]struct {
	Action    string
	Timestamp time.Time
	Details   string
}, error) {
	rows, err := db.Query(`
		SELECT action, timestamp, details FROM po_history WHERE po_id = ? ORDER BY timestamp ASC`, poID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []struct {
		Action    string
		Timestamp time.Time
		Details   string
	}
	for rows.Next() {
		var h struct {
			Action    string
			Timestamp time.Time
			Details   string
		}
		if err := rows.Scan(&h.Action, &h.Timestamp, &h.Details); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}

// GetNextPONumber generates PO number in format IT-YYYYMMDD-### (e.g. IT-20261007-001)
// Sequence is per-day and stored in the po_sequence table.
func GetNextPONumber() (string, error) {
	now := time.Now()
	dateKey := now.Format("20060102")

	// Upsert: insert new date_key with sequence=1, or increment existing
	_, err := db.Exec(`
		INSERT INTO po_sequence (date_key, last_sequence) VALUES (?, 1)
		ON CONFLICT(date_key) DO UPDATE SET last_sequence = last_sequence + 1`,
		dateKey)
	if err != nil {
		return "", fmt.Errorf("failed to increment sequence: %w", err)
	}

	// Read the current sequence value
	var seq int
	err = db.QueryRow(`SELECT last_sequence FROM po_sequence WHERE date_key = ?`, dateKey).Scan(&seq)
	if err != nil {
		return "", fmt.Errorf("failed to read sequence: %w", err)
	}

	return fmt.Sprintf("IT-%s-%03d", dateKey, seq), nil
}
