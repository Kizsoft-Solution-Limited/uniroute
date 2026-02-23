# Custom Domains

Use your own domain instead of random subdomains with UniRoute. Manage domains through the CLI or web dashboard.

## Overview

Custom domains allow you to use your own domain name for tunnels, giving you:
- Professional URLs
- Brand consistency
- Automatic SSL certificates
- Full control over DNS
- Centralized domain management

## Adding a Custom Domain

### Add Domain to Account

You can add a domain to your account first, then assign it to tunnels later:

```bash
# Add domain to your account (not assigned to any tunnel yet)
uniroute domain example.com
```

This creates the domain in your account. You can view and manage it in the dashboard at `/dashboard/domains`.

### Add and Assign in One Command

You can also add a domain and assign it to a tunnel in one command:

```bash
# Add domain AND assign to tunnel by subdomain (shortcut - recommended)
uniroute domain example.com abc123

# Add domain AND assign to tunnel (flag syntax)
uniroute domain example.com --subdomain abc123

# Add domain AND assign to specific tunnel by ID
uniroute domain example.com --tunnel-id <tunnel-id>
```

If you don't specify a tunnel, the domain will be assigned to your last active tunnel (if available).

## Domain Management Commands

UniRoute provides comprehensive domain management through the CLI:

### List All Domains

View all your custom domains and their status:

```bash
uniroute domain list
```

This shows:
- Domain names
- DNS configuration status
- Domain IDs
- CNAME setup instructions

### Show Domain Details

Get detailed information about a specific domain:

```bash
uniroute domain show example.com
```

Shows:
- Domain name and ID
- DNS configuration status
- Setup instructions if not configured

### Verify DNS Configuration

Check if your DNS is properly configured:

```bash
uniroute domain verify example.com
```

This will:
- Check if CNAME record is correctly set
- Verify DNS resolution
- Update domain status in your account

### Resume Domain Assignment

Resume a previously saved domain-to-tunnel assignment:

```bash
# Resume last used domain assignment
uniroute domain resume

# Resume by subdomain
uniroute domain resume abc123

# Resume by domain name
uniroute domain resume example.com
```

**How Resume Works:**
- When you assign a domain to a tunnel, the assignment is automatically saved
- You can resume it later using the domain name or subdomain
- The system automatically looks up the current tunnel ID
- Works across CLI sessions (persistent storage)

**Example:**
```bash
# First time: assign domain to tunnel
uniroute domain billspot.co abc123

# Later: resume the same assignment
uniroute domain resume abc123
# or
uniroute domain resume billspot.co
```

### Remove Domain

Remove a domain from your account:

```bash
uniroute domain remove example.com
```

This will:
- Remove the domain from your account
- Unassign it from any tunnels
- Require confirmation before deletion

## DNS Configuration

After adding a domain, you need to configure DNS in your provider (Cloudflare, Namecheap, GoDaddy, etc.):

### Step 1: Add DNS Records

Add these two records (replace `example.com` with your domain):

| Type        | Host | Value / Target      | TTL        |
|------------|------|---------------------|------------|
| **A Record**   | `@`  | `75.119.141.27`     | Automatic  |
| **CNAME Record** | `www` | `example.com.`      | Automatic  |

- **A Record:** Host `@` (root/apex), IP Address `75.119.141.27`, TTL Automatic.
- **CNAME Record:** Host `www`, Target your apex domain with a trailing dot (e.g. `example.com.`), TTL Automatic.

To confirm the current tunnel server IP: `dig tunnel.uniroute.co +short`

### Step 2: Verify DNS

After configuring DNS, verify it:

```bash
uniroute domain verify example.com
```

Or use the dashboard:
1. Go to `/dashboard/domains`
2. Find your domain
3. Click "Verify DNS" button

### Step 3: Wait for Propagation

DNS changes can take a few minutes to propagate. You can check with:

```bash
dig example.com
# or
nslookup example.com
```

## SSL Certificate

UniRoute automatically provisions SSL certificates for your custom domain using Let's Encrypt. This happens automatically once DNS is configured correctly and verified.

## Domain Management in Dashboard

You can also manage domains through the web dashboard:

- **View all domains**: Navigate to `/dashboard/domains`
- **Add new domains**: Click "Add Domain" button
- **Verify DNS**: Click "Verify DNS" button for each domain
- **Delete domains**: Click delete button (with confirmation)

**Important:** Both CLI and dashboard use the same backend system. Domains created via CLI appear in the dashboard and vice versa.

## Domain Validation

UniRoute validates your domain by:
1. Checking DNS resolution
2. Verifying CNAME record points to `tunnel.uniroute.co`
3. Confirming SSL certificate can be issued
4. Updating domain status (`dns_configured` flag)

## Workflow Examples

### Example 1: Add Domain First, Assign Later

```bash
# Step 1: Add domain to account
uniroute domain example.com

# Step 2: Configure DNS in your DNS provider
# (Add CNAME: example.com → tunnel.uniroute.co)

# Step 3: Verify DNS
uniroute domain verify example.com

# Step 4: Assign to tunnel
uniroute domain example.com abc123
```

### Example 2: Add and Assign in One Step

```bash
# Add domain and assign to tunnel
uniroute domain example.com abc123

# Configure DNS
# (Add CNAME: example.com → tunnel.uniroute.co)

# Verify DNS
uniroute domain verify example.com
```

### Example 3: Resume Previous Assignment

```bash
# Resume domain assignment (saved from previous session)
uniroute domain resume abc123

# Or resume by domain name
uniroute domain resume example.com
```

## Troubleshooting

### Domain Not Resolving

- Check DNS records are correct
- Wait for DNS propagation (can take up to 48 hours)
- Verify domain is pointing to `tunnel.uniroute.co`
- Use `uniroute domain verify example.com` to check status

### SSL Certificate Issues

- Ensure DNS is correctly configured
- Wait a few minutes for certificate provisioning
- Verify domain status: `uniroute domain show example.com`
- Check domain validation status in dashboard

### Domain Not Found

If you get "domain not found" errors:
- Make sure you've added the domain: `uniroute domain list`
- Check spelling of domain name
- Verify you're logged in: `uniroute auth login`

### Resume Not Working

If resume doesn't work:
- Make sure you previously assigned the domain to a tunnel
- Check if tunnel still exists: `uniroute tunnel --list`
- Try assigning again: `uniroute domain example.com abc123`

## Next Steps

- [Reserved Subdomains](/docs/tunnels/reserved-subdomains) - Reserve subdomains for your account
- [Tunnels Overview](/docs/tunnels) - Learn more about tunnels
- [Tunnel Resume](/docs/tunnels/resume) - Learn about tunnel resume functionality
