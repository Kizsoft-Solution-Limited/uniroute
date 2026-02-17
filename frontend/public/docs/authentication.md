# Authentication

UniRoute supports multiple authentication methods for both the CLI and API.

## CLI Authentication

UniRoute CLI supports two login methods: **Email/Password** and **API Key**. Both methods authenticate you with the UniRoute service and allow you to create and manage tunnels.

### Email/Password Login

Standard login with session expiration:

```bash
# Login to managed service
uniroute auth login

# Login with email flag
uniroute auth login --email user@example.com

# Login to self-hosted instance
uniroute auth login --server http://localhost:8084
```

This will:
1. Prompt you for your email and password
2. Authenticate with the UniRoute server
3. Save your authentication token (JWT) with expiration
4. Associate tunnels with your account

**Note:** JWT tokens expire after a set period. You'll need to log in again when the token expires.

### API Key Login (Recommended for Automation)

API keys provide longer sessions without expiration, making them ideal for automation, CI/CD pipelines, and scripts:

```bash
# Login with API key
uniroute auth login --api-key ur_xxxxxxxxxxxxx

# Or using short flag
uniroute auth login -k ur_xxxxxxxxxxxxx

# Login to self-hosted instance with API key
uniroute auth login --api-key ur_xxxxxxxxxxxxx --server http://localhost:8084
```

**Benefits of API Key Login:**
- ✅ **No expiration** - Unlike JWT tokens, API keys don't expire
- ✅ **Perfect for automation** - Ideal for scripts and CI/CD pipelines
- ✅ **Longer sessions** - No need to re-authenticate frequently
- ✅ **Same functionality** - Full access to all CLI features

**When to use API Key Login:**
- Running in CI/CD pipelines
- Automated scripts and tools
- Long-running processes
- When you want to avoid frequent re-authentication

**Getting an API Key:**
You can create API keys from the dashboard or using the CLI:
```bash
# List your API keys
uniroute keys list

# Create a new API key
uniroute keys create --name "My Automation Key"
```

### Logout

```bash
uniroute auth logout
```

### Check Status

```bash
uniroute auth status
```

## API Authentication

### API Keys

Generate an API key from the dashboard or CLI:

```bash
# List your API keys
uniroute keys list

# Create a new API key
uniroute keys create --name "My App"
```

### Using API Keys

Include your API key in the `Authorization` header:

```bash
curl -X POST https://app.uniroute.co/v1/chat \
  -H "Authorization: Bearer ur_your-api-key" \
  -H "Content-Type: application/json" \
  -d '{...}'
```

### JWT Tokens

For web applications, use JWT tokens obtained from the login endpoint:

```bash
# Login and get JWT token
curl -X POST https://app.uniroute.co/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password"}'

# Use JWT token
curl -X GET https://app.uniroute.co/auth/tunnels \
  -H "Authorization: Bearer your-jwt-token"
```

## OAuth Providers

UniRoute supports OAuth login with:

- **Google** - Sign in with Google account
- **X (Twitter)** - Sign in with X account

OAuth login is available through the web dashboard.

## Security Best Practices

1. **Never commit API keys** to version control
2. **Rotate keys regularly** for production applications
3. **Use environment variables** to store credentials
4. **Enable IP whitelisting** for production API keys
5. **Monitor key usage** in the dashboard

## Next Steps

- [Tunnels](/docs/tunnels) - Start creating tunnels
- [API Reference](/docs/api) - Make authenticated API requests
- [Security](/docs/security) - Learn about security features
