# Reserved Subdomains

Request specific subdomains for your tunnels.

## Overview

Instead of random subdomains, you can request specific subdomains like `myapp` instead of `abc123`.

## Requesting a Subdomain

### When Creating a Tunnel

```bash
# Request a specific subdomain (shortcut syntax - recommended)
uniroute http 8080 myapp
uniroute http 8080 myapp --new
uniroute tcp 3306 mydb
uniroute tcp 3306 mydb --new
uniroute tls 5432 mydb
uniroute tls 5432 mydb --new
uniroute udp 53 dns
uniroute udp 53 dns --new

# Request a specific subdomain (flag syntax - also works)
uniroute http 8080 --host myapp
uniroute http 8080 --host myapp --new
uniroute tcp 3306 --host mydb
uniroute tls 5432 --host mydb
```

### Availability

Subdomains are allocated on a first-come, first-served basis. If a subdomain is already taken, you'll receive an error:

```
Error: subdomain 'myapp' is not available
```

## Subdomain Rules

- Must be alphanumeric (a-z, 0-9) and hyphens (-)
- Cannot start or end with a hyphen
- Maximum 63 characters
- Must be unique across all users

## Reserved Subdomains

Some subdomains are reserved for system use and cannot be requested:
- `www`
- `api`
- `app`
- `admin`
- `dashboard`
- `docs`

## Best Practices

1. **Use descriptive names** - `myapp` is better than `app1`
2. **Check availability first** - Try creating a tunnel to see if it's available
3. **Use consistent naming** - Keep naming consistent across your tunnels
4. **Reserve early** - Popular subdomains get taken quickly

## Next Steps

- [Custom Domains](/docs/tunnels/custom-domains) - Use your own domain
- [Opening a Tunnel](/docs/tunnels/opening) - Create tunnels
