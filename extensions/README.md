# UniRoute IDE Extensions

IDE extensions for UniRoute: in-editor chat with accept/reject code edits and tunnel management.

## Supported IDEs

| IDE | Extension path | Chat + Accept/Reject | Tunnels |
|-----|----------------|----------------------|---------|
| VS Code | `extensions/vscode` | Yes | Yes |
| IntelliJ IDEA | `extensions/jetbrains` | Yes | Yes |
| Android Studio | `extensions/jetbrains` | Yes | Yes |

The JetBrains plugin supports both IntelliJ IDEA and Android Studio.

## Features

**Chat**
- Open UniRoute chat from the IDE.
- Context: open files, selection, project structure.
- Accept/Reject suggested code edits in the editor.

**Tunnels**
- Start, stop, and list tunnels from the IDE via the UniRoute CLI or API.

## How to use (after install)

### VS Code

1. **Configure once**  
   Open **Settings** (Ctrl+, / Cmd+,), search for **UniRoute**, and set **UniRoute: Frontend URL** to your app URL (e.g. `https://uniroute.co`). Optionally set **Tunnel Server URL** (e.g. `https://tunnel.uniroute.co`) if you use tunnels.

2. **Open chat**  
   Click the UniRoute icon in the Activity Bar (left), or run the command **UniRoute: Open Chat**. The Chat panel opens and loads your app’s chat page in the sidebar.

3. **Sign in**  
   If the chat page shows a login screen, sign in there (same account as in the browser). The extension does not store credentials; the loaded app handles auth.

4. **Use chat**  
   Type messages and send. When the AI suggests a code edit, the message shows **Accept**, **Reject**, and **Copy as JSON**. Click **Accept** to apply the edit in the current workspace; **Reject** to dismiss it.

5. **Tunnels (optional)**  
   If the UniRoute CLI is in your PATH: **UniRoute: Start Tunnel** (enter a local port), **UniRoute: List Tunnels**, **UniRoute: Stop Tunnel** (stop from terminal or list).

### JetBrains (IntelliJ / Android Studio)

1. **Open chat in the browser**  
   Use your normal UniRoute app URL in the browser and sign in. The plugin does not embed the chat UI yet.

2. **Apply suggested edits**  
   When the AI suggests a code edit in the app, use **Copy as JSON** (or **Accept** to copy the edit JSON). In the IDE: **Tools → UniRoute → Apply suggested edit from clipboard**. The edit is applied to the project file.

3. **Tunnels**  
   **Tools → UniRoute → Start Tunnel** / **List Tunnels** (uses the UniRoute CLI if in PATH).

## Project layout

```
extensions/
├── README.md
├── vscode/        # VS Code extension (TypeScript)
└── jetbrains/     # JetBrains plugin (Kotlin)
```

## Building and developing

### VS Code

```bash
cd extensions/vscode
npm install
npm run compile
```

Run: F5 in VS Code (Launch Extension Development Host).  
Package: `vsce package`.

### JetBrains

```bash
cd extensions/jetbrains
./gradlew build
```

Run: `./gradlew runIde`.  
Install: use `build/distributions/UniRoute-*.zip` in Settings → Plugins → Install Plugin from Disk.

## Requirements

- **Chat API**: UniRoute chat endpoint (optional workspace context).
- **Tunnels**: `uniroute` CLI in PATH or tunnel API for start/stop/list.
- **Auth**: User authenticated via CLI or API token; extensions use stored config or prompt for login.

## Where to set apiUrl (frontend URL)

The extension loads the chat UI from your **frontend** in an iframe. Configure that URL once:

- **VS Code**: Settings → search “UniRoute” → **UniRoute: Frontend URL** (`uniroute.apiUrl`). Set to your app URL, e.g. `https://uniroute.co`. The chat panel will load `apiUrl/chat`.
- **JetBrains**: No iframe in the plugin today; use the same frontend URL in the browser. Apply suggested edit from clipboard uses the JSON from chat.

Your frontend (e.g. `https://uniroute.co`) must be able to call your backend (e.g. `https://app.uniroute.co`) for chat and auth; configure that in the web app as you do today.

## Deployment (no extra services)

The same frontend, backend, and tunnel server you already run are enough for extensions and MCP. No URLs are hardcoded in the extension or MCP code; all are from config or env.

| What | Where | Config |
|------|--------|--------|
| **Frontend** | Serves the chat UI at `/chat`. | Extension: set `uniroute.apiUrl` to this URL (e.g. `https://uniroute.co`). |
| **Backend (gateway)** | Chat streams, MCP, auth. | Optional: `MCP_SERVERS` env (comma-separated MCP server URLs). |
| **Tunnel server** | Used by CLI for “Start tunnel”. | Extension: `uniroute.tunnelServerUrl` (e.g. `https://tunnel.uniroute.co`). |

**MCP and local servers**: MCP servers that run only on the user’s machine are not reachable from a cloud backend. Public MCP URLs work from the deployed backend when set in `MCP_SERVERS` or passed per request.

## Accept/Reject flow

1. User requests a code change in chat.
2. Backend streams a suggested edit (file path, range, new text).
3. Chat UI shows Accept, Reject, and Copy as JSON for that message.
4. **VS Code**: Accept sends the edit to the extension, which applies it in the workspace.
5. **JetBrains**: Copy as JSON (or Accept in the dashboard to copy the edit), then Tools → UniRoute → Apply suggested edit from clipboard. Future: direct apply when chat is embedded via JS/Java bridge.

## Model Context Protocol (MCP)

The backend runs an MCP client and exposes MCP over the API. Configure server URLs via the gateway (e.g. `MCP_SERVERS` env: comma-separated URLs). The extensions store MCP URLs in settings; the dashboard or app can send those URLs when listing/calling tools or when requesting chat with MCP context.

**Backend**
- Env: `MCP_SERVERS` (comma-separated MCP server URLs). No default; set only if you want a default list.
- Auth (JWT or API key): `GET /auth/mcp/servers`, `GET /auth/mcp/tools?server_url=...`, `POST /auth/mcp/call` (body: `server_url`, `name`, `arguments`).
- API key: same under `/v1/mcp/servers`, `/v1/mcp/tools`, `/v1/mcp/call`.
- Chat: in `POST /auth/chat/stream` or `POST /v1/chat/stream`, optional body field `mcp_tool_calls`: `[{ "server_url", "name", "arguments" }]`. Those tools are called before the LLM; results are injected as a system message.

**Extensions**
| IDE | Location | Settings |
|-----|----------|----------|
| VS Code | Settings → UniRoute | `uniroute.mcp.figmaDesktop`, `uniroute.mcp.figmaRemote`, `uniroute.mcp.servers` |
| IntelliJ / Android Studio | `uniroute.xml` (UniRouteSettings) | `mcpFigmaDesktop`, `mcpFigmaRemote`, `mcpServerUrls` |

See [Figma MCP Server](https://developers.figma.com/docs/figma-mcp-server) for desktop and remote URLs. Configure in extension settings or `MCP_SERVERS`.

---

## Edit format reference

Structured edit payload used by backend, frontend, and extensions:

```json
{
  "file": "src/foo.ts",
  "range": [1, 0, 3, 0],
  "oldText": "optional",
  "newText": "replacement text"
}
```

- `file`: Path relative to workspace root or absolute.
- `range`: `[startLine, startCol, endLine, endCol]` 0-based. Optional if `oldText` is used for matching.
- `oldText`: Optional.
- `newText`: Text to insert or replace the range.

Extensions use `{ file, range, newText }` for the accept-edit command.

### Implementation status

| Layer | Status |
|-------|--------|
| Backend | `SuggestedEdit` in stream chunk; post-process emits `suggested_edit`. |
| Frontend | Parses `suggested_edit`; Accept/Reject/Copy as JSON; postMessage(applyEdit) in VS Code iframe. |
| VS Code | Chat iframe posts applyEdit; extension applies via `uniroute.acceptEdit`. |
| JetBrains | Apply suggested edit from clipboard; optional bridge when chat is embedded. |
