# PostgreSQL UI Tools Guide

This guide shows you how to connect to your PostgreSQL database using various UI tools to view and manage your UniRoute database.

## Prerequisites

1. PostgreSQL must be running
2. You need your database connection details:
   - Host: `localhost` (or your PostgreSQL host)
   - Port: `5432` (default PostgreSQL port)
   - Database: `uniroute` (or your database name)
   - Username: `postgres` (or your PostgreSQL user)
   - Password: Your PostgreSQL password

## Recommended PostgreSQL UI Tools

### 1. **pgAdmin** (Free, Open Source)
**Best for**: Full-featured PostgreSQL administration

**Installation:**
- **macOS**: `brew install --cask pgadmin4`
- **Linux**: `sudo apt install pgadmin4` or download from [pgadmin.org](https://www.pgadmin.org/download/)
- **Windows**: Download from [pgadmin.org](https://www.pgadmin.org/download/)

**Connection Steps:**
1. Open pgAdmin
2. Right-click "Servers" → "Create" → "Server"
3. General tab:
   - Name: `UniRoute Local`
4. Connection tab:
   - Host: `localhost`
   - Port: `5432`
   - Maintenance database: `postgres`
   - Username: `postgres`
   - Password: Your PostgreSQL password
   - Save password: ✓
5. Click "Save"

### 2. **DBeaver** (Free, Open Source)
**Best for**: Universal database tool (supports many databases)

**Installation:**
- **macOS**: `brew install --cask dbeaver-community`
- **Linux**: Download from [dbeaver.io](https://dbeaver.io/download/)
- **Windows**: Download from [dbeaver.io](https://dbeaver.io/download/)

**Connection Steps:**
1. Open DBeaver
2. Click "New Database Connection" (plug icon)
3. Select "PostgreSQL"
4. Enter connection details:
   - Host: `localhost`
   - Port: `5432`
   - Database: `uniroute`
   - Username: `postgres`
   - Password: Your PostgreSQL password
5. Click "Test Connection" → "Finish"

### 3. **TablePlus** (Free for limited connections, Paid for unlimited)
**Best for**: Beautiful, fast, native macOS/Windows app

**Installation:**
- **macOS**: `brew install --cask tableplus`
- **Windows/Linux**: Download from [tableplus.com](https://tableplus.com/)

**Connection Steps:**
1. Open TablePlus
2. Click "Create a new connection"
3. Select "PostgreSQL"
4. Enter:
   - Name: `UniRoute`
   - Host: `localhost`
   - Port: `5432`
   - User: `postgres`
   - Password: Your PostgreSQL password
   - Database: `uniroute`
5. Click "Test" → "Connect"

### 4. **Postico** (macOS only, Free trial, Paid)
**Best for**: Simple, beautiful macOS PostgreSQL client

**Installation:**
- **macOS**: `brew install --cask postico` or download from [eggerapps.at/postico](https://eggerapps.at/postico/)

**Connection Steps:**
1. Open Postico
2. Click "New Favorite"
3. Enter:
   - Host: `localhost`
   - Port: `5432`
   - User: `postgres`
   - Password: Your PostgreSQL password
   - Database: `uniroute`
4. Click "Connect"

### 5. **VS Code Extension: PostgreSQL** (Free)
**Best for**: If you already use VS Code

**Installation:**
1. Open VS Code
2. Go to Extensions (Cmd+Shift+X)
3. Search for "PostgreSQL" by Chris Kolkman
4. Install

**Connection Steps:**
1. Open Command Palette (Cmd+Shift+P)
2. Type "PostgreSQL: Add Connection"
3. Enter connection details:
   - Host: `localhost`
   - Port: `5432`
   - Database: `uniroute`
   - Username: `postgres`
   - Password: Your PostgreSQL password

## Quick Connection String Reference

If you're using a connection string format (like in `.env`):

```
postgres://postgres:your_password@localhost:5432/uniroute?sslmode=disable
```

Breakdown:
- `postgres://` - Protocol
- `postgres` - Username
- `your_password` - Password
- `localhost` - Host
- `5432` - Port
- `uniroute` - Database name
- `?sslmode=disable` - SSL mode (disable for local development)

## Common Database Operations

### View All Tables
```sql
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public';
```

### View Users Table
```sql
SELECT * FROM users;
```

### View API Keys Table
```sql
SELECT id, name, user_id, created_at, expires_at, is_active 
FROM api_keys;
```

### View Tunnels Table
```sql
SELECT * FROM tunnels;
```

### View Requests Table (Analytics)
```sql
SELECT * FROM requests ORDER BY created_at DESC LIMIT 100;
```

## Troubleshooting

### Connection Refused
- **Check if PostgreSQL is running:**
  ```bash
  # macOS
  brew services list | grep postgresql
  
  # Linux
  sudo systemctl status postgresql
  
  # Or try connecting via command line
  psql -U postgres -h localhost
  ```

### Authentication Failed
- Check your username and password
- Verify PostgreSQL authentication settings in `pg_hba.conf`
- For local development, you might need to set `trust` authentication

### Database Not Found
- Create the database:
  ```bash
  psql -U postgres -c "CREATE DATABASE uniroute;"
  ```
- Or run migrations:
  ```bash
  psql $DATABASE_URL -f migrations/001_initial_schema.sql
  ```

## Recommended Tool for Quick Access

**For macOS**: **TablePlus** or **Postico** (beautiful, fast, native)
**For Cross-platform**: **DBeaver** (free, feature-rich)
**For Full Administration**: **pgAdmin** (most features, but heavier)

## Next Steps

After connecting:
1. Browse the `users` table to see registered users
2. Check `api_keys` table for API keys
3. View `tunnels` table for active tunnels
4. Explore `requests` table for analytics data

