import * as vscode from 'vscode';
import * as cp from 'child_process';

let panel: vscode.WebviewPanel | undefined;

function getWebviewContent(body: string, stderr: string, isLoading: boolean): string {
	const stderrSection = stderr ? `
		<div class="stderr">
			<div class="stderr-header">Info</div>
			<pre>${escapeHtml(stderr)}</pre>
		</div>
	` : '';

	const bodyContent = isLoading ? '<div class="loading">Running yapi...</div>' : `<pre class="json">${escapeHtml(body)}</pre>`;

	return `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>yapi Response</title>
	<style>
		body {
			font-family: var(--vscode-editor-font-family);
			font-size: var(--vscode-editor-font-size);
			padding: 16px;
			color: var(--vscode-editor-foreground);
			background: var(--vscode-editor-background);
			margin: 0;
		}
		.json {
			white-space: pre-wrap;
			word-wrap: break-word;
			margin: 0;
			line-height: 1.5;
		}
		.loading {
			color: var(--vscode-descriptionForeground);
			font-style: italic;
		}
		.stderr {
			margin-top: 24px;
			padding-top: 16px;
			border-top: 1px solid var(--vscode-panel-border);
		}
		.stderr-header {
			font-weight: bold;
			color: var(--vscode-descriptionForeground);
			margin-bottom: 8px;
			font-size: 0.9em;
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}
		.stderr pre {
			color: var(--vscode-descriptionForeground);
			margin: 0;
			font-size: 0.9em;
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

function getWebviewContentRaw(bodyHtml: string, stderr: string): string {
	const stderrSection = stderr ? `
		<div class="stderr">
			<div class="stderr-header">Info</div>
			<pre>${escapeHtml(stderr)}</pre>
		</div>
	` : '';

	return `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>yapi Response</title>
	<style>
		body {
			font-family: var(--vscode-editor-font-family);
			font-size: var(--vscode-editor-font-size);
			padding: 16px;
			color: var(--vscode-editor-foreground);
			background: var(--vscode-editor-background);
			margin: 0;
		}
		.json {
			white-space: pre-wrap;
			word-wrap: break-word;
			margin: 0;
			line-height: 1.5;
		}
		.stderr {
			margin-top: 24px;
			padding-top: 16px;
			border-top: 1px solid var(--vscode-panel-border);
		}
		.stderr-header {
			font-weight: bold;
			color: var(--vscode-descriptionForeground);
			margin-bottom: 8px;
			font-size: 0.9em;
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}
		.stderr pre {
			color: var(--vscode-descriptionForeground);
			margin: 0;
			font-size: 0.9em;
		}
		.string { color: var(--vscode-debugTokenExpression-string, #ce9178); }
		.number { color: var(--vscode-debugTokenExpression-number, #b5cea8); }
		.boolean { color: var(--vscode-debugTokenExpression-boolean, #569cd6); }
		.null { color: var(--vscode-debugTokenExpression-value, #569cd6); }
		.key { color: var(--vscode-symbolIcon-propertyForeground, #9cdcfe); }
	</style>
</head>
<body>
	<pre class="json">${bodyHtml}</pre>
	${stderrSection}
</body>
</html>`;
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
						match = match.slice(0, -1); // remove colon, add back after
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
		'yapi Response',
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
	if (!filePath.endsWith('.yapi.yml') && !filePath.endsWith('.yapi.yaml')) {
		vscode.window.showErrorMessage('Not a yapi file');
		return;
	}

	await editor.document.save();

	const webview = getOrCreatePanel(context);
	webview.webview.html = getWebviewContent('', '', true);

	cp.exec(`yapi -c "${filePath}"`, (error, stdout, stderr) => {
		let body = stdout || '';
		if (error && !stdout && !stderr) {
			body = `Error: ${error.message}`;
		}

		// Syntax highlight if JSON
		const highlightedBody = syntaxHighlightJson(body.trim());

		if (panel) {
			panel.webview.html = getWebviewContentRaw(highlightedBody, stderr);
		}
	});
}

export function activate(context: vscode.ExtensionContext) {
	console.log('yapi extension is now active');

	const runCommand = vscode.commands.registerCommand('yapi.runCurrent', () => runYapi(context));
	context.subscriptions.push(runCommand);

	// Status bar button
	const statusBar = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
	statusBar.text = '$(play) yapi';
	statusBar.command = 'yapi.runCurrent';
	statusBar.tooltip = 'Run yapi on current file';
	context.subscriptions.push(statusBar);

	const updateStatusBar = () => {
		const editor = vscode.window.activeTextEditor;
		if (editor && (editor.document.fileName.endsWith('.yapi.yml') || editor.document.fileName.endsWith('.yapi.yaml'))) {
			statusBar.show();
		} else {
			statusBar.hide();
		}
	};

	vscode.window.onDidChangeActiveTextEditor(updateStatusBar, null, context.subscriptions);
	updateStatusBar();

	// Hot reload on save
	vscode.workspace.onDidSaveTextDocument((doc) => {
		if (doc.fileName.endsWith('.yapi.yml') || doc.fileName.endsWith('.yapi.yaml')) {
			if (panel) {
				runYapi(context);
			}
		}
	}, null, context.subscriptions);
}

export function deactivate() {}
