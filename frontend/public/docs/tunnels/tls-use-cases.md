# TLS Tunnel Use Cases

A TLS tunnel forwards encrypted TCP traffic so your local TLS service is reachable from the internet. The tunnel does not terminate TLS—it forwards bytes, so the client’s TLS session is end-to-end with your local server.

**Generic steps for any use case:**

1. Start your local service with **SSL/TLS enabled** on a port (e.g. 5432, 443, 3306).
2. Run **`uniroute tls <port>`** to get a public URL like `abc123.tunnel.uniroute.co:20000`.
3. Connect from anywhere using that host and port with **SSL enabled** in your client.

Below are concrete examples for PostgreSQL, MySQL, HTTPS, and any generic TLS service.

---

## PostgreSQL

Expose your local PostgreSQL (with SSL) so you can connect from cloud apps, BI tools, or other servers.

### 1. Start PostgreSQL with SSL

Ensure PostgreSQL is running and SSL is enabled (e.g. in `postgresql.conf`: `ssl = on`). Use your normal port (default 5432).

### 2. Start the TLS tunnel

```bash
uniroute tls 5432
```

You’ll see a public URL like `abc123.tunnel.uniroute.co:20000`. Use the **subdomain** and **port** in your connection string.

### 3. Connect from anywhere

```bash
psql "postgresql://USER:PASSWORD@abc123.tunnel.uniroute.co:20000/DATABASE?sslmode=require"
```

Replace `USER`, `PASSWORD`, `abc123`, `20000`, and `DATABASE` with your values.

**GUI clients (DBeaver, pgAdmin, TablePlus):** Host = `abc123.tunnel.uniroute.co`, Port = `20000`, enable SSL (e.g. “Require”). If your server uses a self-signed certificate, allow insecure SSL or add the CA in the client.

---

## MySQL

Expose your local MySQL (with SSL) for remote access from apps or DB clients.

### 1. Start MySQL with SSL

Ensure MySQL is running with SSL enabled and listening on a port (e.g. 3306). Configure SSL in your MySQL server (e.g. `require_secure_transport = ON` if you want to force SSL).

### 2. Start the TLS tunnel

```bash
uniroute tls 3306
```

Note the public URL (e.g. `abc123.tunnel.uniroute.co:20000`).

### 3. Connect from anywhere

**Command line (mysql client):**

```bash
mysql -h abc123.tunnel.uniroute.co -P 20000 -u USER -p --ssl-mode=REQUIRED DATABASE
```

**Connection string (e.g. for apps):**

```
mysql://USER:PASSWORD@abc123.tunnel.uniroute.co:20000/DATABASE?tls=true
```

**GUI clients (DBeaver, MySQL Workbench, etc.):** Host = `abc123.tunnel.uniroute.co`, Port = `20000`, enable SSL/Use TLS. For self-signed certs, enable “Allow public key retrieval” or add the CA as needed.

---

## HTTPS (local dev or internal service)

Expose a local HTTPS server (e.g. dev server with TLS, or an internal API) so you can hit it from the internet.

### 1. Start your HTTPS service

Run your app or server with TLS on a port (e.g. 8443). It can use a self-signed certificate.

### 2. Start the TLS tunnel

```bash
uniroute tls 8443
```

You’ll get something like `abc123.tunnel.uniroute.co:20000`.

### 3. Connect from anywhere

**Browser:** Open `https://abc123.tunnel.uniroute.co:20000`. You may need to accept a self-signed cert warning.

**curl:**

```bash
curl -k https://abc123.tunnel.uniroute.co:20000/
```

Use `-k` only for self-signed certs; in production use proper certificates.

**Other clients:** Use the same host and port with TLS/HTTPS. The tunnel forwards the TLS handshake to your local server.

---

## Generic TLS service (any TCP + TLS)

For any other TLS-over-TCP service (custom apps, other databases, etc.), the steps are the same.

### 1. Start your service with TLS

Run your server with TLS enabled on a port (e.g. 9999).

### 2. Start the TLS tunnel

```bash
uniroute tls 9999
```

Note the public host and port (e.g. `abc123.tunnel.uniroute.co:20000`).

### 3. Connect from a TLS client

Use any TLS-capable client. Example with OpenSSL:

```bash
openssl s_client -connect abc123.tunnel.uniroute.co:20000
```

For self-signed certs you may need `-showcerts` and to accept the fingerprint. Your application or tool should connect to the same host and port with TLS enabled.

---

## See also

- [Tunnel Protocols](/docs/tunnels/protocols) – HTTP, TCP, TLS, UDP  
- [Opening a Tunnel](/docs/tunnels/opening) – Create and manage tunnels  
