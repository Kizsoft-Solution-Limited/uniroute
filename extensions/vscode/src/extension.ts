import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import * as os from 'os';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

let chatProvider: ChatViewProvider | null = null;

const RELEASES_URL = 'https://github.com/Kizsoft-Solution-Limited/uniroute/releases/latest'
const GITHUB_API_LATEST = 'https://api.github.com/repos/Kizsoft-Solution-Limited/uniroute/releases/latest'
const CHECK_INTERVAL_MS = 24 * 60 * 60 * 1000

function parseVersion(s: string): number[] {
  const v = (s || '').replace(/^v/, '').trim()
  const parts = v.split('.').map((n) => parseInt(n, 10) || 0)
  return [parts[0] ?? 0, parts[1] ?? 0, parts[2] ?? 0]
}

function isNewer(latest: string, current: string): boolean {
  const a = parseVersion(latest)
  const b = parseVersion(current)
  for (let i = 0; i < 3; i++) {
    if (a[i] > b[i]) return true
    if (a[i] < b[i]) return false
  }
  return false
}

async function installUpdateInIde(vsixUrl: string, context: vscode.ExtensionContext): Promise<boolean> {
  try {
    const res = await fetch(vsixUrl, { redirect: 'follow' })
    if (!res.ok || !res.body) return false
    const buf = await res.arrayBuffer()
    const vsixPath = path.join(os.tmpdir(), `uniroute-${Date.now()}.vsix`)
    fs.writeFileSync(vsixPath, Buffer.from(buf))
    const execPath = process.execPath
    await execAsync(`"${execPath}" --install-extension "${vsixPath}" --force`)
    try { fs.unlinkSync(vsixPath) } catch { /* ignore */ }
    const reload = await vscode.window.showInformationMessage(
      'UniRoute: Extension updated. Reload the window to use the new version.',
      'Reload',
      'Later'
    )
    if (reload === 'Reload') {
      await vscode.commands.executeCommand('workbench.action.reloadWindow')
    }
    return true
  } catch (e) {
    console.error('UniRoute update install failed:', e)
    return false
  }
}

async function checkForUpdate(context: vscode.ExtensionContext) {
  const current = (context.extension.packageJSON?.version as string) || '0.0.0'
  const lastCheck = context.globalState.get<number>('uniroute.lastUpdateCheck') || 0
  if (Date.now() - lastCheck < CHECK_INTERVAL_MS) return
  context.globalState.update('uniroute.lastUpdateCheck', Date.now())

  try {
    const res = await fetch(GITHUB_API_LATEST, {
      headers: { Accept: 'application/vnd.github.v3+json' },
    })
    if (!res.ok) return
    const data = (await res.json()) as { tag_name?: string; assets?: { name: string; browser_download_url: string }[] }
    const latest = (data.tag_name || '').replace(/^v/, '').trim()
    if (!latest || !isNewer(latest, current)) return
    const lastNotified = context.globalState.get<string>('uniroute.lastNotifiedVersion')
    if (lastNotified === latest) return
    context.globalState.update('uniroute.lastNotifiedVersion', latest)

    const vsixAsset = data.assets?.find((a) => a.name.endsWith('.vsix'))
    const vsixUrl = vsixAsset?.browser_download_url

    const action = await vscode.window.showInformationMessage(
      `UniRoute: A new version (${latest}) is available. You have ${current}.`,
      'Update',
      'Download from GitHub',
      'Dismiss'
    )
    if (action === 'Update' && vsixUrl) {
      await vscode.window.withProgress(
        { location: vscode.ProgressLocation.Notification, title: 'UniRoute: Installing update…' },
        async () => {
          const ok = await installUpdateInIde(vsixUrl, context)
          if (!ok) {
            vscode.window.showErrorMessage(
              'UniRoute: Update failed. Try "Download from GitHub" and install the .vsix manually.'
            )
          }
        }
      )
    } else if (action === 'Download from GitHub' || (action === 'Update' && !vsixUrl)) {
      await vscode.env.openExternal(vscode.Uri.parse(RELEASES_URL))
    }
  } catch {}
}

export function activate(context: vscode.ExtensionContext) {
  const apiUrl = vscode.workspace.getConfiguration('uniroute').get<string>('apiUrl') ?? '';
  const cliPath = vscode.workspace.getConfiguration('uniroute').get<string>('cliPath') ?? 'uniroute';

  setTimeout(() => checkForUpdate(context), 5000);

  chatProvider = new ChatViewProvider(context.extensionUri, apiUrl);
  context.subscriptions.push(
    vscode.window.registerWebviewViewProvider('uniroute.chatView', chatProvider)
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('uniroute.chat.open', () => {
      vscode.commands.executeCommand('workbench.view.extension.uniroute');
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('uniroute.chat.sendContext', () => {
      const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath ?? '';
      const editor = vscode.window.activeTextEditor;
      let filePath = '';
      let selection = '';
      if (editor) {
        const doc = editor.document;
        filePath = workspaceRoot ? path.relative(workspaceRoot, doc.uri.fsPath) : doc.uri.fsPath;
        const range = editor.selection;
        if (!range.isEmpty) {
          selection = doc.getText(range);
        }
      }
      chatProvider?.postIdeContext({ workspaceRoot, filePath, selection });
      vscode.window.showInformationMessage(
        selection ? 'Current file and selection sent to UniRoute Chat.' : 'Current file path sent to UniRoute Chat.'
      );
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('uniroute.tunnel.start', async () => {
      const portStr = await vscode.window.showInputBox({
        prompt: 'Local port to expose (e.g. 3000)',
        placeHolder: '3000',
      });
      if (portStr == null) return;
      const port = parseInt(portStr.trim(), 10);
      if (Number.isNaN(port) || port < 1 || port > 65535) {
        vscode.window.showErrorMessage('Enter a valid port (1–65535).');
        return;
      }
      try {
        const { stdout } = await execAsync(`${cliPath} http ${port}`, {
          cwd: vscode.workspace.workspaceFolders?.[0]?.uri.fsPath,
        });
        vscode.window.showInformationMessage('Tunnel started. ' + (stdout || '').slice(0, 200));
      } catch (e: unknown) {
        const err = e as { message?: string };
        vscode.window.showErrorMessage('Tunnel failed: ' + (err?.message ?? String(e)));
      }
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('uniroute.tunnel.stop', async () => {
      vscode.window.showInformationMessage('Stop tunnel from the terminal or run: uniroute tunnel list then close the process.');
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('uniroute.tunnel.list', async () => {
      try {
        const { stdout } = await execAsync(`${cliPath} tunnel list`);
        const doc = await vscode.workspace.openTextDocument({
          content: stdout || 'No tunnels.',
          language: 'plaintext',
        });
        await vscode.window.showTextDocument(doc);
      } catch (e: unknown) {
        const err = e as { message?: string };
        vscode.window.showErrorMessage('List failed: ' + (err?.message ?? String(e)));
      }
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('uniroute.acceptEdit', (edit: { file: string; range: [number, number, number, number]; newText: string }) => {
      if (!edit?.file || !edit.range || edit.newText === undefined) return;
      const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
      if (!workspaceRoot) return;
      const [sL, sC, eL, eC] = edit.range;
      const maxLine = 10_000_000;
      if ([sL, sC, eL, eC].some((n) => typeof n !== 'number' || !Number.isFinite(n) || n < 0 || n > maxLine)) return;
      const resolvedPath = path.isAbsolute(edit.file)
        ? path.normalize(edit.file)
        : path.normalize(path.join(workspaceRoot, edit.file));
      if (!resolvedPath.startsWith(workspaceRoot) || resolvedPath.includes('\0')) return;
      try {
        const uri = vscode.Uri.file(resolvedPath);
        const range = new vscode.Range(sL, sC, eL, eC);
        const we = new vscode.WorkspaceEdit();
        we.replace(uri, range, edit.newText);
        vscode.workspace.applyEdit(we);
      } catch (_) {}
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('uniroute.rejectEdit', () => {
      vscode.window.showInformationMessage('Edit rejected.');
    })
  );
}

export function deactivate() {
  chatProvider = null;
}

class ChatViewProvider implements vscode.WebviewViewProvider {
  private webview: vscode.Webview | null = null;

  constructor(
    private readonly extensionUri: vscode.Uri,
    private readonly apiUrl: string,
  ) {}

  postIdeContext(ctx: { workspaceRoot?: string; filePath?: string; selection?: string }): void {
    this.webview?.postMessage({ type: 'uniroute.ideContext', ...ctx });
  }

  resolveWebviewView(
    webviewView: vscode.WebviewView,
    _context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken,
  ): void | Thenable<void> {
    this.webview = webviewView.webview;
    webviewView.webview.options = {
      enableScripts: true,
      localResourceRoots: [this.extensionUri],
    };
    const chatHtml = getChatWebviewHtml(webviewView.webview, this.extensionUri, this.apiUrl);
    webviewView.webview.html = chatHtml;

    webviewView.webview.onDidReceiveMessage((msg: { type: string; edit?: unknown }) => {
      if (msg.type === 'rejectEdit') {
        vscode.commands.executeCommand('uniroute.rejectEdit');
        return;
      }
      if (msg.type !== 'applyEdit' || msg.edit == null || typeof msg.edit !== 'object') return;
      const e = msg.edit as Record<string, unknown>;
      if (typeof e.file !== 'string' || typeof e.newText !== 'string' || !Array.isArray(e.range) || e.range.length !== 4) return;
      vscode.commands.executeCommand('uniroute.acceptEdit', {
        file: e.file,
        range: e.range as [number, number, number, number],
        newText: e.newText,
      });
    });
  }
}

function getChatWebviewHtml(webview: vscode.Webview, _extensionUri: vscode.Uri, apiUrl: string): string {
  const base = (apiUrl ?? '').replace(/\/$/, '').trim();
  const allowed = base.startsWith('https://') || base.startsWith('http://') ? base : '';
  if (allowed === '') {
    return `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>UniRoute Chat</title></head>
<body style="font-family: var(--vscode-font-family); padding: 1rem;">
  <p>Set <strong>UniRoute: Frontend URL</strong> in Settings to your app URL (only needed for self-hosted). Default: <code>https://uniroute.co</code>.</p>
</body>
</html>`;
  }
  const chatAppUrl = allowed + '/dashboard/chat';
  return `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta http-equiv="Content-Security-Policy" content="default-src 'none'; frame-src ${allowed} https:; script-src 'unsafe-inline'; style-src 'unsafe-inline';">
  <title>UniRoute Chat</title>
  <style>
    body, html { margin: 0; padding: 0; height: 100%; font-family: var(--vscode-font-family); }
    iframe { width: 100%; height: 100%; border: none; }
    .toolbar { padding: 8px; background: var(--vscode-sideBar-background); border-bottom: 1px solid var(--vscode-panel-border); }
    .toolbar a { color: var(--vscode-textLink-foreground); margin-right: 12px; }
  </style>
</head>
<body>
  <div class="toolbar">
    <span>UniRoute Chat</span> – Not signed in? The panel below shows the login page. After signing in, chat runs in your codebase.
    <a href="${chatAppUrl}" target="_blank" rel="noopener">Open in browser</a>
  </div>
  <iframe id="uniroute-chat-iframe" src="${chatAppUrl}" title="UniRoute Chat"></iframe>
  <script>
    (function() {
      const vscode = acquireVsCodeApi();
      const iframeOrigin = new URL('${allowed}').origin;
      window.addEventListener('message', function(e) {
        if (e.data && e.data.type === 'applyEdit' && e.data.edit) {
          vscode.postMessage({ type: 'applyEdit', edit: e.data.edit });
        } else if (e.data && e.data.type === 'rejectEdit') {
          vscode.postMessage({ type: 'rejectEdit' });
        }
      });
      window.addEventListener('message', function(e) {
        if (e.data && e.data.type === 'uniroute.ideContext') {
          var iframe = document.getElementById('uniroute-chat-iframe');
          if (iframe && iframe.contentWindow) {
            iframe.contentWindow.postMessage(e.data, iframeOrigin);
          }
        }
      });
    })();
  </script>
</body>
</html>`;
}
