# UDP Tunnel Use Cases

A UDP tunnel forwards UDP packets to your local service. Use it for connectionless protocols like DNS, gaming, or custom UDP apps.

**Generic steps:**

1. Start your local UDP service on a port (e.g. 53 for DNS, 19132 for a game).
2. Run **`uniroute udp <port>`** to get a public host and port.
3. Send UDP traffic to that host and port from anywhere.

Below are examples for DNS, gaming, and generic UDP.

---

## DNS server

Expose a local DNS server (e.g. Pi-hole, dnsmasq, CoreDNS) so you can use it from another network.

### 1. Start your DNS server

Run your DNS server bound to a port (e.g. 53). Ensure it listens on UDP (and optionally TCP if your server supports it; you’d need a separate TCP tunnel for that).

### 2. Start the UDP tunnel

```bash
uniroute udp 53
```

Note the public URL (e.g. `abc123.tunnel.uniroute.co:20000`).

### 3. Use from another machine

Point clients at the tunnel’s host and port as the DNS server. The **port** is the one shown by the tunnel (e.g. 20000), not 53.

**Linux/macOS (temporary):** Edit `/etc/resolv.conf` or use:

```bash
# Resolve a host through the tunneled DNS
dig @abc123.tunnel.uniroute.co -p 20000 example.com
```

**Windows:** Set the DNS server in adapter settings to the tunnel host and use the tunnel port if your client allows custom DNS port (some only allow 53; use a local forwarder if needed).

**Note:** Many systems expect DNS on port 53. You may need a local stub that forwards to `abc123.tunnel.uniroute.co:20000`, or use the tunnel in apps that support custom DNS host and port.

---

## Game server (UDP)

Expose a game server that uses UDP (e.g. Minecraft Bedrock, some FPS games).

### 1. Start the game server

Run your game server and note the UDP port (e.g. Minecraft Bedrock 19132).

### 2. Start the UDP tunnel

```bash
uniroute udp 19132
```

Note the public host and port (e.g. `abc123.tunnel.uniroute.co:20000`).

### 3. Connect from the game client

In the game, use:

- **Server address:** `abc123.tunnel.uniroute.co` (or the host shown).
- **Port:** the port shown by the tunnel (e.g. `20000`).

Not all games allow custom ports; check your game’s server configuration. If the game only accepts the default port, you might need a local port forward on the client side to map 19132 → tunnel host:20000.

---

## Generic UDP (custom app, streaming)

For any other UDP service (custom protocol, streaming, real-time app):

### 1. Start your service

Run your UDP server or daemon on a port.

### 2. Start the UDP tunnel

```bash
uniroute udp <port>
```

### 3. Send traffic to the public host and port

Use the tunnel host and port in your client or script. For example, with a simple UDP sender:

```bash
echo "test" | nc -u abc123.tunnel.uniroute.co 20000
```

(Exact behavior depends on your protocol.)

---

## See also

- [Tunnel Protocols](/docs/tunnels/protocols) – HTTP, TCP, TLS, UDP  
- [Opening a Tunnel](/docs/tunnels/opening) – Create and manage tunnels  
