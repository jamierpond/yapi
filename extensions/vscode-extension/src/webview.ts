import * as vscode from 'vscode';
import * as path from 'path';

/**
 * Gets the HTML content for the webview panel.
 * Loads the bundled Vite app with proper CSP headers.
 */
export function getWebviewHtml(
  webview: vscode.Webview,
  extensionUri: vscode.Uri
): string {
  // Path to the webview dist folder
  const webviewDistPath = vscode.Uri.joinPath(
    extensionUri,
    '../../apps/vscode-webview/dist'
  );

  // Get URIs for the built assets
  const scriptUri = webview.asWebviewUri(
    vscode.Uri.joinPath(webviewDistPath, 'assets', 'index.js')
  );
  const styleUri = webview.asWebviewUri(
    vscode.Uri.joinPath(webviewDistPath, 'assets', 'index.css')
  );

  // Debug log the paths
  console.log('[yapi webview] Extension URI:', extensionUri.fsPath);
  console.log('[yapi webview] Webview dist path:', webviewDistPath.fsPath);
  console.log('[yapi webview] Script URI:', scriptUri.toString());
  console.log('[yapi webview] Style URI:', styleUri.toString());

  return `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta
      http-equiv="Content-Security-Policy"
      content="
        default-src 'none';
        img-src ${webview.cspSource} https: data:;
        style-src ${webview.cspSource} 'unsafe-inline';
        script-src ${webview.cspSource};
        font-src ${webview.cspSource};
      "
    />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <link rel="stylesheet" href="${styleUri}" />
  </head>
  <body>
    <div id="root">Loading...</div>
    <script type="module" src="${scriptUri}"></script>
  </body>
</html>`;
}
