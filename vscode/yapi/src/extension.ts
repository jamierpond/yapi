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

function getWebviewContent(body: string, stderr: string, isLoading: boolean, executionTime?: number): string {
    const stderrSection = stderr ? `
        <div class="info-section">
            <div class="info-header">Info</div>
            <pre class="info-content">${escapeHtml(stderr)}</pre>
        </div>
    ` : '';

    const timeSection = executionTime !== undefined ? `
        <div class="timing">Request completed in ${executionTime}ms</div>
    ` : '';

    const bodyContent = isLoading
        ? '<div class="loading"><div class="spinner"></div><span>Running yapi...</span></div>'
        : `<pre class="json">${body}</pre>`;

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>yapi Response</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: var(--vscode-editor-font-family);
            font-size: var(--vscode-editor-font-size);
            padding: 24px;
            color: var(--vscode-editor-foreground);
            background: var(--vscode-editor-background);
            line-height: 1.6;
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
        }

        .loading {
            display: flex;
            align-items: center;
            gap: 12px;
            color: var(--vscode-descriptionForeground);
            font-size: 14px;
            padding: 32px 0;
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

        .info-section {
            margin-top: 24px;
            padding: 16px;
            background: var(--vscode-textBlockQuote-background);
            border-left: 3px solid var(--vscode-textBlockQuote-border);
            border-radius: 4px;
        }

        .info-header {
            font-weight: 600;
            color: var(--vscode-foreground);
            margin-bottom: 8px;
            font-size: 12px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            opacity: 0.8;
        }

        .info-content {
            color: var(--vscode-descriptionForeground);
            font-size: 12px;
            line-height: 1.5;
        }

        .timing {
            margin-top: 16px;
            padding: 8px 12px;
            background: var(--vscode-badge-background);
            color: var(--vscode-badge-foreground);
            border-radius: 4px;
            font-size: 11px;
            font-weight: 500;
            display: inline-block;
        }

        /* JSON syntax highlighting */
        .string { color: var(--vscode-debugTokenExpression-string, #ce9178); }
        .number { color: var(--vscode-debugTokenExpression-number, #b5cea8); }
        .boolean { color: var(--vscode-debugTokenExpression-boolean, #569cd6); }
        .null { color: var(--vscode-debugTokenExpression-value, #569cd6); }
        .key { color: var(--vscode-symbolIcon-propertyForeground, #9cdcfe); }
    </style>
</head>
<body>
    ${bodyContent}
    ${timeSection}
    ${stderrSection}
</body>
</html>`;
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
        { enableScripts: false }
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
    webview.webview.html = getWebviewContent('', '', true);

    const startTime = Date.now();

    // Use yapi run command with proper environment
    cp.exec(`"${yapiPath}" run "${filePath}"`, {
        shell: '/bin/bash',
        env: { ...process.env }
    }, (error, stdout, stderr) => {
        const executionTime = Date.now() - startTime;
        let body = stdout || '';

        console.log('[yapi] Command output:', { error, stdout, stderr });

        if (error && !stdout && !stderr) {
            body = `Error: ${error.message}`;
        }

        const highlightedBody = syntaxHighlightJson(body.trim());

        if (panel) {
            panel.webview.html = getWebviewContent(highlightedBody, stderr, false, executionTime);
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
