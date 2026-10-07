# snipe-po

A CLI purchase order system for Snipe-IT that creates POs from assets with PDF generation and email workflow.

## Installation

### Binary Release

Download the latest release for your platform from [GitHub Releases](https://github.com/wzantopulos/snipe-po/releases).

### Build from Source

```bash
# Clone the repository
git clone git@github.com:wzantopulos/snipe-po.git
cd snipe-po

# Install dependencies
go mod tidy

# Build
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o snipe-po .
```

### Docker

```bash
# Pull the latest image
docker pull ghcr.io/wzantopulos/snipe-po:latest

# Run with persistent data
docker run -d \
  --name snipe-po \
  -p 8080:8080 \
  -v snipe-po-data:/app/data \
  ghcr.io/wzantopulos/snipe-po:latest
```

## Configuration

Copy `settings.example.yaml` to `settings.yaml` in the same directory as the binary, or one of these locations:
- `./settings.yaml`
- `/etc/snipe-po/settings.yaml`

```yaml
company:
  name: "Your Company Name"
  address: "123 Main St"
  city_state_zip: "City, ST 12345"
  phone: "555-555-5555"
  email: "po@yourcompany.com"

snipe_it:
  url: "https://snipe.yourcompany.com"
  api_key: ""

smtp:
  host: "smtp.yourcompany.com"
  port: 587
  username: ""
  password: ""
  from: "po@yourcompany.com"

po:
  next_number: 1
  number_prefix: "PO"
  default_terms: "Net 30"
  default_payment_type: "ACH"
```

## Commands

### snipe-po serve

Start the web UI server for managing purchase orders.

```bash
# Start server on default port 8080
snipe-po serve

# Start server on custom port
snipe-po serve --port 9000
```

The web UI provides:
- **Dashboard**: View and filter all purchase orders
- **Create PO**: Form to create new purchase orders with line items
- **View PO**: View PO details, approve/reject, download PDF

**REST API endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/pos` | List all POs (optional `?status=` filter) |
| GET | `/api/pos/:id` | Get single PO |
| POST | `/api/pos` | Create new PO |
| POST | `/api/pos/:id/send` | Send PO for approval |
| POST | `/api/pos/:id/approve` | Approve PO |
| POST | `/api/pos/:id/reject` | Reject PO |
| POST | `/api/pos/:id/send-to-ap` | Send approved PO to AP |
| POST | `/api/pos/:id/mark-paid` | Mark PO as paid |
| GET | `/api/pos/:id/pdf` | Download PO PDF |

### snipe-po create

Create a new purchase order.

**From Snipe-IT assets:**
```bash
snipe-po create \
  --snipeit-url "https://snipe.yourcompany.com" \
  --snipeit-token "your-api-token" \
  --asset-ids "123,456,789" \
  --supplier "Acme Corp" \
  --department "IT" \
  --gl-code "5200" \
  --approver "manager@yourcompany.com" \
  --ap-email "ap@yourcompany.com"
```

**Interactive mode:**
```bash
snipe-po create --interactive
```
You'll be prompted for each line item details.

**Manual mode:**
```bash
snipe-po create \
  --supplier "Acme Corp" \
  --department "IT" \
  --gl-code "5200" \
  --approver "manager@yourcompany.com"
```

**Flags:**
- `--snipeit-url` - Snipe-IT URL
- `--snipeit-token` - Snipe-IT API token
- `--asset-ids` - Comma-separated Snipe-IT asset IDs
- `--supplier` - Supplier name
- `--terms` - Payment terms (e.g., "Net 30")
- `--payment-type` - Payment type (Credit Card, ACH, Check)
- `--department` - Department name
- `--gl-code` - GL code
- `--shipping` - Shipping cost (default: 0)
- `--tax` - Tax cost (default: 0)
- `--approver` - Approver email
- `--ap-email` - AP email (after approval)
- `--interactive` / `-i` - Interactive line item entry

### snipe-po send

Send a purchase order PDF to approvers via email.

```bash
snipe-po send --id PO-1 --to manager@yourcompany.com
```

If `--to` is not specified, uses the approver email from when the PO was created.

**Flags:**
- `--id` - PO ID or number (required)
- `--to` - Email recipient(s), comma-separated

### snipe-po approve

Approve a purchase order and optionally forward to Accounts Payable.

```bash
snipe-po approve --id PO-1 --note "Approved for Q4 budget"
```

If an AP email was set when creating the PO, it will automatically be forwarded after approval.

**Flags:**
- `--id` - PO ID or number (required)
- `--note` - Optional approval note

### snipe-po list

List all purchase orders.

```bash
# List all POs
snipe-po list

# Filter by status
snipe-po list --status pending_approval

# JSON output
snipe-po list --format json
```

**Flags:**
- `--status` - Filter by status (draft, pending_approval, approved, sent_to_ap, paid, rejected)
- `--format` - Output format: `table` (default) or `json`

### snipe-po version

Print the version number.

```bash
snipe-po version
```

## PO Statuses

| Status | Description |
|--------|-------------|
| `draft` | Created but not sent |
| `pending_approval` | Sent to approver, awaiting response |
| `approved` | Approved by approver |
| `sent_to_ap` | Forwarded to Accounts Payable |
| `paid` | Payment completed |
| `rejected` | Rejected by approver |

## Web UI Screenshots

### Dashboard
```
┌─────────────────────────────────────────────────────────────────┐
│ 📋 snipe-po                          Dashboard    + New PO       │
├─────────────────────────────────────────────────────────────────┤
│ Purchase Orders                          [Filter by Status ▼]    │
├─────────────────────────────────────────────────────────────────┤
│ PO #     │ Date    │ Supplier   │ Dept │ Status           │ $ │
├─────────────────────────────────────────────────────────────────┤
│ PO-1     │ 10/01   │ Acme Corp  │ IT   │ [pending_approval] │2500.00│
│ PO-2     │ 09/28   │ TechSupply │ Ops  │ [approved]        │1200.00│
│ PO-3     │ 09/25   │ Office Co  │ Admin│ [draft]           │ 450.00│
└─────────────────────────────────────────────────────────────────┘
```

### Create PO
```
┌─────────────────────────────────────────────────────────────────┐
│ 📋 snipe-po                          Dashboard    + New PO       │
├─────────────────────────────────────────────────────────────────┤
│ ┌─── Create Purchase Order ────────────────────────────────┐    │
│ │ Supplier *    [Acme Corp        ]  Date *   [2026-10-07]│    │
│ │ Department     [IT               ]  GL Code  [5200      ]│    │
│ │                                                                 │
│ │ Payment Terms [Net 30 ▼]      Payment Type [ACH ▼]          │    │
│ │                                                                 │
│ │ ── Line Items ────────────────────────────────────────────  │    │
│ │ Description      Model      Serial    Qty  Unit Price  [X] │    │
│ │ MacBook Pro      14" M3     ABC123      1    2500.00    [X] │    │
│ │ [+ Add Item]                                              │    │
│ │                                                                 │
│ │ Subtotal     Shipping    Tax                                │    │
│ │ $2500.00     [0.00   ]   [0.00   ]                          │    │
│ │                          Grand Total: $2500.00               │    │
│ │                                                                 │
│ │ Approver Email       AP Email (optional)                      │    │
│ │ [manager@...]       [ap@company.com]                        │    │
│ │                                                                 │
│ │ [Create PO]  [Cancel]                                        │    │
│ └──────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### View PO
```
┌─────────────────────────────────────────────────────────────────┐
│ 📋 snipe-po                          Dashboard    + New PO       │
├─────────────────────────────────────────────────────────────────┤
│ PO #PO-1                                    [approved] 📄 PDF  │
│                                                                 │
│ ┌─── PO Details ──────────────────┐  ┌─── Actions ──────────┐ │
│ │ Supplier: Acme Corp              │  │                        │ │
│ │ Date: 10/01/2026                  │  │ [Send to AP]           │ │
│ │ Department: IT    GL Code: 5200   │  │                        │ │
│ │ Terms: Net 30  Type: ACH         │  │ ────────────────────── │ │
│ │ Approver: manager@company.com    │  │ Info                   │ │
│ │ AP Email: ap@company.com          │  │ Created: 10/01 14:30   │ │
│ └──────────────────────────────────┘  │ Updated: 10/02 09:15   │ │
│                                       │ Approved: 10/02 09:15  │ │
│ ┌─── Line Items ────────────────────────────────────────┐     │ │
│ │ Desc        Model      Serial    Qty  Price   Total   │     │ │
│ │ MacBook Pro 14" M3     ABC123      1  $2500  $2500    │     │ │
│ ├───────────────────────────────────────────────────────┤     │ │
│ │                      Subtotal:           $2500.00      │     │ │
│ │                        Shipping:             $0.00      │     │ │
│ │                             Tax:             $0.00      │     │ │
│ │                      Grand Total:         $2500.00      │     │ │
│ └───────────────────────────────────────────────────────┘     │ │
│                                                                 │
│ ┌─── History ────────────────────────────────────────────┐     │ │
│ │ 10/01 14:30 created PO created with 1 line items       │     │ │
│ │ 10/01 14:35 sent PO sent for approval                   │     │ │
│ │ 10/02 09:15 approved PO approved                        │     │ │
│ └───────────────────────────────────────────────────────┘     │ │
└─────────────────────────────────────────────────────────────────┘
```

## Docker Compose

See `docker-compose.example.yml` for a full setup with snipe-po, Snipe-IT, MariaDB, and Mailhog.

```bash
# Copy and customize
cp docker-compose.example.yml docker-compose.yml
# Edit docker-compose.yml and set your APP_KEY for Snipe-IT

# Start all services
docker-compose up -d

# Access services:
# - snipe-po Web UI: http://localhost:8080
# - Snipe-IT:        http://localhost:8081
# - Mailhog (SMTP):  http://localhost:8025
```

## Database

PO data is stored in `po.db` in the same directory as the binary. The database contains:
- Purchase orders with all header information
- Line items for each PO
- History log of all actions taken on each PO

## PDF Output

PDFs are generated in the `pdfs/` directory next to the binary and include:
- Company header with contact info
- Supplier information
- Line item table (Qty, Description, Model, MAC/Serial, Unit Price, Total)
- Totals section (Subtotal, Shipping, Tax, Grand Total)
- Payment information (Terms, Payment Type)
- Signature line for approval
