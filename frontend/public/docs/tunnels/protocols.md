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

## TLS

TLS tunnels provide encrypted TCP connections.

```bash
# TLS tunnel (shortcut - recommended)
uniroute tls 5432
```

**Use cases:**
- PostgreSQL (with SSL)
- HTTPS services
- Secure database connections
- Encrypted protocols

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
