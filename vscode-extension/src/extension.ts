import * as vscode from 'vscode';

export function activate(context: vscode.ExtensionContext): void {
  vscode.window.showInformationMessage('Hatch activated');
}

export function deactivate(): void {
  // cleanup
}
