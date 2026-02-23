# Tunnel Protocols

UniRoute supports multiple protocols for different use cases.

## HTTP

HTTP tunnels are the most common, used for web applications and APIs.

```bash
# HTTP tunnel (shortcut - recommended)
uniroute http 8080
```

**Use cases:**
- Web applications
- REST APIs
- Development servers
- Local testing

## TCP

TCP tunnels provide raw TCP connections, useful for databases and other TCP services.

```bash
# TCP tunnel (shortcut - recommended)
uniroute tcp 3306
```

**Use cases:**
- MySQL databases
- Redis
- Custom TCP services
- Game servers

Step-by-step guides: [TCP use cases](/docs/tunnels/tcp-use-cases).

## TLS

TLS tunnels forward encrypted TCP traffic so your local TLS service (e.g. PostgreSQL) is reachable from the internet. The tunnel does not terminate TLS—it forwards bytes, so the client’s TLS session is end-to-end with your local server.

```bash
# TLS tunnel (shortcut - recommended)
uniroute tls 5432
```

You’ll get a public URL like `abc123.tunnel.uniroute.co:20000` that forwards to `localhost:5432`.

**How to use it:**

1. **Start your local service** (e.g. PostgreSQL) with SSL enabled on the port you use (e.g. 5432).
2. **Start the tunnel** with that port: `uniroute tls 5432`.
3. **Connect from anywhere** using the public URL and port. For PostgreSQL:
   ```bash
   psql "postgresql://USER:PASSWORD@abc123.tunnel.uniroute.co:20000/DATABASE?sslmode=require"
   ```
   Replace `abc123` and `20000` with the subdomain and port shown when you started the tunnel. Use the same URL in GUI clients (DBeaver, pgAdmin, etc.) with SSL enabled.

**Use cases:**
- PostgreSQL (with SSL)
- MySQL (with SSL)
- HTTPS services
- Secure database connections
- Encrypted protocols

Step-by-step guides for each: [TLS use cases](/docs/tunnels/tls-use-cases).

## UDP

UDP tunnels support connectionless protocols.

```bash
# UDP tunnel (shortcut - recommended)
uniroute udp 53
```

**Use cases:**
- DNS servers
- Gaming servers
- Real-time applications
- Streaming protocols

Step-by-step guides: [UDP use cases](/docs/tunnels/udp-use-cases).

## Protocol Comparison

| Protocol | Connection Type | Encryption | Use Cases |
|----------|----------------|------------|-----------|
| HTTP | Connection-oriented | HTTPS available | Web apps, APIs |
| TCP | Connection-oriented | No | Databases, custom services |
| TLS | Connection-oriented | Yes | Secure databases, HTTPS |
| UDP | Connectionless | No | DNS, gaming, streaming |

## Next Steps

- [Opening a Tunnel](/docs/tunnels/opening) - Create your first tunnel
- [Custom Domains](/docs/tunnels/custom-domains) - Use your own domain
