import * as vscode from 'vscode';

export function activate(context: vscode.ExtensionContext) {
	console.log('yapi extension is now active');

	// Register the run command
	const runCommand = vscode.commands.registerCommand('yapi.runCurrent', async () => {
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

		// Save the file first
		await editor.document.save();

		// Create/reuse terminal and run yapi
		let terminal = vscode.window.terminals.find(t => t.name === 'yapi');
		if (!terminal) {
			terminal = vscode.window.createTerminal('yapi');
		}
		terminal.show();
		terminal.sendText(`yapi -c "${filePath}"`);
	});

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
}

export function deactivate() {}
