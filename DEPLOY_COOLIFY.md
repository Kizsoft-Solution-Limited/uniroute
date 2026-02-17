# Deploy Frontend, Backend, and Tunnel Separately in Coolify

Same repo, three Coolify applications. Use one **Server** and one **Project** (or three projects), then add three **Applications** from the same Git repository.

---

## 1. Frontend (uniroute.co, www.uniroute.co)

- **Source:** Same repo, same branch (e.g. `main`).
- **Build Pack:** Dockerfile.
- **Dockerfile Location:** `frontend/Dockerfile`.
- **Build Context / Base Directory:** leave **empty** or `.` (repo root). Do not set to `frontend` or Coolify may resolve `frontend/frontend`.
- **Port:** 80 (nginx).
- **Domain:** `uniroute.co`, `www.uniroute.co`. Enable ACME.
- **Build env (optional):** `VITE_API_BASE_URL=https://app.uniroute.co` (default in Dockerfile; override if different).

In Coolify: New Application → from Git → select repo → set **Dockerfile path** to `frontend/Dockerfile` and **Build context** (or “Root directory”) to `frontend`.

---

## 2. Backend / Gateway (app.uniroute.co)

- **Source:** Same repo, same branch.
- **Build Pack:** Dockerfile.
- **Dockerfile Location:** `Dockerfile` (root).
- **Build Context:** `.` (root).
- **Port:** 8084 (exposed in Dockerfile).
- **Domain:** `app.uniroute.co`. Enable ACME.
- **Env:** Set in Coolify (e.g. `BASE_URL`, `DATABASE_URL`, `JWT_SECRET`, `API_KEY_SECRET`, `TUNNEL_BASE_DOMAIN`, `WEBSITE_URL`, etc.). See `.env.example`.

In Coolify: New Application → from Git → same repo → **Dockerfile path** `Dockerfile`, **Build context** `.` (default).

---

## 3. Tunnel Server (tunnel.uniroute.co + *.uniroute.co)

- **Source:** Same repo, same branch.
- **Build Pack:** Dockerfile.
- **Dockerfile Location:** `Dockerfile.tunnel`.
- **Build Context:** `.` (root).
- **Port:** 8080 (exposed in Dockerfile.tunnel).
- **Domain:** `*.uniroute.co` (wildcard). Use **wildcard SSL** (DNS challenge with Cloudflare in proxy settings).
- **Env:** `TUNNEL_BASE_DOMAIN=uniroute.co`, `WEBSITE_URL=https://uniroute.co`, `DATABASE_URL`, `JWT_SECRET`, `API_KEY_SECRET` (same as gateway).

In Coolify: New Application → from Git → same repo → **Dockerfile path** `Dockerfile.tunnel`, **Build context** `.`.

---

## Summary

| App       | Dockerfile           | Context  | Port | Domain(s)        |
|----------|----------------------|----------|------|-------------------|
| Frontend | `frontend/Dockerfile`| `frontend` | 80   | uniroute.co, www  |
| Backend  | `Dockerfile`         | `.`      | 8084 | app.uniroute.co   |
| Tunnel   | `Dockerfile.tunnel`  | `.`      | 8080 | *.uniroute.co     |

All three point at the **same repository and branch**; only the Dockerfile path and build context differ. After deployment, set env vars for backend and tunnel (and optional build arg for frontend) as in the main docs.
