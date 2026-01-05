# 🔐 CLI Authentication Design

## Current Implementation

UniRoute CLI uses **email/password authentication** for login, which then stores a JWT token locally for subsequent commands.

### How It Works

1. **Login**: User runs `uniroute auth login` with email/password
2. **Token Storage**: JWT token is saved to `~/.uniroute/auth.json`
3. **Automatic Auth**: Subsequent commands automatically use the stored token
4. **Session Management**: Token can expire and be refreshed

---

## Why Email/Password is Better for CLI Login

### ✅ Advantages of Email/Password

1. **User-Friendly**
   - Familiar pattern (users know their email/password)
   - No need to generate API keys first
   - Better for interactive CLI sessions
   - Matches industry standards (GitHub CLI, AWS CLI, Heroku CLI)

2. **Security Benefits**
   - Password can be reset if forgotten
   - Session tokens can expire automatically
   - Can implement MFA in the future
   - Better audit trail (know who logged in)
   - Can revoke all sessions if password is compromised

3. **Better UX**
   - Interactive prompts for credentials
   - Clear login/logout flow
   - Status command shows current user
   - Can support "remember me" functionality

4. **Flexibility**
   - Can support OAuth/SSO in the future
   - Can add passwordless login (magic links)
   - Can implement device trust

### ❌ Disadvantages of API Key for Login

1. **User Experience**
   - Users must generate keys first (extra step)
   - Keys are long strings, hard to remember
   - Not intuitive for interactive use
   - Keys can be lost/forgotten

2. **Security Concerns**
   - Harder to revoke if compromised
   - No expiration by default
   - No session management
   - Keys often stored insecurely

3. **Limited Features**
   - Can't implement MFA
   - No password reset flow
   - Harder to track user activity
   - No "remember me" concept

---

## Recommended Approach: Hybrid Model

### Primary: Email/Password (Current Implementation) ✅

**For Interactive CLI Use:**
```bash
# Login with email/password
uniroute auth login
# Email: user@example.com
# Password: ********

# Token automatically saved, all commands work
uniroute keys list
uniroute tunnel --port 8080
```

**Benefits:**
- ✅ User-friendly
- ✅ Secure (token-based after login)
- ✅ Session management
- ✅ Better for interactive use

### Secondary: API Key Support (For Automation)

**For CI/CD and Automation:**
```bash
# Use API key directly (no login needed)
uniroute keys create --jwt-token YOUR_API_KEY
uniroute tunnel --token YOUR_API_KEY --port 8080
```

**Benefits:**
- ✅ Good for automation
- ✅ No interactive prompts
- ✅ Works in CI/CD pipelines
- ✅ Can be stored in environment variables

---

## Current Implementation Details

### Authentication Flow

```
1. User runs: uniroute auth login
   ↓
2. Prompts for email/password (or via flags)
   ↓
3. Sends POST /auth/login
   ↓
4. Receives JWT token
   ↓
5. Saves token to ~/.uniroute/auth.json
   ↓
6. All subsequent commands use stored token
```

### Token Storage

**Location**: `~/.uniroute/auth.json`

**Format**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "email": "user@example.com",
  "server_url": "https://api.uniroute.dev",
  "expires_at": "2024-12-27T12:00:00Z"
}
```

**Security**:
- File permissions: `0600` (read/write for owner only)
- Token is JWT (can expire)
- Can be revoked server-side

### Commands That Use Auth

- ✅ `uniroute auth login` - Login with email/password
- ✅ `uniroute auth logout` - Clear stored token
- ✅ `uniroute auth status` - Show current auth status
- ✅ `uniroute keys list` - Requires auth (uses stored token)
- ✅ `uniroute keys revoke` - Requires auth (uses stored token)
- ✅ `uniroute tunnel` - Requires auth for public servers
- ✅ `uniroute projects list` - Requires auth (uses stored token)

---

## Comparison with Industry Standards

### GitHub CLI (`gh`)
- ✅ Uses email/password for login
- ✅ Stores OAuth token after login
- ✅ Token used for all subsequent commands
- ✅ `gh auth login` → `gh auth status` → `gh auth logout`

### AWS CLI (`aws`)
- ✅ Uses access keys (similar to API keys)
- ⚠️ But also supports SSO with email/password
- ✅ Stores credentials in `~/.aws/credentials`

### Heroku CLI (`heroku`)
- ✅ Uses email/password for login
- ✅ Stores API token after login
- ✅ Token used for all commands

### Google Cloud CLI (`gcloud`)
- ✅ Uses email/password for login
- ✅ Stores OAuth token after login
- ✅ Supports SSO

**Conclusion**: Most modern CLIs use email/password for login, then store tokens. This is the industry standard.

---

## Future Enhancements

### 1. Password Masking (Current TODO)
```go
// Current: Password visible in terminal
fmt.Scanln(&authPassword)

// Future: Use golang.org/x/term for hidden input
import "golang.org/x/term"
password, _ := term.ReadPassword(int(os.Stdin.Fd()))
```

### 2. Token Refresh
- Automatically refresh expired tokens
- Seamless re-authentication
- Better user experience

### 3. MFA Support
- Two-factor authentication
- TOTP codes
- SMS/Email verification

### 4. OAuth/SSO Support
- Login with GitHub/Google
- Enterprise SSO
- SAML support

### 5. API Key as Alternative
- Allow `uniroute auth login --api-key KEY`
- For automation use cases
- Still store as token for consistency

### 6. Multiple Accounts
- Support multiple server URLs
- Switch between accounts
- `uniroute auth switch`

---

## Security Best Practices

### Current Implementation ✅

1. **Token Storage**
   - ✅ Stored in user's home directory
   - ✅ File permissions: `0600` (owner only)
   - ✅ JSON format (easy to parse/validate)

2. **Token Usage**
   - ✅ Sent as Bearer token in Authorization header
   - ✅ Not logged or exposed
   - ✅ Can expire server-side

3. **Password Handling**
   - ✅ Hidden in terminal (masked input)
   - ✅ Not stored locally
   - ✅ Sent over HTTPS only
   - ✅ Cross-platform compatible

### ✅ Implemented Improvements

1. **Password Masking** ✅
   - ✅ Implemented using `golang.org/x/term`
   - ✅ Password input is now hidden in terminal
   - ✅ Cross-platform compatible (macOS, Linux, Windows)
   - ✅ Properly handles newline after password input

2. **Token Encryption**
   - Encrypt token at rest (optional)
   - Use OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)

3. **Token Validation**
   - Check token expiration before use
   - Auto-refresh if expired
   - Clear invalid tokens

4. **Audit Logging**
   - Log login attempts
   - Track token usage
   - Monitor for suspicious activity

---

## Usage Examples

### Interactive Login
```bash
# Prompt for credentials
$ uniroute auth login
Email: user@example.com
Password: ********
✅ Successfully logged in!
   Email: user@example.com
   Server: https://api.uniroute.dev
```

### Non-Interactive Login
```bash
# Use flags (for automation)
$ uniroute auth login --email user@example.com --password mypassword
✅ Successfully logged in!
```

### Check Status
```bash
$ uniroute auth status
✅ Logged in
   Email: user@example.com
   Server: https://api.uniroute.dev
   Expires: 2024-12-27T12:00:00Z
```

### Use Commands (Auto-authenticated)
```bash
# Token automatically used
$ uniroute keys list
$ uniroute tunnel --port 8080
$ uniroute projects list
```

### Logout
```bash
$ uniroute auth logout
✅ Successfully logged out
```

---

## API Key Alternative (For Automation)

While email/password is recommended for interactive use, API keys are still useful for automation:

### Use Cases for API Keys

1. **CI/CD Pipelines**
   ```yaml
   # GitHub Actions
   - name: Create API Key
     run: |
       uniroute keys create --jwt-token ${{ secrets.UNIROUTE_API_KEY }}
   ```

2. **Scripts**
   ```bash
   # Use API key directly
   export UNIROUTE_TOKEN="ur_your-api-key"
   uniroute keys list --jwt-token $UNIROUTE_TOKEN
   ```

3. **Non-Interactive Environments**
   - Docker containers
   - Kubernetes jobs
   - Scheduled tasks

### Implementation

The CLI already supports this:
```bash
# Use --jwt-token flag for any command
uniroute keys create --jwt-token YOUR_API_KEY
uniroute keys list --jwt-token YOUR_API_KEY
uniroute tunnel --token YOUR_API_KEY --port 8080
```

---

## Summary

### ✅ Recommended: Email/Password for Login

**Why:**
- Better user experience
- Industry standard
- More secure (session management)
- Better for interactive use
- Supports future features (MFA, SSO)

**Current Status:**
- ✅ Already implemented
- ✅ Works well
- ✅ Password masking implemented (hidden input)

### ✅ Also Support: API Keys for Automation

**Why:**
- Good for CI/CD
- Non-interactive use
- Automation scripts

**Current Status:**
- ✅ Already supported via `--jwt-token` flag
- ✅ Works for all commands

### 🎯 Best Practice

**Use email/password for:**
- Interactive CLI sessions
- First-time setup
- User authentication

**Use API keys for:**
- CI/CD pipelines
- Automation scripts
- Non-interactive environments

**Both approaches are supported and serve different use cases!**

---

## Conclusion

**Email/password is the better choice for CLI login** because:
1. ✅ More user-friendly
2. ✅ Industry standard
3. ✅ Better security features
4. ✅ Supports future enhancements
5. ✅ Already implemented and working

**API keys are still useful** for:
1. ✅ Automation and CI/CD
2. ✅ Non-interactive use
3. ✅ Scripts and tools

The current implementation supports both, which is the ideal approach! 🎉

