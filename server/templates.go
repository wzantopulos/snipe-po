package server

// Snipe-IT brand colors
const (
	snipeDarkBg    = "#2b2b2b"
	snipeDarkBg2   = "#3d3d3d"
	snipeAccent    = "#3c8dbc"
	snipeAccentHov = "#337ab7"
	snipeText      = "#eeeeee"
	snipeTextMuted = "#aaaaaa"
	snipeBorder    = "#444444"
)

var dashboardTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>snipe-po - Dashboard</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css" rel="stylesheet">
    <style id="theme-styles">
        :root {
            --bg-primary: #2b2b2b;
            --bg-secondary: #3d3d3d;
            --accent: #3c8dbc;
            --accent-hover: #337ab7;
            --text-main: #eeeeee;
            --text-muted: #aaaaaa;
            --border-color: #444444;
            --card-bg: #323232;
            --table-bg: #3d3d3d;
            --light-text: #ffffff;
        }
        body { background: var(--bg-primary); color: var(--text-main); }
        .navbar { background: var(--bg-secondary) !important; border-bottom: 1px solid var(--border-color); }
        .navbar-brand { font-weight: 600; color: var(--text-main) !important; display: flex; align-items: center; gap: 8px; }
        .navbar-brand img { height: 28px; }
        .nav-link { color: var(--text-muted) !important; }
        .nav-link:hover, .nav-link.active { color: var(--text-main) !important; }
        .card { background: var(--card-bg); border: 1px solid var(--border-color); }
        .table { color: var(--text-main); }
        .table-light { background: var(--table-bg) !important; color: var(--text-main) !important; }
        .form-control, .form-select { background: var(--bg-secondary); color: var(--text-main); border-color: var(--border-color); }
        .form-control:focus, .form-select:focus { background: var(--bg-secondary); color: var(--text-main); border-color: var(--accent); }
        .form-control::placeholder { color: var(--text-muted); }
        .status-badge { text-transform: capitalize; }
        .status-draft { background: #6c757d; }
        .status-pending_approval { background: #ffc107; color: #000; }
        .status-approved { background: #198754; }
        .status-sent_to_ap { background: #0d6efd; }
        .status-paid { background: #17a2b8; }
        .status-rejected { background: #dc3545; }
        .table-hover tbody tr:hover { cursor: pointer; background: var(--bg-secondary) !important; }
        .amount { text-align: right; font-family: monospace; }
        .text-muted { color: var(--text-muted) !important; }
        h2, h5, h6 { color: var(--text-main); }
        .badge { color: var(--light-text); }
        /* Dark theme toggle button */
        .theme-toggle {
            background: transparent;
            border: 1px solid var(--border-color);
            color: var(--text-muted);
            padding: 4px 10px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.85rem;
        }
        .theme-toggle:hover { border-color: var(--accent); color: var(--accent); }
        /* Light theme overrides */
        body.theme-light { background: #f8f9fa; color: #212529; }
        body.theme-light .navbar { background: var(--accent) !important; border-bottom: none; }
        body.theme-light .navbar-brand { color: #fff !important; }
        body.theme-light .nav-link { color: rgba(255,255,255,0.75) !important; }
        body.theme-light .nav-link:hover, body.theme-light .nav-link.active { color: #fff !important; }
        body.theme-light .card { background: #fff; border: 1px solid #dee2e6; }
        body.theme-light .table { color: #212529; }
        body.theme-light .table-light { background: #f8f9fa !important; color: #212529 !important; }
        body.theme-light .card-body { color: #212529; }
        body.theme-light .card-body strong { color: #212529; }
        body.theme-light thead { background: #f8f9fa !important; color: #212529 !important; }
        body.theme-light thead th { background: #f8f9fa !important; color: #212529 !important; }
        body.theme-light tfoot { background: #f8f9fa !important; color: #212529 !important; }
        body.theme-light tfoot td { background: #f8f9fa !important; color: #212529 !important; }
        body.theme-light .form-control, body.theme-light .form-select { background: #fff; color: #212529; border-color: #ced4da; }
        body.theme-light .form-control:focus { border-color: var(--accent); }
        body.theme-light .form-control::placeholder { color: #6c757d; }
        body.theme-light h2, body.theme-light h5, body.theme-light h6 { color: #212529; }
        body.theme-light .text-muted { color: #6c757d !important; }
        body.theme-light .theme-toggle { border-color: rgba(255,255,255,0.5); color: rgba(255,255,255,0.8); }
        body.theme-light .theme-toggle:hover { border-color: #fff; color: #fff; }
    </style>
</head>
<body>
    <nav class="navbar navbar-expand-lg navbar-dark">
        <div class="container">
            <a class="navbar-brand" href="/">
                <img src="https://snipe-it.io/wp-content/uploads/2022/09/snipe-it-logo.svg" alt="snipe-it" onerror="this.style.display='none'">
                <span>&#128196; snipe-po</span>
            </a>
            <div class="navbar-nav">
                <a class="nav-link active" href="/">Dashboard</a>
                <a class="nav-link" href="/create">+ New PO</a>
            </div>
            <button class="theme-toggle" onclick="toggleTheme()">&#9788; Dark</button>
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

    <script>
        // Theme toggle
        function toggleTheme() {
            var body = document.body;
            var btn = document.querySelector('.theme-toggle');
            if (body.classList.contains('theme-light')) {
                body.classList.remove('theme-light');
                localStorage.setItem('theme', 'dark');
                btn.innerHTML = '&#9788; Dark';
            } else {
                body.classList.add('theme-light');
                localStorage.setItem('theme', 'light');
                btn.innerHTML = '&#9790; Light';
            }
        }
        // Restore theme from localStorage
        (function() {
            var theme = localStorage.getItem('theme');
            if (theme === 'light') {
                document.body.classList.add('theme-light');
                document.querySelector('.theme-toggle').innerHTML = '&#9790; Light';
            }
        })();
    </script>
</body>
</html>`

var createTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>snipe-po - Create PO</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/select2@4.1.0-rc.0/dist/css/select2.min.css" rel="stylesheet">
    <style id="theme-styles">
        :root {
            --bg-primary: #2b2b2b;
            --bg-secondary: #3d3d3d;
            --accent: #3c8dbc;
            --accent-hover: #337ab7;
            --text-main: #eeeeee;
            --text-muted: #aaaaaa;
            --border-color: #444444;
            --card-bg: #323232;
            --table-bg: #3d3d3d;
            --light-text: #ffffff;
        }
        body { background: var(--bg-primary); color: var(--text-main); }
        .navbar { background: var(--bg-secondary) !important; border-bottom: 1px solid var(--border-color); }
        .navbar-brand { font-weight: 600; color: var(--text-main) !important; display: flex; align-items: center; gap: 8px; }
        .navbar-brand img { height: 28px; }
        .nav-link { color: var(--text-muted) !important; }
        .nav-link:hover, .nav-link.active { color: var(--text-main) !important; }
        .card { background: var(--card-bg); border: 1px solid var(--border-color); }
        .card-header { background: var(--bg-secondary) !important; border-bottom: 1px solid var(--border-color); color: var(--text-main); }
        .form-control, .form-select { background: var(--bg-secondary); color: var(--text-main); border-color: var(--border-color); }
        .form-control:focus, .form-select:focus { background: var(--bg-secondary); color: var(--text-main); border-color: var(--accent); }
        .form-control::placeholder { color: var(--text-muted); }
        .form-label { color: var(--text-muted); }
        h5, h6, .mb-0 { color: var(--text-main); }
        hr { border-color: var(--border-color); }
        .text-muted { color: var(--text-muted) !important; }
        .line-item-row { display: flex; gap: 10px; margin-bottom: 10px; align-items: flex-start; flex-wrap: wrap; }
        .line-item-row input { flex: 1; min-width: 120px; }
        .line-item-row input[type="number"] { max-width: 100px; flex: 0 0 100px; }
        .remove-item { flex: 0 0 40px; }
        #line-items-container .card { margin-bottom: 10px; background: var(--bg-secondary); border: 1px solid var(--border-color); }
        /* Dark theme toggle button */
        .theme-toggle {
            background: transparent;
            border: 1px solid var(--border-color);
            color: var(--text-muted);
            padding: 4px 10px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.85rem;
        }
        .theme-toggle:hover { border-color: var(--accent); color: var(--accent); }
        /* Light theme overrides */
        body.theme-light { background: #f8f9fa; color: #212529; }
        body.theme-light .navbar { background: var(--accent) !important; border-bottom: none; }
        body.theme-light .navbar-brand { color: #fff !important; }
        body.theme-light .nav-link { color: rgba(255,255,255,0.75) !important; }
        body.theme-light .nav-link:hover, body.theme-light .nav-link.active { color: #fff !important; }
        body.theme-light .card { background: #fff; border: 1px solid #dee2e6; }
        body.theme-light .card-header { background: var(--accent) !important; color: #fff; border-bottom: none; }
        body.theme-light .form-control, body.theme-light .form-select { background: #fff; color: #212529; border-color: #ced4da; }
        body.theme-light .form-control:focus { border-color: var(--accent); }
        body.theme-light .form-control::placeholder { color: #6c757d; }
        body.theme-light .form-label { color: #495057; }
        body.theme-light h5, body.theme-light h6, body.theme-light .mb-0 { color: #212529; }
        body.theme-light hr { border-color: #dee2e6; }
        body.theme-light .text-muted { color: #6c757d !important; }
        body.theme-light .theme-toggle { border-color: rgba(255,255,255,0.5); color: rgba(255,255,255,0.8); }
        body.theme-light .theme-toggle:hover { border-color: #fff; color: #fff; }
        body.theme-light #line-items-container .card { background: #fff; border: 1px solid #dee2e6; }
        /* Select2 dark theme overrides */
        .select2-container--default .select2-selection--single { background: var(--bg-secondary) !important; border-color: var(--border-color) !important; }
        .select2-container--default .select2-selection--single .select2-selection__rendered { color: var(--text-main) !important; }
        .select2-dropdown { background: var(--bg-secondary); border-color: var(--border-color); }
        .select2-container--default .select2-results__option { color: var(--text-main); }
        .select2-container--default .select2-search--dropdown .select2-search__field { background: var(--bg-primary); color: var(--text-main); border-color: var(--border-color); }
        body.theme-light .select2-container--default .select2-selection--single { background: #fff !important; border-color: #ced4da !important; }
        body.theme-light .select2-container--default .select2-selection--single .select2-selection__rendered { color: #212529 !important; }
        body.theme-light .select2-dropdown { background: #fff; border-color: #ced4da; }
        body.theme-light .select2-container--default .select2-results__option { color: #212529; }
        body.theme-light .select2-container--default .select2-search--dropdown .select2-search__field { background: #fff; color: #212529; border-color: #ced4da; }
        /* Snipe-IT link button */
        .snipe-link {
            display: inline-flex;
            align-items: center;
            gap: 4px;
            color: var(--accent);
            font-size: 0.85rem;
            text-decoration: none;
            padding: 4px 8px;
            border-radius: 4px;
        }
        .snipe-link:hover { color: var(--accent-hover); background: rgba(60, 140, 188, 0.1); }
    </style>
</head>
<body>
    <nav class="navbar navbar-expand-lg navbar-dark">
        <div class="container">
            <a class="navbar-brand" href="/">
                <img src="https://snipe-it.io/wp-content/uploads/2022/09/snipe-it-logo.svg" alt="snipe-it" onerror="this.style.display='none'">
                <span>&#128196; snipe-po</span>
            </a>
            <div class="navbar-nav">
                <a class="nav-link" href="/">Dashboard</a>
                <a class="nav-link active" href="/create">+ New PO</a>
            </div>
            <button class="theme-toggle" onclick="toggleTheme()">&#9788; Dark</button>
        </div>
    </nav>

    <div class="container mt-4">
        <div class="row">
            <div class="col-lg-8">
                <div class="card shadow-sm">
                    <div class="card-header">
                        <h5 class="mb-0">Create Purchase Order</h5>
                    </div>
                    <div class="card-body">
                        <form id="po-form" method="post" action="/api/pos">
                            <div class="row mb-3">
                                <div class="col-md-6">
                                    <label class="form-label">Supplier *</label>
                                    <select name="supplier" id="supplier-select" class="form-select" required>
                                        <option value="">Select or type supplier...</option>
                                    </select>
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
                            <div class="d-flex justify-content-between align-items-center mb-3">
                                <h5 class="mb-0">Line Items</h5>
                                <a href="{{.SnipeITWebURL}}/hardware/create" target="_blank" class="snipe-link">
                                    Create in Snipe-IT &#8594;
                                </a>
                            </div>
                            <div id="line-items-container">
                                <div class="line-item-row" data-index="0">
                                    <input type="text" name="items[0].description" placeholder="Description" required>
                                    <input type="text" name="items[0].model" placeholder="Model">
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
                                    <div class="form-control" id="grand_total"><strong>$0.00</strong></div>
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
                            <li>Suppliers are loaded from Snipe-IT manufacturers</li>
                        </ul>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script src="https://code.jquery.com/jquery-3.7.1.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/select2@4.1.0-rc.0/dist/js/select2.min.js"></script>
    <script>
        var itemIndex = 1;
        var snipeITURL = "{{.SnipeITURL}}";

        // Theme toggle
        function toggleTheme() {
            var body = document.body;
            var btn = document.querySelector('.theme-toggle');
            if (body.classList.contains('theme-light')) {
                body.classList.remove('theme-light');
                localStorage.setItem('theme', 'dark');
                btn.innerHTML = '&#9788; Dark';
            } else {
                body.classList.add('theme-light');
                localStorage.setItem('theme', 'light');
                btn.innerHTML = '&#9790; Light';
            }
        }
        (function() {
            var theme = localStorage.getItem('theme');
            if (theme === 'light') {
                document.body.classList.add('theme-light');
                document.querySelector('.theme-toggle').innerHTML = '&#9790; Light';
            }
        })();

        // Initialize supplier dropdown with Select2
        $(document).ready(function() {
            var $select = $('#supplier-select').select2({
                ajax: {
                    url: '/api/suppliers',
                    dataType: 'json',
                    delay: 250,
                    processResults: function(data) {
                        return {
                            results: data.map(function(item) {
                                return { id: item.name, text: item.name };
                            })
                        };
                    },
                    cache: true
                },
                placeholder: 'Select or type supplier...',
                allowClear: true,
                tags: true,
                tokenSeparators: [',']
            });
        });

        function addLineItem() {
            var container = document.getElementById("line-items-container");
            var div = document.createElement("div");
            div.className = "line-item-row";
            div.setAttribute("data-index", itemIndex);
            div.innerHTML = "<input type=\"text\" name=\"items[" + itemIndex + "].description\" placeholder=\"Description\" required>" +
                "<input type=\"text\" name=\"items[" + itemIndex + "].model\" placeholder=\"Model\">" +
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
            if (e.target.name && (e.target.name.includes(".unit_price") || e.target.name.includes(".quantity"))) {
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
    <title>snipe-po - View PO #{{.PO.PONumber}}</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css" rel="stylesheet">
    <style id="theme-styles">
        :root {
            --bg-primary: #2b2b2b;
            --bg-secondary: #3d3d3d;
            --accent: #3c8dbc;
            --accent-hover: #337ab7;
            --text-main: #eeeeee;
            --text-muted: #aaaaaa;
            --border-color: #444444;
            --card-bg: #323232;
            --table-bg: #3d3d3d;
            --light-text: #ffffff;
        }
        body { background: var(--bg-primary); color: var(--text-main); }
        .navbar { background: var(--bg-secondary) !important; border-bottom: 1px solid var(--border-color); }
        .navbar-brand { font-weight: 600; color: var(--text-main) !important; display: flex; align-items: center; gap: 8px; }
        .navbar-brand img { height: 28px; }
        .nav-link { color: var(--text-muted) !important; }
        .nav-link:hover, .nav-link.active { color: var(--text-main) !important; }
        .card { background: var(--card-bg); border: 1px solid var(--border-color); }
        .card-header { background: var(--bg-secondary) !important; border-bottom: 1px solid var(--border-color); color: var(--text-main); }
        .table { color: var(--text-main); }
        .table-light { background: var(--table-bg) !important; color: var(--text-main) !important; }
        .card-body { color: var(--text-main); }
        .card-body strong { color: var(--text-main); }
        thead { background: var(--table-bg) !important; color: var(--text-main) !important; }
        thead th { background: var(--table-bg) !important; color: var(--text-main) !important; border-color: var(--border-color) !important; }
        tfoot { background: var(--table-bg) !important; color: var(--text-main) !important; }
        tfoot td { background: var(--table-bg) !important; color: var(--text-main) !important; border-color: var(--border-color) !important; }
        .form-control { background: var(--bg-secondary); color: var(--text-main); border-color: var(--border-color); }
        .form-label { color: var(--text-muted); }
        .status-badge { text-transform: capitalize; }
        .status-draft { background: #6c757d; }
        .status-pending_approval { background: #ffc107; color: #000; }
        .status-approved { background: #198754; }
        .status-sent_to_ap { background: #0d6efd; }
        .status-paid { background: #17a2b8; }
        .status-rejected { background: #dc3545; }
        .amount { font-family: monospace; }
        .history-timeline { font-size: 0.9em; }
        .text-muted { color: var(--text-muted) !important; }
        h2, h5, h6, .mb-0 { color: var(--text-main); }
        /* Dark theme toggle button */
        .theme-toggle {
            background: transparent;
            border: 1px solid var(--border-color);
            color: var(--text-muted);
            padding: 4px 10px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.85rem;
        }
        .theme-toggle:hover { border-color: var(--accent); color: var(--accent); }
        /* Light theme overrides */
        body.theme-light { background: #f8f9fa; color: #212529; }
        body.theme-light .navbar { background: var(--accent) !important; border-bottom: none; }
        body.theme-light .navbar-brand { color: #fff !important; }
        body.theme-light .nav-link { color: rgba(255,255,255,0.75) !important; }
        body.theme-light .nav-link:hover, body.theme-light .nav-link.active { color: #fff !important; }
        body.theme-light .card { background: #fff; border: 1px solid #dee2e6; }
        body.theme-light .card-header { background: var(--accent) !important; color: #fff; border-bottom: none; }
        body.theme-light .table { color: #212529; }
        body.theme-light .table-light { background: #f8f9fa !important; color: #212529 !important; }
        body.theme-light .card-body { color: #212529; }
        body.theme-light .card-body strong { color: #212529; }
        body.theme-light thead { background: #f8f9fa !important; color: #212529 !important; }
        body.theme-light thead th { background: #f8f9fa !important; color: #212529 !important; }
        body.theme-light tfoot { background: #f8f9fa !important; color: #212529 !important; }
        body.theme-light tfoot td { background: #f8f9fa !important; color: #212529 !important; }
        body.theme-light .form-control { background: #fff; color: #212529; border-color: #ced4da; }
        body.theme-light .form-label { color: #495057; }
        body.theme-light h2, body.theme-light h5, body.theme-light h6, body.theme-light .mb-0 { color: #212529; }
        body.theme-light .text-muted { color: #6c757d !important; }
        body.theme-light .theme-toggle { border-color: rgba(255,255,255,0.5); color: rgba(255,255,255,0.8); }
        body.theme-light .theme-toggle:hover { border-color: #fff; color: #fff; }
    </style>
</head>
<body>
    <nav class="navbar navbar-expand-lg navbar-dark">
        <div class="container">
            <a class="navbar-brand" href="/">
                <img src="https://snipe-it.io/wp-content/uploads/2022/09/snipe-it-logo.svg" alt="snipe-it" onerror="this.style.display='none'">
                <span>&#128196; snipe-po</span>
            </a>
            <div class="navbar-nav">
                <a class="nav-link" href="/">Dashboard</a>
                <a class="nav-link" href="/create">+ New PO</a>
            </div>
            <button class="theme-toggle" onclick="toggleTheme()">&#9788; Dark</button>
        </div>
    </nav>

    <div class="container mt-4">
        <div class="d-flex justify-content-between align-items-center mb-3">
            <h2>PO #{{.PO.PONumber}}</h2>
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
                    <div class="card-header">
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
                        <form method="post" action="/api/pos/{{.PO.ID}}/send">
                            <button type="submit" class="btn btn-outline-secondary w-100">&#128228; Resend Approval Email</button>
                        </form>
                        {{end}}

                        {{if eq .PO.Status "approved"}}
                        <form method="post" action="/api/pos/{{.PO.ID}}/send-to-ap">
                            <button type="submit" class="btn btn-primary w-100">&#128228; Send to AP</button>
                        </form>
                        {{end}}

                        {{if eq .PO.Status "sent_to_ap"}}
                        <div class="mb-2">
                            <form method="post" action="/api/pos/{{.PO.ID}}/mark-paid" enctype="multipart/form-data">
                                <label class="form-label small text-muted">Upload Packing Slip (PDF)</label>
                                <div class="input-group">
                                    <input type="file" name="packing_slip" accept=".pdf" class="form-control" required>
                                    <button type="submit" class="btn btn-info">Upload</button>
                                </div>
                            </form>
                        </div>
                        {{end}}

                        {{if .PO.PackingSlipPath}}
                        <a href="/api/pos/{{.PO.ID}}/packing-slip" class="btn btn-outline-primary btn-sm w-100 mb-2">&#128196; Download Packing Slip</a>
                        {{end}}

                        {{if eq .PO.Status "draft"}}
                        <form method="post" action="/api/pos/{{.PO.ID}}/delete" onsubmit="return confirm('Delete this PO?');">
                            <button type="submit" class="btn btn-danger w-100">&#128465; Delete PO</button>
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

    <script>
        // Theme toggle
        function toggleTheme() {
            var body = document.body;
            var btn = document.querySelector('.theme-toggle');
            if (body.classList.contains('theme-light')) {
                body.classList.remove('theme-light');
                localStorage.setItem('theme', 'dark');
                btn.innerHTML = '&#9788; Dark';
            } else {
                body.classList.add('theme-light');
                localStorage.setItem('theme', 'light');
                btn.innerHTML = '&#9790; Light';
            }
        }
        (function() {
            var theme = localStorage.getItem('theme');
            if (theme === 'light') {
                document.body.classList.add('theme-light');
                document.querySelector('.theme-toggle').innerHTML = '&#9790; Light';
            }
        })();
    </script>
</body>
</html>`
