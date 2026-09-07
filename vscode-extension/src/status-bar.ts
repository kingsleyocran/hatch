import * as vscode from 'vscode';
import { DaemonClient } from './daemon-client';

export class HatchStatusBar {
  private item: vscode.StatusBarItem;

  constructor(private client: DaemonClient) {
    this.item = vscode.window.createStatusBarItem(
      vscode.StatusBarAlignment.Left,
      100
    );
    this.item.command = 'hatch.openDomain';
    this.item.show();
  }

  async update(): Promise<void> {
    const running = await this.client.isRunning();
    if (!running) {
      this.item.text = '$(globe) Hatch: offline';
      this.item.tooltip = 'Hatch daemon is not running';
      this.item.command = 'hatch.startDaemon';
      return;
    }

    const workspaceDir = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
    if (!workspaceDir) {
      this.item.text = '$(globe) Hatch';
      this.item.tooltip = 'No workspace open';
      return;
    }

    try {
      const projects = await this.client.list();
      const match = projects.find((p) => p.dir === workspaceDir);

      if (match) {
        const proto = match.https ? 'https' : 'http';
        this.item.text = `$(globe) ${match.domain}`;
        this.item.tooltip = `${proto}://${match.domain} → localhost:${match.port}\nClick to open`;
        this.item.command = 'hatch.openDomain';
      } else {
        this.item.text = '$(globe) Hatch: no domain';
        this.item.tooltip = 'No domain mapped for this workspace';
        this.item.command = 'hatch.addDomain';
      }
    } catch {
      this.item.text = '$(globe) Hatch';
    }
  }

  dispose(): void {
    this.item.dispose();
  }
}
