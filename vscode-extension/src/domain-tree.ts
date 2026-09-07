import * as vscode from 'vscode';
import { DaemonClient } from './daemon-client';
import { ProjectStatus } from './types';

export class DomainItem extends vscode.TreeItem {
  constructor(
    public readonly project: ProjectStatus,
  ) {
    super(project.domain, vscode.TreeItemCollapsibleState.None);

    const statusText = project.alive ? 'active' : 'stopped';
    const httpsText = project.https ? ' (https)' : '';

    this.description = `:${project.port} ${statusText}${httpsText}`;
    this.tooltip = `${project.domain} → localhost:${project.port}\nStatus: ${statusText}\nDirectory: ${project.dir}`;
    this.iconPath = new vscode.ThemeIcon(
      project.alive ? 'circle-filled' : 'circle-outline',
      new vscode.ThemeColor(project.alive ? 'testing.iconPassed' : 'testing.iconFailed')
    );
    this.contextValue = 'domain';
  }
}

export class DomainTreeProvider implements vscode.TreeDataProvider<DomainItem> {
  private _onDidChangeTreeData = new vscode.EventEmitter<DomainItem | undefined>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  constructor(private client: DaemonClient) {}

  refresh(): void {
    this._onDidChangeTreeData.fire(undefined);
  }

  getTreeItem(element: DomainItem): vscode.TreeItem {
    return element;
  }

  async getChildren(): Promise<DomainItem[]> {
    const running = await this.client.isRunning();
    if (!running) {
      return [];
    }

    try {
      const projects = await this.client.list();
      return projects.map((p) => new DomainItem(p));
    } catch {
      return [];
    }
  }
}
