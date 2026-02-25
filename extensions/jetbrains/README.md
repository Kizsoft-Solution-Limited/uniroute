# UniRoute JetBrains Plugin (IntelliJ + Android Studio)

One plugin for **IntelliJ IDEA** and **Android Studio**: UniRoute chat (with accept/reject for AI edits) and tunnels.

## Features

- **UniRoute Chat** – Tool window with link to open chat in browser; chat works in your codebase; accept/reject can be added via embedded browser (JCEF) + message passing.
- **Start Tunnel** – Runs `uniroute http <port>` (prompts for port).
- **List Tunnels** – Runs `uniroute tunnel list` and shows output.

## Build & run

```bash
./gradlew build
./gradlew runIde   # Run IntelliJ with plugin loaded
```

## Install in Android Studio / IntelliJ

1. Build: `./gradlew build`
2. Zip is under `build/distributions/UniRoute-0.1.0.zip`
3. In IDE: **Settings → Plugins → ⚙ → Install Plugin from Disk…** → select the zip.

## Send context so the AI works in your codebase

- **Tools → UniRoute → Send current file/selection to Chat**: Opens the chat in your browser with the current file path and selection in the URL. The chat page reads these and includes them in your next message so the AI can see your code.
- In the **UniRoute Chat** tool window, use **Send current file/selection to Chat** to do the same (uses the currently active editor file and selection if any).

## Config

Stored in `UniRouteSettings.State` (and in `uniroute.xml`):

- **API URL**, **tunnel server URL**, **CLI path**
- **MCP**: `mcpFigmaDesktop`, `mcpFigmaRemote`, `mcpServerUrls`. Configure as needed for Figma or other MCP servers.

Add a **Settings → Tools → UniRoute** configurable later to edit these from the UI in IntelliJ and Android Studio.

## Accept / Reject (future)

To support accept/reject of AI-written code inside the IDE:

1. Embed chat in a browser (e.g. JCEF: `JBCefBrowser`) or build a simple Swing chat UI that calls your chat API.
2. When the backend returns a suggested edit (file path, range, new text), show a diff or confirmation dialog.
3. On Accept: run `WriteCommandAction.runWriteCommandAction` and apply the edit to the project file.
4. On Reject: discard the suggestion.
