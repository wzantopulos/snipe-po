# MCP Server - Snipe-PO

Model Context Protocol server for managing snipe-po purchase order system.

## Overview

This MCP server provides tools for managing the snipe-po application including:
- Checking build/deployment status
- Managing application settings
- Viewing logs and troubleshooting
- Managing the Git workflow

## Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Access to the snipe-po GitHub repository
- SSH key configured for GitHub (`/root/.ssh/id_ed25519_snipe_po`)

## Available Tools

### check_build_status
Check the status of the latest GitHub Actions build.

### get_recent_builds
Get the status of the most recent builds.

### restart_services
Restart the snipe-po Docker services.

### view_logs
View recent Docker logs for snipe-po.

### update_and_deploy
Pull latest changes, rebuild and deploy.

## System Preferences

### User Interface Conventions

1. **Form Actions Must Redirect** - All POST form submissions (buttons that modify data) must redirect to the `/view?id=PO_ID` page, NOT return JSON. The user should see the updated PO status, not raw JSON.

2. **Status Notifications** - When a user clicks a button that performs an action (Send for Approval, Approve, Reject, Send to AP, Mark Paid), they should see the result page with updated status, not a raw JSON response.

3. **Handler Pattern** - For any handler that modifies PO state, use:
   ```go
   http.Redirect(w, r, "/view?id="+po.ID, http.StatusSeeOther)
   ```
   NOT:
   ```go
   jsonResponse(w, po, http.StatusOK)
   ```

4. **Error Handling** - Errors should be shown as user-friendly messages or logs, not raw JSON to the browser.

### Email Configuration

- **Port 25** - Plaintext SMTP (no TLS), common for IP-based SMTP relays
- **Port 587** - STARTTLS (explicit TLS upgrade after connection)
- **Port 465** - Implicit TLS (direct TLS connection)

The email code (`email/email.go`) handles all three modes based on the configured port.

### Snipe-IT Integration

- **Suppliers vs Manufacturers** - Snipe-IT has both suppliers (companies you buy from) and manufacturers (companies that make products). The app fetches **suppliers** from `/api/v1/suppliers`.
- **Internal URL** - From inside Docker, use `http://snipeit:80` instead of the public URL.
- **Live Fetch** - Supplier list is fetched live on each page load, not cached.

### Snipe-PO Status Flow

```
draft -> pending_approval -> approved -> sent_to_ap -> paid
                |
                v
             rejected
```

## Deployment

```bash
# Pull latest and restart
sudo docker compose pull && sudo docker compose up -d

# View logs
sudo docker compose logs -f snipe-po

# Restart services
sudo docker compose restart snipe-po
```

## Configuration

Settings file is typically at `/root/snipe-po/settings.yaml` or mounted via Docker volume.

Key settings:
- `smtp.port` - SMTP port (25, 587, or 465)
- `snipe_it.url` - Snipe-IT URL (used for API calls from inside Docker)
- `po.number_prefix` - PO number prefix (e.g., "IT-")

## Common Issues

### Supplier dropdown empty
- Check if the app can reach Snipe-IT from inside Docker network
- Verify using `http://snipeit:80` for internal calls
- Check container logs for API errors

### Emails not sending
- Check SMTP port matches relay requirements
- Port 25 = plaintext, no TLS
- Port 587 = STARTTLS
- Port 465 = implicit TLS
- View logs: `sudo docker compose logs snipe-po | grep -i email`

### Build failures
- Check GitHub Actions for error details
- Try building locally: `go build ./...`
