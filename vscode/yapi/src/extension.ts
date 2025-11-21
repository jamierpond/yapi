import * as vscode from 'vscode';
import * as cp from 'child_process';

let outputDoc: vscode.TextDocument | undefined;
let outputEditor: vscode.TextEditor | undefined;

async function getOrCreateOutputPanel(): Promise<vscode.TextEditor> {
	// Check if output editor still exists and is visible
	if (outputEditor && vscode.window.visibleTextEditors.includes(outputEditor)) {
		return outputEditor;
	}

	// Create untitled document for output
	outputDoc = await vscode.workspace.openTextDocument({
		content: '',
		language: 'json'
	});

	// Open in side-by-side view
	outputEditor = await vscode.window.showTextDocument(outputDoc, {
		viewColumn: vscode.ViewColumn.Beside,
		preserveFocus: true,
		preview: false
	});

	return outputEditor;
}

async function setOutputContent(content: string) {
	const editor = await getOrCreateOutputPanel();
	const doc = editor.document;

	const edit = new vscode.WorkspaceEdit();
	const fullRange = new vscode.Range(
		doc.positionAt(0),
		doc.positionAt(doc.getText().length)
	);
	edit.replace(doc.uri, fullRange, content);
	await vscode.workspace.applyEdit(edit);
}

async function runYapi() {
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
	await setOutputContent('Running yapi...');

	cp.exec(`yapi -c "${filePath}"`, (error, stdout, stderr) => {
		let output = stdout || '';
		if (stderr) {
			output += '\n\n--- stderr ---\n' + stderr;
		}
		if (error && !stdout && !stderr) {
			output = `Error: ${error.message}`;
		}

		// Try to pretty-print JSON
		try {
			const parsed = JSON.parse(output.trim());
			output = JSON.stringify(parsed, null, 2);
		} catch {
			// Not JSON, keep as-is
		}

		setOutputContent(output);
	});
}

export function activate(context: vscode.ExtensionContext) {
	console.log('yapi extension is now active');

	// Register the run command
	const runCommand = vscode.commands.registerCommand('yapi.runCurrent', runYapi);
	context.subscriptions.push(runCommand);

	// Status bar button
	const statusBar = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
	statusBar.text = '$(play) yapi';
	statusBar.command = 'yapi.runCurrent';
	statusBar.tooltip = 'Run yapi on current file';
	context.subscriptions.push(statusBar);

	// Show status bar only for yapi files
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
			// Only auto-run if we already have an output panel open
			if (outputEditor && vscode.window.visibleTextEditors.includes(outputEditor)) {
				runYapi();
			}
		}
	}, null, context.subscriptions);
}

export function deactivate() {}
