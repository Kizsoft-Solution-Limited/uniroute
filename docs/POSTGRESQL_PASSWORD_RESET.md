# 🔐 PostgreSQL Password Reset Guide

## Problem: Password Authentication Failed

If you're getting this error:
```
FATAL: password authentication failed for user "postgres"
```

This means the password in your `.env` file doesn't match your PostgreSQL password.

---

## Solution 1: Reset PostgreSQL Password (Recommended)

### Step 1: Stop PostgreSQL
```bash
brew services stop postgresql@14
```

### Step 2: Start PostgreSQL in Single-User Mode
```bash
# Find your PostgreSQL data directory
PGDATA=$(brew --prefix)/var/postgresql@14

# Start PostgreSQL in single-user mode (no password required)
postgres --single -D "$PGDATA" postgres
```

### Step 3: Reset Password
In the PostgreSQL prompt, run:
```sql
ALTER USER postgres WITH PASSWORD 'postgres';
\q
```

### Step 4: Restart PostgreSQL
```bash
brew services start postgresql@14
```

### Step 5: Test Connection
```bash
psql -U postgres -h localhost -c "SELECT version();"
# Enter password: postgres
```

---

## Solution 2: Use Your Current Password

If you know your PostgreSQL password but it's different from `postgres`:

### Update `.env` File
```bash
# Edit .env file
DATABASE_URL=postgres://postgres:YOUR_ACTUAL_PASSWORD@localhost:5432/uniroute?sslmode=disable
```

---

## Solution 3: Create New User (Alternative)

If you can't reset the postgres password, create a new user:

### Step 1: Connect as postgres (if possible)
```bash
# Try connecting - you'll be prompted for password
psql -U postgres -h localhost
```

### Step 2: Create New User
```sql
CREATE USER uniroute_user WITH PASSWORD 'uniroute_password';
ALTER USER uniroute_user CREATEDB;
GRANT ALL PRIVILEGES ON DATABASE uniroute TO uniroute_user;
\q
```

### Step 3: Update `.env` File
```bash
DATABASE_URL=postgres://uniroute_user:uniroute_password@localhost:5432/uniroute?sslmode=disable
```

---

## Solution 4: Use Trust Authentication (Development Only)

⚠️ **Warning**: Only for local development! Never use in production!

### Step 1: Edit `pg_hba.conf`
```bash
# Find pg_hba.conf
PGDATA=$(brew --prefix)/var/postgresql@14
nano "$PGDATA/pg_hba.conf"
```

### Step 2: Change Authentication Method
Find this line:
```
host    all             all             127.0.0.1/32            scram-sha-256
```

Change to:
```
host    all             all             127.0.0.1/32            trust
```

### Step 3: Restart PostgreSQL
```bash
brew services restart postgresql@14
```

### Step 4: Test Connection (No Password Required)
```bash
psql -U postgres -h localhost
```

---

## Quick Fix: Try Common Passwords

Sometimes PostgreSQL is installed with a default password. Try these:

```bash
# Try empty password
psql -U postgres -h localhost -W
# Press Enter when prompted for password

# Try common passwords
psql -U postgres -h localhost
# Then try: postgres, admin, password, (empty)
```

---

## Verify PostgreSQL is Running

```bash
# Check if PostgreSQL is running
brew services list | grep postgres

# Check PostgreSQL process
ps aux | grep postgres | grep -v grep

# Check if port 5432 is listening
lsof -i :5432
```

---

## For pgAdmin Connection

After fixing the password, use these settings in pgAdmin:

**Connection Tab:**
- Host name/address: `localhost`
- Port: `5432`
- Maintenance database: `postgres`
- Username: `postgres`
- Password: `postgres` (or your actual password)

**Save Password:** ☑ Check this box

---

## Troubleshooting

### PostgreSQL Not Running
```bash
# Start PostgreSQL
brew services start postgresql@14

# Check status
brew services list | grep postgres
```

### Wrong PostgreSQL Version
If you have multiple PostgreSQL versions:
```bash
# List all PostgreSQL versions
brew list | grep postgres

# Use specific version
psql -U postgres -h localhost -p 5432
```

### Permission Issues
```bash
# Check PostgreSQL data directory permissions
ls -la $(brew --prefix)/var/postgresql@14

# Fix permissions if needed
chmod 700 $(brew --prefix)/var/postgresql@14
```

---

## After Fixing Password

1. **Update `.env` file** with correct password
2. **Test connection** in pgAdmin
3. **Verify database exists**:
   ```bash
   psql -U postgres -h localhost -c "\l" | grep uniroute
   ```
4. **Create database if missing**:
   ```bash
   psql -U postgres -h localhost -c "CREATE DATABASE uniroute;"
   ```

---

## Summary

**Most Common Solution:**
1. Stop PostgreSQL: `brew services stop postgresql@14`
2. Reset password using single-user mode
3. Restart: `brew services start postgresql@14`
4. Update `.env` with correct password
5. Connect in pgAdmin

**Quick Alternative:**
- Use trust authentication for local development
- Or create a new user with known password

