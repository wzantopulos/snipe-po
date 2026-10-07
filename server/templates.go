package server

var dashboardTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>snipe-po - Dashboard</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css" rel="stylesheet">
    <style>
        body { background: #f8f9fa; }
        .navbar-brand { font-weight: 600; }
        .status-badge { text-transform: capitalize; }
        .status-draft { background: #6c757d; }
        .status-pending_approval { background: #ffc107; color: #000; }
        .status-approved { background: #198754; }
        .status-sent_to_ap { background: #0d6efd; }
        .status-paid { background: #17a2b8; }
        .status-rejected { background: #dc3545; }
        .table-hover tbody tr:hover { cursor: pointer; }
        .amount { text-align: right; font-family: monospace; }
        .nav-link.active { font-weight: 600; }
    </style>
</head>
<body>
    <nav class="navbar navbar-expand-lg navbar-dark bg-primary">
        <div class="container">
            <a class="navbar-brand" href="/">&#128196; snipe-po</a>
            <div class="navbar-nav">
                <a class="nav-link active" href="/">Dashboard</a>
                <a class="nav-link" href="/create">+ New PO</a>
            </div>
        </div>
    </nav>

    <div class="container mt-4">
        <div class="row mb-4">
            <div class="col">
                <h2>Purchase Orders</h2>
            </div>
            <div class="col-auto">
                <form method="get" class="d-flex gap-2">
                    <select name="status" class="form-select" onchange="this.form.submit()">
                        <option value="">All Statuses</option>
                        <option value="draft" {{if eq .Filter "draft"}}selected{{end}}>Draft</option>
                        <option value="pending_approval" {{if eq .Filter "pending_approval"}}selected{{end}}>Pending Approval</option>
                        <option value="approved" {{if eq .Filter "approved"}}selected{{end}}>Approved</option>
                        <option value="sent_to_ap" {{if eq .Filter "sent_to_ap"}}selected{{end}}>Sent to AP</option>
                        <option value="paid" {{if eq .Filter "paid"}}selected{{end}}>Paid</option>
                        <option value="rejected" {{if eq .Filter "rejected"}}selected{{end}}>Rejected</option>
                    </select>
                </form>
            </div>
        </div>

        <div class="card shadow-sm">
            <table class="table table-hover mb-0">
                <thead class="table-light">
                    <tr>
                        <th>PO #</th>
                        <th>Date</th>
                        <th>Supplier</th>
                        <th>Department</th>
                        <th>Status</th>
                        <th class="amount">Total</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .POs}}
                    <tr onclick="window.location='/view?id={{.ID}}'">
                        <td><strong>{{.PONumber}}</strong></td>
                        <td>{{.Date}}</td>
                        <td>{{.Supplier}}</td>
                        <td>{{.Department}}</td>
                        <td><span class="badge status-badge status-{{.Status}}">{{.Status}}</span></td>
                        <td class="amount">${{printf "%.2f" .GrandTotal}}</td>
                    </tr>
                    {{else}}
                    <tr>
                        <td colspan="6" class="text-center text-muted py-4">No purchase orders found</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
    </div>
</body>
</html>`

var createTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>snipe-po - Create PO</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css" rel="stylesheet">
    <style>
        body { background: #f8f9fa; }
        .navbar-brand { font-weight: 600; }
        .nav-link.active { font-weight: 600; }
        .line-item-row { display: flex; gap: 10px; margin-bottom: 10px; align-items: flex-start; }
        .line-item-row input { flex: 1; }
        .line-item-row input[type="number"] { max-width: 100px; flex: 0 0 100px; }
        .remove-item { flex: 0 0 40px; }
        #line-items-container .card { margin-bottom: 10px; }
    </style>
</head>
<body>
    <nav class="navbar navbar-expand-lg navbar-dark bg-primary">
        <div class="container">
            <a class="navbar-brand" href="/">&#128196; snipe-po</a>
            <div class="navbar-nav">
                <a class="nav-link" href="/">Dashboard</a>
                <a class="nav-link active" href="/create">+ New PO</a>
            </div>
        </div>
    </nav>

    <div class="container mt-4">
        <div class="row">
            <div class="col-lg-8">
                <div class="card shadow-sm">
                    <div class="card-header bg-primary text-white">
                        <h5 class="mb-0">Create Purchase Order</h5>
                    </div>
                    <div class="card-body">
                        <form id="po-form" method="post" action="/api/pos">
                            <div class="row mb-3">
                                <div class="col-md-6">
                                    <label class="form-label">Supplier *</label>
                                    <input type="text" name="supplier" class="form-control" required>
                                </div>
                                <div class="col-md-6">
                                    <label class="form-label">Date *</label>
                                    <input type="date" name="date" class="form-control" value="{{.Today}}" required>
                                </div>
                            </div>

                            <div class="row mb-3">
                                <div class="col-md-6">
                                    <label class="form-label">Department</label>
                                    <input type="text" name="department" class="form-control">
                                </div>
                                <div class="col-md-6">
                                    <label class="form-label">GL Code</label>
                                    <input type="text" name="gl_code" class="form-control">
                                </div>
                            </div>

                            <div class="row mb-3">
                                <div class="col-md-6">
                                    <label class="form-label">Payment Terms</label>
                                    <select name="terms" class="form-select">
                                        <option value="Net 30">Net 30</option>
                                        <option value="Net 15">Net 15</option>
                                        <option value="Net 60">Net 60</option>
                                        <option value="Due on Receipt">Due on Receipt</option>
                                    </select>
                                </div>
                                <div class="col-md-6">
                                    <label class="form-label">Payment Type</label>
                                    <select name="payment_type" class="form-select">
                                        <option value="ACH">ACH</option>
                                        <option value="Credit Card">Credit Card</option>
                                        <option value="Check">Check</option>
                                    </select>
                                </div>
                            </div>

                            <hr>
                            <h5>Line Items</h5>
                            <div id="line-items-container">
                                <div class="line-item-row" data-index="0">
                                    <input type="text" name="items[0].description" placeholder="Description" required>
                                    <input type="text" name="items[0].model" placeholder="Model">
                                    <input type="text" name="items[0].serial" placeholder="Serial">
                                    <input type="number" name="items[0].quantity" placeholder="Qty" min="1" value="1">
                                    <input type="number" name="items[0].unit_price" placeholder="Unit Price" step="0.01" min="0">
                                    <button type="button" class="btn btn-outline-danger btn-sm remove-item" onclick="removeLineItem(this)">X</button>
                                </div>
                            </div>
                            <button type="button" class="btn btn-outline-primary btn-sm mt-2" onclick="addLineItem()">+ Add Item</button>

                            <hr>
                            <div class="row mb-3">
                                <div class="col-md-4">
                                    <label class="form-label">Subtotal</label>
                                    <div class="form-control" id="subtotal">$0.00</div>
                                </div>
                                <div class="col-md-4">
                                    <label class="form-label">Shipping</label>
                                    <input type="number" name="shipping_cost" class="form-control" step="0.01" min="0" value="0" onchange="calculateTotals()">
                                </div>
                                <div class="col-md-4">
                                    <label class="form-label">Tax</label>
                                    <input type="number" name="tax_cost" class="form-control" step="0.01" min="0" value="0" onchange="calculateTotals()">
                                </div>
                            </div>
                            <div class="row mb-3">
                                <div class="col-md-4 offset-md-8">
                                    <label class="form-label"><strong>Grand Total</strong></label>
                                    <div class="form-control bg-light" id="grand_total"><strong>$0.00</strong></div>
                                </div>
                            </div>

                            <hr>
                            <h5>Workflow</h5>
                            <div class="row mb-3">
                                <div class="col-md-6">
                                    <label class="form-label">Approver Email</label>
                                    <input type="email" name="approver_email" class="form-control" placeholder="manager@company.com">
                                </div>
                                <div class="col-md-6">
                                    <label class="form-label">AP Email (after approval)</label>
                                    <input type="email" name="ap_email" class="form-control" placeholder="ap@company.com">
                                </div>
                            </div>

                            <div class="d-flex gap-2">
                                <button type="submit" class="btn btn-primary">Create PO</button>
                                <a href="/" class="btn btn-outline-secondary">Cancel</a>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
            <div class="col-lg-4">
                <div class="card shadow-sm">
                    <div class="card-header">
                        <h6 class="mb-0">Quick Tips</h6>
                    </div>
                    <div class="card-body">
                        <ul class="mb-0">
                            <li>Create POs as drafts and send for approval when ready</li>
                            <li>Use the CLI for Snipe-IT asset imports</li>
                            <li>PDFs are generated automatically</li>
                        </ul>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        var itemIndex = 1;

        function addLineItem() {
            var container = document.getElementById("line-items-container");
            var div = document.createElement("div");
            div.className = "line-item-row";
            div.setAttribute("data-index", itemIndex);
            div.innerHTML = "<input type=\"text\" name=\"items[" + itemIndex + "].description\" placeholder=\"Description\" required>" +
                "<input type=\"text\" name=\"items[" + itemIndex + "].model\" placeholder=\"Model\">" +
                "<input type=\"text\" name=\"items[" + itemIndex + "].serial\" placeholder=\"Serial\">" +
                "<input type=\"number\" name=\"items[" + itemIndex + "].quantity\" placeholder=\"Qty\" min=\"1\" value=\"1\">" +
                "<input type=\"number\" name=\"items[" + itemIndex + "].unit_price\" placeholder=\"Unit Price\" step=\"0.01\" min=\"0\">" +
                "<button type=\"button\" class=\"btn btn-outline-danger btn-sm remove-item\" onclick=\"removeLineItem(this)\">X</button>";
            container.appendChild(div);
            itemIndex++;
        }

        function removeLineItem(btn) {
            var rows = document.querySelectorAll(".line-item-row");
            if (rows.length > 1) {
                btn.closest(".line-item-row").remove();
                calculateTotals();
            }
        }

        document.getElementById("line-items-container").addEventListener("input", function(e) {
            if (e.target.name && e.target.name.includes(".unit_price") || e.target.name && e.target.name.includes(".quantity")) {
                calculateTotals();
            }
        });

        function calculateTotals() {
            var subtotal = 0;
            var rows = document.querySelectorAll(".line-item-row");
            rows.forEach(function(row) {
                var qty = parseFloat(row.querySelector('input[name*=".quantity"]').value) || 0;
                var price = parseFloat(row.querySelector('input[name*=".unit_price"]').value) || 0;
                subtotal += qty * price;
            });
            var shipping = parseFloat(document.querySelector('input[name="shipping_cost"]').value) || 0;
            var tax = parseFloat(document.querySelector('input[name="tax_cost"]').value) || 0;
            var grandTotal = subtotal + shipping + tax;

            document.getElementById("subtotal").textContent = "$" + subtotal.toFixed(2);
            document.getElementById("grand_total").innerHTML = "<strong>$" + grandTotal.toFixed(2) + "</strong>";
        }

        document.getElementById("po-form").addEventListener("submit", function(e) {
            e.preventDefault();
            var form = e.target;
            var formData = new FormData(form);

            var data = {
                supplier: formData.get("supplier"),
                date: formData.get("date"),
                department: formData.get("department"),
                gl_code: formData.get("gl_code"),
                terms: formData.get("terms"),
                payment_type: formData.get("payment_type"),
                shipping_cost: parseFloat(formData.get("shipping_cost")) || 0,
                tax_cost: parseFloat(formData.get("tax_cost")) || 0,
                approver_email: formData.get("approver_email"),
                ap_email: formData.get("ap_email"),
                line_items: []
            };

            document.querySelectorAll(".line-item-row").forEach(function(row) {
                var desc = row.querySelector('input[name*=".description"]').value;
                if (desc) {
                    data.line_items.push({
                        description: desc,
                        model: row.querySelector('input[name*=".model"]').value,
                        serial: row.querySelector('input[name*=".serial"]').value,
                        quantity: parseInt(row.querySelector('input[name*=".quantity"]').value) || 1,
                        unit_price: parseFloat(row.querySelector('input[name*=".unit_price"]').value) || 0
                    });
                }
            });

            fetch("/api/pos", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(data)
            }).then(function(response) {
                if (response.ok) {
                    return response.json();
                } else {
                    throw new Error("Error creating PO");
                }
            }).then(function(result) {
                window.location.href = "/view?id=" + result.id;
            }).catch(function(err) {
                alert("Error: " + err.message);
            });
        });
    </script>
</body>
</html>`

var viewTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>snipe-po - View PO #` + "{{.PO.PONumber}}" + `</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css" rel="stylesheet">
    <style>
        body { background: #f8f9fa; }
        .navbar-brand { font-weight: 600; }
        .nav-link.active { font-weight: 600; }
        .status-badge { text-transform: capitalize; }
        .status-draft { background: #6c757d; }
        .status-pending_approval { background: #ffc107; color: #000; }
        .status-approved { background: #198754; }
        .status-sent_to_ap { background: #0d6efd; }
        .status-paid { background: #17a2b8; }
        .status-rejected { background: #dc3545; }
        .amount { font-family: monospace; }
        .history-timeline { font-size: 0.9em; }
    </style>
</head>
<body>
    <nav class="navbar navbar-expand-lg navbar-dark bg-primary">
        <div class="container">
            <a class="navbar-brand" href="/">&#128196; snipe-po</a>
            <div class="navbar-nav">
                <a class="nav-link" href="/">Dashboard</a>
                <a class="nav-link" href="/create">+ New PO</a>
            </div>
        </div>
    </nav>

    <div class="container mt-4">
        <div class="d-flex justify-content-between align-items-center mb-3">
            <h2>PO #` + "{{.PO.PONumber}}" + `</h2>
            <div>
                <span class="badge status-badge status-{{.PO.Status}} fs-6">{{.PO.Status}}</span>
                {{if .PO.PDFPath}}
                <a href="/api/pos/{{.PO.ID}}/pdf" class="btn btn-outline-secondary btn-sm">&#128196; Download PDF</a>
                {{end}}
            </div>
        </div>

        <div class="row">
            <div class="col-lg-8">
                <!-- PO Details -->
                <div class="card shadow-sm mb-4">
                    <div class="card-header">
                        <h5 class="mb-0">PO Details</h5>
                    </div>
                    <div class="card-body">
                        <div class="row mb-3">
                            <div class="col-md-6">
                                <strong>Supplier:</strong> {{.PO.Supplier}}<br>
                                <strong>Date:</strong> {{.PO.Date}}
                            </div>
                            <div class="col-md-6">
                                <strong>Department:</strong> {{.PO.Department}}<br>
                                <strong>GL Code:</strong> {{.PO.GLCode}}
                            </div>
                        </div>
                        <div class="row mb-3">
                            <div class="col-md-6">
                                <strong>Payment Terms:</strong> {{.PO.Terms}}
                            </div>
                            <div class="col-md-6">
                                <strong>Payment Type:</strong> {{.PO.PaymentType}}
                            </div>
                        </div>
                        <div class="row">
                            <div class="col-md-6">
                                <strong>Approver:</strong> {{.PO.ApproverEmail}}
                            </div>
                            <div class="col-md-6">
                                <strong>AP Email:</strong> {{.PO.APEmail}}
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Line Items -->
                <div class="card shadow-sm mb-4">
                    <div class="card-header">
                        <h5 class="mb-0">Line Items</h5>
                    </div>
                    <table class="table table-hover mb-0">
                        <thead>
                            <tr>
                                <th>Description</th>
                                <th>Model</th>
                                <th>Serial</th>
                                <th class="text-center">Qty</th>
                                <th class="text-end">Unit Price</th>
                                <th class="text-end">Total</th>
                            </tr>
                        </thead>
                        <tbody>
                            {{range .LineItems}}
                            <tr>
                                <td>{{.Description}}</td>
                                <td>{{.Model}}</td>
                                <td><small>{{.Serial}}</small></td>
                                <td class="text-center">{{.Quantity}}</td>
                                <td class="text-end amount">${{printf "%.2f" .UnitPrice}}</td>
                                <td class="text-end amount">${{printf "%.2f" .Total}}</td>
                            </tr>
                            {{end}}
                        </tbody>
                        <tfoot class="table-light">
                            <tr>
                                <td colspan="5" class="text-end"><strong>Subtotal:</strong></td>
                                <td class="text-end amount">${{printf "%.2f" .PO.Subtotal}}</td>
                            </tr>
                            <tr>
                                <td colspan="5" class="text-end">Shipping:</td>
                                <td class="text-end amount">${{printf "%.2f" .PO.ShippingCost}}</td>
                            </tr>
                            <tr>
                                <td colspan="5" class="text-end">Tax:</td>
                                <td class="text-end amount">${{printf "%.2f" .PO.TaxCost}}</td>
                            </tr>
                            <tr>
                                <td colspan="5" class="text-end"><strong>Grand Total:</strong></td>
                                <td class="text-end amount"><strong>${{printf "%.2f" .PO.GrandTotal}}</strong></td>
                            </tr>
                        </tfoot>
                    </table>
                </div>

                <!-- Approval Note -->
                {{if .PO.ApprovalNote}}
                <div class="card shadow-sm mb-4">
                    <div class="card-header">
                        <h5 class="mb-0">Approval Note</h5>
                    </div>
                    <div class="card-body">
                        <p class="mb-0">{{.PO.ApprovalNote}}</p>
                    </div>
                </div>
                {{end}}

                <!-- History -->
                {{if .History}}
                <div class="card shadow-sm">
                    <div class="card-header">
                        <h5 class="mb-0">History</h5>
                    </div>
                    <div class="card-body history-timeline">
                        {{range .History}}
                        <div class="d-flex mb-2">
                            <div class="me-3 text-muted">{{.Timestamp.Format "01/02 15:04"}}</div>
                            <div><strong>{{.Action}}</strong> {{.Details}}</div>
                        </div>
                        {{end}}
                    </div>
                </div>
                {{end}}
            </div>

            <div class="col-lg-4">
                <!-- Actions -->
                <div class="card shadow-sm mb-4">
                    <div class="card-header bg-primary text-white">
                        <h5 class="mb-0">Actions</h5>
                    </div>
                    <div class="card-body d-grid gap-2">
                        {{if eq .PO.Status "draft"}}
                        <form method="post" action="/api/pos/{{.PO.ID}}/send">
                            <button type="submit" class="btn btn-primary w-100">&#128228; Send for Approval</button>
                        </form>
                        {{end}}

                        {{if eq .PO.Status "pending_approval"}}
                        <form method="post" action="/api/pos/{{.PO.ID}}/approve">
                            <div class="mb-2">
                                <label class="form-label">Note (optional)</label>
                                <textarea name="note" class="form-control" rows="2"></textarea>
                            </div>
                            <button type="submit" class="btn btn-success w-100">&#10004; Approve</button>
                        </form>
                        <form method="post" action="/api/pos/{{.PO.ID}}/reject">
                            <button type="submit" class="btn btn-danger w-100">&#10008; Reject</button>
                        </form>
                        {{end}}

                        {{if eq .PO.Status "approved"}}
                        <form method="post" action="/api/pos/{{.PO.ID}}/send-to-ap">
                            <button type="submit" class="btn btn-primary w-100">&#128228; Send to AP</button>
                        </form>
                        {{end}}

                        {{if eq .PO.Status "sent_to_ap"}}
                        <form method="post" action="/api/pos/{{.PO.ID}}/mark-paid">
                            <button type="submit" class="btn btn-info w-100">&#128176; Mark as Paid</button>
                        </form>
                        {{end}}

                        <a href="/" class="btn btn-outline-secondary">&larr; Back to Dashboard</a>
                    </div>
                </div>

                <!-- Info -->
                <div class="card shadow-sm">
                    <div class="card-header">
                        <h6 class="mb-0">Info</h6>
                    </div>
                    <div class="card-body">
                        <p class="mb-1"><small class="text-muted">Created:</small><br>{{.PO.CreatedAt.Format "01/02/2006 3:04 PM"}}</p>
                        <p class="mb-1"><small class="text-muted">Updated:</small><br>{{.PO.UpdatedAt.Format "01/02/2006 3:04 PM"}}</p>
                        {{if .PO.ApprovedAt}}
                        <p class="mb-0"><small class="text-muted">Approved:</small><br>{{.PO.ApprovedAt.Format "01/02/2006 3:04 PM"}}</p>
                        {{end}}
                    </div>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`
