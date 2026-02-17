# Security

UniRoute implements enterprise-grade security measures.

## Authentication

### API Keys

- Bcrypt hashed storage
- Scoped permissions
- Rate limiting per key
- IP whitelisting support

### JWT Tokens

- Secure token generation
- Expiration and refresh
- Role-based access control

## Rate Limiting

UniRoute implements progressive rate limiting:

- **Per API Key**: Limits requests per key
- **Per IP**: Limits requests per IP address
- **Progressive**: Stricter limits for unauthenticated requests

## Data Protection

- **Encryption at Rest**: Sensitive data encrypted in database
- **Encryption in Transit**: All connections use TLS/HTTPS
- **Input Validation**: All inputs validated and sanitized
- **SQL Injection Prevention**: Parameterized queries only

## Security Headers

UniRoute sets security headers on all responses:

- `Content-Security-Policy`
- `Strict-Transport-Security`
- `X-Content-Type-Options`
- `X-Frame-Options`

## Best Practices

1. **Use HTTPS** - Always use HTTPS in production
2. **Rotate Keys** - Regularly rotate API keys
3. **IP Whitelisting** - Restrict API key access by IP
4. **Monitor Usage** - Regularly review API usage logs
5. **Keep Updated** - Keep UniRoute updated to latest version

## Reporting Security Issues

If you discover a security vulnerability, please email security@uniroute.co instead of opening a public issue.

## Next Steps

- [Authentication](/docs/authentication) - Set up secure authentication
- [Deployment](/docs/deployment) - Secure deployment practices
