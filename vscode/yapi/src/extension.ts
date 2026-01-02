import * as vscode from 'vscode';
import * as cp from 'child_process';
import * as path from 'path';
import * as fs from 'fs';
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
    Executable
} from 'vscode-languageclient/node';

let client: LanguageClient;
let panel: vscode.WebviewPanel | undefined;
let outputChannel: vscode.OutputChannel;

function isYapiFile(fileName: string): boolean {
    return fileName.endsWith('.yapi') ||
           fileName.endsWith('.yapi.yml') ||
           fileName.endsWith('.yapi.yaml') ||
           fileName.endsWith('yapi.config.yml') ||
           fileName.endsWith('yapi.config.yaml');
}

const EXAMPLES = {
    http: {
        label: 'HTTP POST',
        yaml: `# yaml-language-server: $schema=https://pond.audio/yapi/schema
url: https://httpbin.org/post
method: POST
content_type: application/json

body:
  title: "Hello from yapi"
  content: "This is a test post"
  tags:
    - testing
    - api
  metadata:
    source: yapi
    version: "1.0"
`
    },
    grpc: {
        label: 'gRPC Example',
        yaml: `# yaml-language-server: $schema=https://pond.audio/yapi/schema
# gRPC Hello Service example
url: grpc://grpcb.in:9000

service: hello.HelloService
rpc: SayHello
plaintext: true

body:
  greeting: "World"
`
    },
    tcp: {
        label: 'TCP Echo',
        yaml: `# yaml-language-server: $schema=https://pond.audio/yapi/schema
# TCP echo server test
url: tcp://tcpbin.com:4242

method: tcp
data: "Hello from yapi!\\n"
encoding: text
read_timeout: 5
close_after_send: true
`
    }
};

interface YapiJsonOutput {
    success: boolean;
    body: string;
    transport?: string;
    statusCode?: number;
    headers?: Record<string, string>;
    requestUrl?: string;
    method?: string;
    service?: string;
    contentType?: string;
    sizeBytes?: number;
    sizeLines?: number;
    sizeChars?: number;
    timing?: number;
    warnings?: string[];
    error?: string;
}

function getWebviewContent(result: YapiJsonOutput | null, isLoading: boolean): string {
    if (isLoading) {
        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>yapi Response</title>
    <style>${getStyles()}</style>
</head>
<body>
    <div class="loading"><div class="spinner"></div><span>Running yapi...</span></div>
</body>
</html>`;
    }

    if (!result) {
        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>yapi Response</title>
    <style>${getStyles()}</style>
</head>
<body>
    <div class="empty-state">Ready to execute</div>
</body>
</html>`;
    }

    const hasHeaders = result.headers && Object.keys(result.headers).length > 0;
    const hasWarnings = result.warnings && result.warnings.length > 0;

    const tabsHtml = `
        <div class="tabs">
            <button class="tab active" onclick="switchTab('body')">Body</button>
            ${hasHeaders ? `<button class="tab" onclick="switchTab('headers')">Headers <span class="badge">${Object.keys(result.headers!).length}</span></button>` : ''}
            <button class="tab" onclick="switchTab('info')">Info</button>
            ${hasWarnings ? `<button class="tab" onclick="switchTab('warnings')">Warnings <span class="badge warning">${result.warnings!.length}</span></button>` : ''}
        </div>
    `;

    const bodyHtml = `<pre class="json">${syntaxHighlightJson(result.body)}</pre>`;

    const headersHtml = hasHeaders ? `
        <table class="headers-table">
            ${Object.entries(result.headers!).map(([key, value]) => `
                <tr>
                    <td class="header-key">${escapeHtml(key)}</td>
                    <td class="header-value">${escapeHtml(value)}</td>
                </tr>
            `).join('')}
        </table>
    ` : '';

    const infoHtml = `
        ${result.transport ? `<div class="info-row"><span class="info-label">Transport</span><span class="badge">${result.transport.toUpperCase()}</span></div>` : ''}
        ${result.requestUrl ? `<div class="info-row"><span class="info-label">URL</span><span class="info-value">${escapeHtml(result.requestUrl)}</span></div>` : ''}
        ${result.method ? `<div class="info-row"><span class="info-label">Method</span><span class="badge">${escapeHtml(result.method)}</span></div>` : ''}
        ${result.service ? `<div class="info-row"><span class="info-label">Service</span><span class="info-value">${escapeHtml(result.service)}</span></div>` : ''}
        ${result.statusCode !== undefined ? `<div class="info-row"><span class="info-label">Status</span><span class="badge status-${getStatusClass(result.statusCode)}">${result.statusCode}</span></div>` : ''}
        ${result.timing !== undefined ? `<div class="info-row"><span class="info-label">Time</span><span class="info-value">${result.timing}ms</span></div>` : ''}
        ${result.sizeBytes !== undefined ? `<div class="info-row"><span class="info-label">Size</span><span class="info-value">${result.sizeBytes} bytes${result.sizeLines !== undefined ? ` • ${result.sizeLines} lines` : ''}${result.sizeChars !== undefined ? ` • ${result.sizeChars} chars` : ''}</span></div>` : ''}
        ${result.contentType ? `<div class="info-row"><span class="info-label">Content-Type</span><span class="info-value">${escapeHtml(result.contentType)}</span></div>` : ''}
    `;

    const warningsHtml = hasWarnings ? result.warnings!.map(warning => `
        <div class="warning-item">⚠ ${escapeHtml(warning)}</div>
    `).join('') : '';

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>yapi Response</title>
    <style>${getStyles()}</style>
</head>
<body>
    ${tabsHtml}
    <div class="tab-content">
        <div id="tab-body" class="tab-panel active">${bodyHtml}</div>
        ${hasHeaders ? `<div id="tab-headers" class="tab-panel">${headersHtml}</div>` : ''}
        <div id="tab-info" class="tab-panel">${infoHtml}</div>
        ${hasWarnings ? `<div id="tab-warnings" class="tab-panel">${warningsHtml}</div>` : ''}
    </div>
    <script>
        function switchTab(tabName) {
            document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
            document.querySelectorAll('.tab-panel').forEach(panel => panel.classList.remove('active'));
            event.target.classList.add('active');
            document.getElementById('tab-' + tabName).classList.add('active');
        }
    </script>
</body>
</html>`;
}

function getStatusClass(statusCode: number): string {
    if (statusCode >= 200 && statusCode < 300) return 'success';
    if (statusCode >= 400) return 'error';
    return 'warning';
}

function getStyles(): string {
    return `
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: var(--vscode-editor-font-family);
            font-size: var(--vscode-editor-font-size);
            color: var(--vscode-editor-foreground);
            background: var(--vscode-editor-background);
            line-height: 1.6;
            overflow: hidden;
            display: flex;
            flex-direction: column;
            height: 100vh;
        }

        .tabs {
            display: flex;
            gap: 4px;
            padding: 8px 16px;
            background: var(--vscode-editorGroupHeader-tabsBackground);
            border-bottom: 1px solid var(--vscode-panel-border);
        }

        .tab {
            padding: 6px 12px;
            background: transparent;
            border: none;
            color: var(--vscode-tab-inactiveForeground);
            cursor: pointer;
            font-size: 12px;
            border-bottom: 2px solid transparent;
            transition: all 0.2s;
        }

        .tab:hover {
            background: var(--vscode-tab-hoverBackground);
        }

        .tab.active {
            color: var(--vscode-tab-activeForeground);
            border-bottom-color: var(--vscode-focusBorder);
        }

        .badge {
            display: inline-block;
            padding: 2px 6px;
            margin-left: 6px;
            background: var(--vscode-badge-background);
            color: var(--vscode-badge-foreground);
            border-radius: 10px;
            font-size: 10px;
            font-weight: 600;
        }

        .badge.warning {
            background: var(--vscode-editorWarning-foreground);
            color: var(--vscode-editor-background);
        }

        .badge.status-success {
            background: var(--vscode-testing-iconPassed);
            color: var(--vscode-editor-background);
        }

        .badge.status-error {
            background: var(--vscode-testing-iconFailed);
            color: var(--vscode-editor-background);
        }

        .badge.status-warning {
            background: var(--vscode-editorWarning-foreground);
            color: var(--vscode-editor-background);
        }

        .tab-content {
            flex: 1;
            overflow: auto;
        }

        .tab-panel {
            display: none;
            padding: 16px;
            height: 100%;
            overflow: auto;
        }

        .tab-panel.active {
            display: block;
        }

        .json {
            white-space: pre-wrap;
            word-wrap: break-word;
            font-family: var(--vscode-editor-font-family);
            font-size: 13px;
            line-height: 1.6;
            padding: 16px;
            background: var(--vscode-editor-background);
            border-radius: 6px;
            border: 1px solid var(--vscode-panel-border);
            margin: 0;
        }

        .loading {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 12px;
            color: var(--vscode-descriptionForeground);
            font-size: 14px;
            height: 100vh;
        }

        .empty-state {
            display: flex;
            align-items: center;
            justify-center;
            height: 100vh;
            color: var(--vscode-descriptionForeground);
            font-size: 14px;
        }

        .spinner {
            width: 16px;
            height: 16px;
            border: 2px solid var(--vscode-progressBar-background);
            border-top-color: transparent;
            border-radius: 50%;
            animation: spin 0.8s linear infinite;
        }

        @keyframes spin {
            to { transform: rotate(360deg); }
        }

        .info-row {
            display: flex;
            align-items: flex-start;
            gap: 16px;
            padding: 12px;
            background: var(--vscode-input-background);
            border: 1px solid var(--vscode-input-border);
            border-radius: 4px;
            margin-bottom: 8px;
        }

        .info-label {
            font-weight: 600;
            min-width: 120px;
            color: var(--vscode-descriptionForeground);
            font-size: 12px;
        }

        .info-value {
            flex: 1;
            font-family: var(--vscode-editor-font-family);
            font-size: 12px;
            word-break: break-all;
        }

        .headers-table {
            width: 100%;
            border-collapse: separate;
            border-spacing: 0 8px;
        }

        .headers-table tr {
            background: var(--vscode-input-background);
        }

        .headers-table td {
            padding: 12px;
            border: 1px solid var(--vscode-input-border);
            font-size: 12px;
        }

        .headers-table tr td:first-child {
            border-right: none;
            border-radius: 4px 0 0 4px;
        }

        .headers-table tr td:last-child {
            border-left: none;
            border-radius: 0 4px 4px 0;
        }

        .header-key {
            font-weight: 600;
            color: var(--vscode-descriptionForeground);
            width: 40%;
            font-family: var(--vscode-editor-font-family);
            vertical-align: top;
        }

        .header-value {
            color: var(--vscode-editor-foreground);
            font-family: var(--vscode-editor-font-family);
            word-break: break-all;
        }

        .warning-item {
            padding: 12px;
            background: var(--vscode-inputValidation-warningBackground);
            border: 1px solid var(--vscode-inputValidation-warningBorder);
            border-radius: 4px;
            margin-bottom: 8px;
            font-size: 13px;
        }

        /* JSON syntax highlighting */
        .string { color: var(--vscode-debugTokenExpression-string, #ce9178); }
        .number { color: var(--vscode-debugTokenExpression-number, #b5cea8); }
        .boolean { color: var(--vscode-debugTokenExpression-boolean, #569cd6); }
        .null { color: var(--vscode-debugTokenExpression-value, #569cd6); }
        .key { color: var(--vscode-symbolIcon-propertyForeground, #9cdcfe); }
    `;
}

function escapeHtml(text: string): string {
    return text
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');
}

function syntaxHighlightJson(json: string): string {
    try {
        const parsed = JSON.parse(json);
        const pretty = JSON.stringify(parsed, null, 2);
        return pretty.replace(
            /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g,
            (match) => {
                let cls = 'number';
                if (/^"/.test(match)) {
                    if (/:$/.test(match)) {
                        cls = 'key';
                        match = match.slice(0, -1);
                        return `<span class="${cls}">${escapeHtml(match)}</span>:`;
                    } else {
                        cls = 'string';
                    }
                } else if (/true|false/.test(match)) {
                    cls = 'boolean';
                } else if (/null/.test(match)) {
                    cls = 'null';
                }
                return `<span class="${cls}">${escapeHtml(match)}</span>`;
            }
        );
    } catch {
        return escapeHtml(json);
    }
}

function getOrCreatePanel(context: vscode.ExtensionContext): vscode.WebviewPanel {
    if (panel) {
        panel.reveal(vscode.ViewColumn.Beside, true);
        return panel;
    }

    panel = vscode.window.createWebviewPanel(
        'yapiResponse',
        'RESPONSE',
        { viewColumn: vscode.ViewColumn.Beside, preserveFocus: true },
        { enableScripts: true } // Enable scripts for tab switching
    );

    panel.onDidDispose(() => {
        panel = undefined;
    }, null, context.subscriptions);

    return panel;
}

async function runYapi(context: vscode.ExtensionContext) {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showErrorMessage('No active editor');
        return;
    }

    const filePath = editor.document.uri.fsPath;
    if (!isYapiFile(filePath)) {
        vscode.window.showErrorMessage('Not a yapi file');
        return;
    }

    const yapiPath = findYapiExecutable();
    if (!yapiPath) {
        vscode.window.showErrorMessage('yapi executable not found');
        return;
    }

    await editor.document.save();

    const webview = getOrCreatePanel(context);
    webview.webview.html = getWebviewContent(null, true);

    // Use yapi run command with --json flag for structured output
    cp.exec(`"${yapiPath}" run --json "${filePath}"`, {
        shell: '/bin/bash',
        env: { ...process.env }
    }, (error, stdout, stderr) => {
        console.log('[yapi] Command output:', { error, stdout, stderr });

        let result: YapiJsonOutput | null = null;

        try {
            // Try to parse JSON output from CLI
            result = JSON.parse(stdout);
        } catch (e) {
            // If JSON parsing fails, treat as error
            result = {
                success: false,
                body: stdout || 'No output',
                error: error ? error.message : 'Failed to parse JSON output',
            };
        }

        if (panel) {
            panel.webview.html = getWebviewContent(result, false);
        }
    });
}

async function insertExample(exampleKey: keyof typeof EXAMPLES) {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showErrorMessage('No active editor');
        return;
    }

    const example = EXAMPLES[exampleKey];
    const fullRange = new vscode.Range(
        editor.document.positionAt(0),
        editor.document.positionAt(editor.document.getText().length)
    );

    await editor.edit(editBuilder => {
        editBuilder.replace(fullRange, example.yaml);
    });
}

async function showExamplePicker() {
    const items = Object.entries(EXAMPLES).map(([key, value]) => ({
        label: value.label,
        description: key,
        key: key as keyof typeof EXAMPLES
    }));

    const selected = await vscode.window.showQuickPick(items, {
        placeHolder: 'Select an example to insert'
    });

    if (selected) {
        await insertExample(selected.key);
    }
}

function findYapiExecutable(): string | null {
    // First, try the configured path
    const config = vscode.workspace.getConfiguration('yapi');
    const configuredPath = config.get<string>('executablePath', 'yapi');

    if (configuredPath !== 'yapi') {
        // User specified a custom path
        if (fs.existsSync(configuredPath)) {
            outputChannel.appendLine(`Using configured yapi path: ${configuredPath}`);
            return configuredPath;
        } else {
            outputChannel.appendLine(`Configured path not found: ${configuredPath}`);
        }
    }

    // Try common locations
    const homeDir = process.env.HOME || process.env.USERPROFILE;
    const commonPaths = [
        path.join(homeDir || '', 'go', 'bin', 'yapi'),
        '/usr/local/bin/yapi',
        '/usr/bin/yapi',
    ];

    for (const p of commonPaths) {
        if (fs.existsSync(p)) {
            outputChannel.appendLine(`Found yapi at: ${p}`);
            return p;
        }
    }

    // Try which/where command
    try {
        const result = cp.execSync(process.platform === 'win32' ? 'where yapi' : 'which yapi', {
            encoding: 'utf8',
            env: { ...process.env }
        });
        const yapiPath = result.trim().split('\n')[0];
        if (yapiPath && fs.existsSync(yapiPath)) {
            outputChannel.appendLine(`Found yapi via which/where: ${yapiPath}`);
            return yapiPath;
        }
    } catch (error) {
        outputChannel.appendLine(`Failed to find yapi via which/where: ${error}`);
    }

    return null;
}

export function activate(context: vscode.ExtensionContext) {
    console.log('yapi extension is now active');

    // Create output channel for debugging
    outputChannel = vscode.window.createOutputChannel('yapi');
    context.subscriptions.push(outputChannel);

    // Find yapi executable
    const yapiPath = findYapiExecutable();
    if (!yapiPath) {
        const message = 'yapi executable not found. Please install yapi or configure the path in settings.';
        outputChannel.appendLine(`ERROR: ${message}`);
        vscode.window.showErrorMessage(message, 'Open Settings').then(selection => {
            if (selection === 'Open Settings') {
                vscode.commands.executeCommand('workbench.action.openSettings', 'yapi.executablePath');
            }
        });
    } else {
        outputChannel.appendLine(`Starting yapi language server with: ${yapiPath}`);

        // Set up LSP client
        const serverOptions: Executable = {
            command: yapiPath,
            args: ['lsp'],
            options: {
                env: { ...process.env }
            }
        };

        const clientOptions: LanguageClientOptions = {
            documentSelector: [
                { scheme: 'file', pattern: '**/*.yapi' },
                { scheme: 'file', pattern: '**/*.yapi.yml' },
                { scheme: 'file', pattern: '**/*.yapi.yaml' },
                { scheme: 'file', pattern: '**/yapi.config.yml' },
                { scheme: 'file', pattern: '**/yapi.config.yaml' }
            ],
            synchronize: {
                fileEvents: vscode.workspace.createFileSystemWatcher('**/*.yapi*')
            },
            outputChannel: outputChannel
        };

        client = new LanguageClient(
            'yapiLanguageServer',
            'yapi Language Server',
            serverOptions,
            clientOptions
        );

        // Start the LSP client
        client.start().catch(err => {
            outputChannel.appendLine(`Failed to start language server: ${err}`);
            vscode.window.showErrorMessage(`Failed to start yapi language server: ${err.message}`);
        });
    }

    // Register commands
    const runCommand = vscode.commands.registerCommand('yapi.runCurrent', () => runYapi(context));
    context.subscriptions.push(runCommand);

    const examplesCommand = vscode.commands.registerCommand('yapi.insertExample', showExamplePicker);
    context.subscriptions.push(examplesCommand);

    // Status bar
    const statusBar = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
    statusBar.text = '$(play) Run';
    statusBar.command = 'yapi.runCurrent';
    statusBar.tooltip = 'Run yapi (Cmd+Enter)';
    context.subscriptions.push(statusBar);

    const updateStatusBar = () => {
        const editor = vscode.window.activeTextEditor;
        console.log('[yapi] updateStatusBar called, editor:', editor?.document.fileName);
        if (editor && isYapiFile(editor.document.fileName)) {
            console.log('[yapi] Showing status bar for:', editor.document.fileName);
            statusBar.show();
        } else {
            console.log('[yapi] Hiding status bar, fileName:', editor?.document.fileName);
            statusBar.hide();
        }
    };

    vscode.window.onDidChangeActiveTextEditor(editor => {
        updateStatusBar();
    }, null, context.subscriptions);

    updateStatusBar();
}

export function deactivate(): Thenable<void> | undefined {
    if (!client) {
        return undefined;
    }
    return client.stop();
}
