# snipe-po

A CLI purchase order system for Snipe-IT that creates POs from assets with PDF generation and email workflow.

## Installation

```bash
# Clone the repository
git clone git@github.com:wzantopulos/snipe-po.git
cd snipe-po

# Install dependencies
go mod tidy

# Build
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o snipe-po .
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
