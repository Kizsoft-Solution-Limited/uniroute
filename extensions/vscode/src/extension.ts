import * as vscode from 'vscode';
import * as path from 'path';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

let chatProvider: ChatViewProvider | null = null;

export function activate(context: vscode.ExtensionContext) {
  const apiUrl = vscode.workspace.getConfiguration('uniroute').get<string>('apiUrl') ?? '';
  const cliPath = vscode.workspace.getConfiguration('uniroute').get<string>('cliPath') ?? 'uniroute';

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
  constructor(
    private readonly extensionUri: vscode.Uri,
    private readonly apiUrl: string,
  ) {}

  resolveWebviewView(
    webviewView: vscode.WebviewView,
    _context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken,
  ): void | Thenable<void> {
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
  <p>Set <strong>UniRoute: Frontend URL</strong> in Settings to your app URL (e.g. <code>https://uniroute.co</code>).</p>
</body>
</html>`;
  }
  const chatAppUrl = allowed + '/chat';
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
    <a href="#">UniRoute Chat</a> – runs in your codebase. Use the app for full chat; here you can open it in browser.
  </div>
  <iframe src="${chatAppUrl}" title="UniRoute Chat"></iframe>
  <script>
    (function() {
      const vscode = acquireVsCodeApi();
      window.addEventListener('message', function(e) {
        if (e.data && e.data.type === 'applyEdit' && e.data.edit) {
          vscode.postMessage({ type: 'applyEdit', edit: e.data.edit });
        } else if (e.data && e.data.type === 'rejectEdit') {
          vscode.postMessage({ type: 'rejectEdit' });
        }
      });
    })();
  </script>
</body>
</html>`;
}
