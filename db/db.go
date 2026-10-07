package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mattn/go-sqlite3"
)

var (
	db   *sql.DB
	once sync.Once
)

type PurchaseOrder struct {
	ID              string
	PONumber       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Status          string
	Date            string
	Supplier        string
	SupplierContact string
	Terms           string
	PaymentType     string
	Department      string
	GLCode          string
	Subtotal        float64
	ShippingCost    float64
	TaxCost         float64
	GrandTotal      float64
	ApproverEmail   string
	APEmail         string
	ApprovalNote    string
	ApprovedAt      *time.Time
	PDFPath         string
}

type LineItem struct {
	ID          int64
	POID        string
	AssetID     string
	Description string
	Model       string
	MAC         string
	Serial      string
	Quantity    int
	UnitPrice   float64
	Total       float64
	Department  string
	GLCode      string
}

func Init() error {
	var initErr error
	once.Do(func() {
		initErr = initDB()
	})
	return initErr
}

func initDB() error {
	exePath, err := os.Executable()
	dbPath := "po.db"
	if err == nil {
		dbPath = filepath.Join(filepath.Dir(exePath), "po.db")
	}

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
		pdf_path TEXT DEFAULT ''
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
	`

	_, err = db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

func GetDB() *sql.DB {
	return db
}

func CreatePO(po *PurchaseOrder) error {
	_, err := db.Exec(`
		INSERT INTO purchase_orders (id, po_number, created_at, updated_at, status, date, supplier, supplier_contact, terms, payment_type, department, gl_code, subtotal, shipping_cost, tax_cost, grand_total, approver_email, ap_email, pdf_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		po.ID, po.PONumber, po.CreatedAt, po.UpdatedAt, po.Status, po.Date, po.Supplier, po.SupplierContact, po.Terms, po.PaymentType, po.Department, po.GLCode, po.Subtotal, po.ShippingCost, po.TaxCost, po.GrandTotal, po.ApproverEmail, po.APEmail, po.PDFPath)
	return err
}

func GetPO(idOrNumber string) (*PurchaseOrder, error) {
	var po PurchaseOrder
	var approvedAt sql.NullTime

	err := db.QueryRow(`
		SELECT id, po_number, created_at, updated_at, status, date, supplier, supplier_contact, terms, payment_type, department, gl_code, subtotal, shipping_cost, tax_cost, grand_total, approver_email, ap_email, approval_note, approved_at, pdf_path
		FROM purchase_orders WHERE id = ? OR po_number = ?`,
		idOrNumber, idOrNumber).Scan(
		&po.ID, &po.PONumber, &po.CreatedAt, &po.UpdatedAt, &po.Status, &po.Date, &po.Supplier, &po.SupplierContact, &po.Terms, &po.PaymentType, &po.Department, &po.GLCode, &po.Subtotal, &po.ShippingCost, &po.TaxCost, &po.GrandTotal, &po.ApproverEmail, &po.APEmail, &po.ApprovalNote, &approvedAt, &po.PDFPath)

	if err != nil {
		return nil, err
	}
	if approvedAt.Valid {
		po.ApprovedAt = &approvedAt.Time
	}
	return &po, nil
}

func GetAllPOs(status string) ([]PurchaseOrder, error) {
	query := `SELECT id, po_number, created_at, updated_at, status, date, supplier, supplier_contact, terms, payment_type, department, gl_code, subtotal, shipping_cost, tax_cost, grand_total, approver_email, ap_email, approval_note, approved_at, pdf_path FROM purchase_orders`
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

		err := rows.Scan(&po.ID, &po.PONumber, &po.CreatedAt, &po.UpdatedAt, &po.Status, &po.Date, &po.Supplier, &po.SupplierContact, &po.Terms, &po.PaymentType, &po.Department, &po.GLCode, &po.Subtotal, &po.ShippingCost, &po.TaxCost, &po.GrandTotal, &po.ApproverEmail, &po.APEmail, &po.ApprovalNote, &approvedAt, &po.PDFPath)
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
			updated_at = ?, status = ?, date = ?, supplier = ?, supplier_contact = ?, terms = ?, payment_type = ?, department = ?, gl_code = ?, subtotal = ?, shipping_cost = ?, tax_cost = ?, grand_total = ?, approver_email = ?, ap_email = ?, approval_note = ?, approved_at = ?, pdf_path = ?
		WHERE id = ?`,
		po.UpdatedAt, po.Status, po.Date, po.Supplier, po.SupplierContact, po.Terms, po.PaymentType, po.Department, po.GLCode, po.Subtotal, po.ShippingCost, po.TaxCost, po.GrandTotal, po.ApproverEmail, po.APEmail, po.ApprovalNote, po.ApprovedAt, po.PDFPath, po.ID)
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

func GetNextPONumber(prefix string) (string, error) {
	var nextNum int
	err := db.QueryRow(`SELECT COALESCE(MAX(CAST(SUBSTR(po_number, 5, 4) AS INTEGER)), 0) FROM purchase_orders WHERE po_number LIKE ?`, prefix+"-%").Scan(&nextNum)
	if err != nil {
		nextNum = 0
	}
	nextNum++
	return fmt.Sprintf("%s-%d", prefix, nextNum), nil
}
