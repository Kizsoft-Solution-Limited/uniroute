# Dev Server + Tunnel

Start your dev server and get a public URL in one go. Works with Laravel, Vue, React, Django, Rails, and more.

## Three ways to get a public URL

### Option 1: One command (we start everything)

We start your framework’s dev server and the tunnel. Port is auto-detected.

```bash
cd your-project
uniroute dev
```

- **Laravel:** runs `php artisan serve` and tunnel to 8000
- **Node/Vite/Next:** runs `npm run dev` and tunnel to 5173/3002/3000
- **Django:** runs `python manage.py runserver` and tunnel to 8000
- **Flask:** runs `flask run` and tunnel to 5000
- **FastAPI:** runs `uvicorn main:app --reload` and tunnel to 8000
- **Go:** runs `go run .` and tunnel to 8080
- **Rails:** runs `rails s` and tunnel to 3000

### Option 2: Your command + attach (two terminals)

Keep using your normal command; we only add the tunnel.

**Terminal 1:** run your dev server as usual

```bash
php artisan serve
# or: npm run dev, rails s, etc.
```

**Terminal 2:** from the same project directory

```bash
uniroute dev --attach
```

We detect the port from the project and start the tunnel only.

### Option 3: Your command as-is; we run it and add the tunnel

Use your exact command; we run it and start the tunnel. Port can come from your command.

```bash
uniroute run -- php artisan serve
uniroute run -- npm run dev
uniroute run -- rails s
```

**Port from your command:** if you pass a port, we use it for the tunnel:

```bash
# Laravel on 8080 → tunnel to 8080
uniroute run -- php artisan serve --port=8080

# Or with a space
uniroute run -- php artisan serve --port 8080

# Rails on 3001
uniroute run -- rails s -p 3001
```

Supported forms: `--port=8080`, `--port 8080`, `-p 8080`.

**Port priority:** (1) port in your command, (2) our `--port` flag, (3) auto-detected from project.

## Supported projects (auto-detected)

| Language | Framework / runtime | Detection | Default port |
|----------|---------------------|-----------|--------------|
| JavaScript/TypeScript | Node (Vite, Next, React, etc.) | `package.json` + `dev` script | 3000, 5173, or 3002 |
| PHP | Laravel | `composer.json` + `artisan` | 8000 |
| Python | Django | `manage.py` | 8000 |
| Python | Flask | `requirements.txt` / `pyproject.toml` + flask | 5000 |
| Python | FastAPI | `requirements.txt` / `pyproject.toml` + fastapi | 8000 (uvicorn) |
| Go | stdlib / any | `go.mod` | 8080 |
| Ruby | Rails | `Gemfile` + `config.ru` | 3000 |

## Examples

```bash
# Option 1: one command
uniroute dev

# Option 2: attach (server already running in another terminal)
uniroute dev --attach
uniroute dev --attach --port 8000

# Option 3: your command + tunnel (port from command or project)
uniroute run -- php artisan serve --port=8080
uniroute run -- npm run dev
uniroute run --port 3000 -- npm run dev
```

## Flags

**uniroute dev**

- `--port`, `-p` – Tunnel port (overrides auto-detected)
- `--dir` – Project directory (default: current directory)
- `--no-tunnel` – Start dev server only, no tunnel
- `--attach` – Tunnel only; you run the dev server yourself

**uniroute run**

- `--port`, `-p` – Tunnel port (overrides port from your command or project)
- `--dir` – Project directory (default: current directory)

## Custom domains with dev/run

Tunnels started by `uniroute dev` or `uniroute run` are HTTP tunnels. If you assign a [custom domain](/docs/tunnels/custom-domains) to that tunnel (by subdomain or via the dashboard), traffic to your domain is routed to the same dev tunnel. No extra steps: run `uniroute dev` (or `run`), then use your custom domain in the browser or webhooks. The tunnel server resolves your domain to the active tunnel by subdomain.

## Next Steps

- [Opening a Tunnel](/docs/tunnels/opening) - Manual tunnel commands
- [Protocols](/docs/tunnels/protocols) - HTTP, TCP, TLS, UDP
- [Custom Domains](/docs/tunnels/custom-domains) - Use your own domain
