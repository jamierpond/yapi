import * as vscode from 'vscode';
import { execSync } from 'child_process';
import * as path from 'path';
import * as fs from 'fs';
import {
    LanguageClient,
    LanguageClientOptions,
    Executable
} from 'vscode-languageclient/node';
import { getWebviewHtml } from './webview';
import { runYapiForUI } from '@yapi/client';

let client: LanguageClient | undefined;
let panel: vscode.WebviewPanel | undefined;
let outputChannel: vscode.OutputChannel;
let webviewReady = false;
let pendingMessage: { type: string; loading?: boolean; result?: unknown } | null = null;

function isYapiFile(fileName: string): boolean {
    return fileName.endsWith('.yapi') ||
           fileName.endsWith('.yapi.yml') ||
           fileName.endsWith('.yapi.yaml') ||
           fileName.endsWith('yapi.config.yml') ||
           fileName.endsWith('yapi.config.yaml');
}

function isExecutable(filePath: string): boolean {
    try {
        fs.accessSync(filePath, fs.constants.X_OK);
        return true;
    } catch {
        return false;
    }
}

function sendToWebview(message: { type: string; loading?: boolean; result?: unknown }) {
    if (!panel) return;

    if (webviewReady) {
        panel.webview.postMessage(message);
    } else {
        // Queue message to send when webview is ready
        pendingMessage = message;
    }
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

function getOrCreatePanel(context: vscode.ExtensionContext): vscode.WebviewPanel {
    if (panel) {
        panel.reveal(vscode.ViewColumn.Beside, true);
        return panel;
    }

    panel = vscode.window.createWebviewPanel(
        'yapiResponse',
        'YAPI',
        { viewColumn: vscode.ViewColumn.Beside, preserveFocus: true },
        {
            enableScripts: true,
            retainContextWhenHidden: true,
            localResourceRoots: [
                vscode.Uri.joinPath(context.extensionUri, 'media')
            ]
        }
    );

    // Load the webview HTML once
    panel.webview.html = getWebviewHtml(panel.webview, context.extensionUri);

    // Handle messages from webview
    panel.webview.onDidReceiveMessage(
        message => {
            switch (message.type) {
                case 'ready':
                    webviewReady = true;
                    // Send any pending message that was queued before ready
                    if (pendingMessage) {
                        panel?.webview.postMessage(pendingMessage);
                        pendingMessage = null;
                    }
                    break;
            }
        },
        undefined,
        context.subscriptions
    );

    panel.onDidDispose(() => {
        panel = undefined;
        webviewReady = false;
        pendingMessage = null;
    }, null, context.subscriptions);

    return panel;
}

async function runYapiCommand(context: vscode.ExtensionContext) {
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
        vscode.window.showErrorMessage('yapi executable not found. Please install yapi or configure the path in settings.');
        return;
    }

    await editor.document.save();

    getOrCreatePanel(context);

    // Send loading message (queued if webview not ready)
    sendToWebview({ type: 'setLoading', loading: true });

    // Execute and get UI-ready result
    const result = await runYapiForUI({
        executablePath: yapiPath,
        input: { type: 'file', path: filePath },
        timeout: 30000,
    });

    // Send result to webview
    sendToWebview({ type: 'setResult', result });
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
            if (process.platform !== 'win32' && !isExecutable(configuredPath)) {
                outputChannel.appendLine(`Configured path exists but is not executable: ${configuredPath}`);
            } else {
                outputChannel.appendLine(`Using configured yapi path: ${configuredPath}`);
                return configuredPath;
            }
        } else {
            outputChannel.appendLine(`Configured path not found: ${configuredPath}`);
        }
    }

    // Try common locations
    const homeDir = process.env.HOME || process.env.USERPROFILE;
    const commonPaths = [
        '/usr/local/bin/yapi',
        '/usr/bin/yapi',
        path.join(homeDir || '', 'go', 'bin', 'yapi'),
    ];

    for (const p of commonPaths) {
        if (fs.existsSync(p)) {
            if (process.platform !== 'win32' && !isExecutable(p)) {
                outputChannel.appendLine(`Found yapi at ${p} but it's not executable`);
                continue;
            }
            outputChannel.appendLine(`Found yapi at: ${p}`);
            return p;
        }
    }

    // Try which/where command
    try {
        const result = execSync(process.platform === 'win32' ? 'where yapi' : 'which yapi', {
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

async function startLanguageServer(): Promise<void> {
    const yapiPath = findYapiExecutable();
    if (!yapiPath) {
        const message = 'yapi executable not found. Please install yapi or configure the path in settings.';
        outputChannel.appendLine(`ERROR: ${message}`);
        vscode.window.showErrorMessage(message, 'Open Settings', 'Install yapi').then(selection => {
            if (selection === 'Open Settings') {
                vscode.commands.executeCommand('workbench.action.openSettings', 'yapi.executablePath');
            } else if (selection === 'Install yapi') {
                vscode.env.openExternal(vscode.Uri.parse('https://pond.audio/yapi/install'));
            }
        });
        return;
    }

    outputChannel.appendLine(`Starting yapi language server with: ${yapiPath}`);

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

    try {
        await client.start();
    } catch (err) {
        outputChannel.appendLine(`Failed to start language server: ${err}`);
        vscode.window.showErrorMessage(`Failed to start yapi language server: ${err instanceof Error ? err.message : String(err)}`);
    }
}

async function restartLanguageServer(): Promise<void> {
    if (client) {
        outputChannel.appendLine('Stopping language server for restart...');
        await client.stop();
        client = undefined;
    }
    await startLanguageServer();
}

export function activate(context: vscode.ExtensionContext) {
    console.log('yapi extension is now active');

    // Create output channel for debugging
    outputChannel = vscode.window.createOutputChannel('yapi');
    context.subscriptions.push(outputChannel);

    // Start the language server
    startLanguageServer();

    // Watch for configuration changes
    context.subscriptions.push(
        vscode.workspace.onDidChangeConfiguration(e => {
            if (e.affectsConfiguration('yapi.executablePath')) {
                outputChannel.appendLine('yapi.executablePath changed, restarting language server...');
                restartLanguageServer();
            }
        })
    );

    // Register commands
    const runCommand = vscode.commands.registerCommand('yapi.runCurrent', () => runYapiCommand(context));
    context.subscriptions.push(runCommand);

    const examplesCommand = vscode.commands.registerCommand('yapi.insertExample', showExamplePicker);
    context.subscriptions.push(examplesCommand);

    const restartLspCommand = vscode.commands.registerCommand('yapi.restartLanguageServer', () => {
        restartLanguageServer();
    });
    context.subscriptions.push(restartLspCommand);

    // Status bar
    const statusBar = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
    statusBar.text = '$(play) Run';
    statusBar.command = 'yapi.runCurrent';
    statusBar.tooltip = 'Run yapi (Cmd+Enter)';
    context.subscriptions.push(statusBar);

    const updateStatusBar = () => {
        const editor = vscode.window.activeTextEditor;
        if (editor && isYapiFile(editor.document.fileName)) {
            statusBar.show();
        } else {
            statusBar.hide();
        }
    };

    vscode.window.onDidChangeActiveTextEditor(() => {
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
