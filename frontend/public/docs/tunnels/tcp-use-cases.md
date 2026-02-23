# TCP Tunnel Use Cases

A TCP tunnel forwards raw TCP traffic to your local service. Use it when the service does **not** use TLS (no SSL). If your database or service supports TLS, prefer a [TLS tunnel](/docs/tunnels/tls-use-cases) for encryption.

**Generic steps:**

1. Start your local service on a port (e.g. 3306, 6379).
2. Run **`uniroute tcp <port>`** to get a public host and port (e.g. `abc123.tunnel.uniroute.co:20000`).
3. Connect from anywhere using that host and port.

Below are examples for MySQL (no SSL), Redis, and generic TCP services.

---

## MySQL (without SSL)

Expose local MySQL when you don’t need encryption over the tunnel (e.g. dev or trusted network). For production over the internet, use a [TLS tunnel](/docs/tunnels/tls-use-cases) with MySQL SSL instead.

### 1. Start MySQL

Ensure MySQL is running and listening (e.g. on 3306).

### 2. Start the TCP tunnel

```bash
uniroute tcp 3306
```

Note the public URL (e.g. `abc123.tunnel.uniroute.co:20000`).

### 3. Connect from anywhere

**Command line:**

```bash
mysql -h abc123.tunnel.uniroute.co -P 20000 -u USER -p DATABASE
```

**Connection string:**

```
mysql://USER:PASSWORD@abc123.tunnel.uniroute.co:20000/DATABASE
```

**GUI clients:** Host = `abc123.tunnel.uniroute.co`, Port = `20000`. No SSL required.

---

## Redis

Expose your local Redis instance for remote clients or cloud apps.

### 1. Start Redis

Ensure Redis is running (default port 6379).

### 2. Start the TCP tunnel

```bash
uniroute tcp 6379
```

Note the public host and port (e.g. `abc123.tunnel.uniroute.co:20000`).

### 3. Connect from anywhere

**redis-cli:**

```bash
redis-cli -h abc123.tunnel.uniroute.co -p 20000
```

**Connection string (apps):**

```
redis://abc123.tunnel.uniroute.co:20000
```

If Redis has a password: `redis://:PASSWORD@abc123.tunnel.uniroute.co:20000`. Use the same host and port in Redis GUI clients or SDKs.

---

## Generic TCP (custom app, game server)

For any other TCP service (custom server, game, tool), the steps are the same.

### 1. Start your service

Run your server or daemon on a port (e.g. 25565 for Minecraft, 27017 for MongoDB without TLS, or any custom port).

### 2. Start the TCP tunnel

```bash
uniroute tcp <port>
```

Example for Minecraft:

```bash
uniroute tcp 25565
```

### 3. Connect from anywhere

Use the public host and port in your client or application. For example:

- **Minecraft:** Server address = `abc123.tunnel.uniroute.co`, Port = `20000`.
- **MongoDB (no TLS):** `mongodb://abc123.tunnel.uniroute.co:20000`.
- **Custom client:** Connect to `abc123.tunnel.uniroute.co` on the shown port.

**Testing with netcat (if your service speaks plain text):**

```bash
nc abc123.tunnel.uniroute.co 20000
```

---

## When to use TCP vs TLS

| Use TCP when …              | Use TLS when …                          |
|----------------------------|-----------------------------------------|
| Service has no SSL (MySQL, Redis, etc.) | Service uses SSL (Postgres with SSL, HTTPS) |
| Dev / internal only        | Exposing over the internet securely     |

See [TLS use cases](/docs/tunnels/tls-use-cases) for PostgreSQL, MySQL with SSL, HTTPS, and more.

---

## See also

- [Tunnel Protocols](/docs/tunnels/protocols) – HTTP, TCP, TLS, UDP  
- [Opening a Tunnel](/docs/tunnels/opening) – Create and manage tunnels  
