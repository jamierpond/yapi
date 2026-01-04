import * as vscode from 'vscode';
import * as fs from 'fs';

interface ViteManifestEntry {
  file: string;
  css?: string[];
  isEntry?: boolean;
}

interface ViteManifest {
  [key: string]: ViteManifestEntry;
}

/**
 * Gets the HTML content for the webview panel.
 * Loads the bundled Vite app with proper CSP headers using the Vite manifest.
 */
export function getWebviewHtml(
  webview: vscode.Webview,
  extensionUri: vscode.Uri
): string {
  const mediaRoot = vscode.Uri.joinPath(extensionUri, 'media');
  const manifestUri = vscode.Uri.joinPath(mediaRoot, '.vite', 'manifest.json');

  // Read and parse the Vite manifest
  let manifest: ViteManifest;
  try {
    manifest = JSON.parse(fs.readFileSync(manifestUri.fsPath, 'utf8'));
  } catch (error) {
    console.error('[yapi webview] Failed to read manifest:', error);
    throw new Error('Failed to load webview manifest. Make sure the webview has been built.');
  }

  // Get the entry point from the manifest
  const entry = manifest['index.html'];
  if (!entry) {
    throw new Error('Vite manifest missing index.html entry');
  }

  // Convert script path to webview URI
  const scriptUri = webview.asWebviewUri(vscode.Uri.joinPath(mediaRoot, entry.file));

  // Convert CSS paths to webview URIs
  const styleUris = (entry.css ?? []).map((cssRel: string) =>
    webview.asWebviewUri(vscode.Uri.joinPath(mediaRoot, cssRel))
  );

  // Font URIs for webview
  const fontRegular = webview.asWebviewUri(
    vscode.Uri.joinPath(mediaRoot, 'fonts', 'JetBrains_Mono', 'JetBrainsMono-VariableFont_wght.ttf')
  );
  const fontItalic = webview.asWebviewUri(
    vscode.Uri.joinPath(mediaRoot, 'fonts', 'JetBrains_Mono', 'JetBrainsMono-Italic-VariableFont_wght.ttf')
  );

  // Build CSP
  const csp = [
    "default-src 'none';",
    `img-src ${webview.cspSource} https: data:;`,
    `style-src ${webview.cspSource} 'unsafe-inline';`,
    `script-src ${webview.cspSource};`,
    `font-src ${webview.cspSource};`,
  ].join(' ');

  // Inline font-face rules with proper webview URIs
  const fontStyles = `
    @font-face {
      font-family: "JetBrains Mono";
      src: url("${fontRegular}") format("truetype");
      font-weight: 100 800;
      font-style: normal;
      font-display: swap;
    }
    @font-face {
      font-family: "JetBrains Mono";
      src: url("${fontItalic}") format("truetype");
      font-weight: 100 800;
      font-style: italic;
      font-display: swap;
    }
  `;

  return `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta http-equiv="Content-Security-Policy" content="${csp}" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <style>${fontStyles}</style>
    ${styleUris.map((u: vscode.Uri) => `<link rel="stylesheet" href="${u}">`).join('\n    ')}
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="${scriptUri}"></script>
  </body>
</html>`;
}
