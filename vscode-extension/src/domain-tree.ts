import * as vscode from 'vscode';
import { execFile } from 'child_process';
import { DaemonClient } from './daemon-client';
import { ProjectStatus } from './types';

export class SectionItem extends vscode.TreeItem {
  constructor(
    public readonly label: string,
    public readonly sectionType: 'mapped' | 'unmapped',
    childCount: number,
  ) {
    super(label, childCount > 0 ? vscode.TreeItemCollapsibleState.Expanded : vscode.TreeItemCollapsibleState.None);
    this.contextValue = 'section';
  }
}

export class DomainItem extends vscode.TreeItem {
  constructor(public readonly project: ProjectStatus) {
    super(project.domain, vscode.TreeItemCollapsibleState.None);
    const statusText = project.alive ? 'active' : 'stopped';
    const httpsText = project.https ? ' https' : '';
    this.description = `:${project.port} ${statusText}${httpsText}`;
    this.tooltip = `${project.domain} → localhost:${project.port}\nStatus: ${statusText}\nDirectory: ${project.dir}`;
    this.iconPath = new vscode.ThemeIcon(
      project.alive ? 'circle-filled' : 'circle-outline',
      new vscode.ThemeColor(project.alive ? 'testing.iconPassed' : 'testing.iconFailed')
    );
    this.contextValue = 'domain';
  }
}

export interface DetectedPort {
  port: number;
  name: string;
  type: string;
  dir: string;
}

export class PortItem extends vscode.TreeItem {
  constructor(public readonly detected: DetectedPort) {
    super(`:${detected.port}`, vscode.TreeItemCollapsibleState.None);
    const typeStr = detected.type ? ` (${detected.type})` : '';
    this.description = `${detected.name}${typeStr}`;
    this.tooltip = `Port ${detected.port}\n${detected.name}\n${detected.dir}`;
    this.iconPath = new vscode.ThemeIcon('plug', new vscode.ThemeColor('charts.yellow'));
    this.contextValue = 'unmappedPort';
  }
}

type TreeNode = SectionItem | DomainItem | PortItem;

export class DomainTreeProvider implements vscode.TreeDataProvider<TreeNode> {
  private _onDidChangeTreeData = new vscode.EventEmitter<TreeNode | undefined>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  private mappedDomains: ProjectStatus[] = [];
  private unmappedPorts: DetectedPort[] = [];

  constructor(private client: DaemonClient) {}

  refresh(): void {
    this._onDidChangeTreeData.fire(undefined);
  }

  getTreeItem(element: TreeNode): vscode.TreeItem {
    return element;
  }

  async getChildren(element?: TreeNode): Promise<TreeNode[]> {
    if (!element) {
      const running = await this.client.isRunning();
      if (!running) {
        const offlineItem = new vscode.TreeItem('Daemon offline — click to start');
        offlineItem.command = { command: 'hatch.startDaemon', title: 'Start Daemon' };
        offlineItem.iconPath = new vscode.ThemeIcon('warning', new vscode.ThemeColor('testing.iconFailed'));
        return [offlineItem as TreeNode];
      }

      try {
        this.mappedDomains = await this.client.list();
      } catch {
        this.mappedDomains = [];
      }

      this.unmappedPorts = await this.scanUnmappedPorts();

      const sections: TreeNode[] = [];
      sections.push(new SectionItem('Mapped Domains', 'mapped', this.mappedDomains.length));
      sections.push(new SectionItem('Running Ports', 'unmapped', this.unmappedPorts.length));
      return sections;
    }

    if (element instanceof SectionItem) {
      if (element.sectionType === 'mapped') {
        return this.mappedDomains.map(p => new DomainItem(p));
      }
      if (element.sectionType === 'unmapped') {
        return this.unmappedPorts.map(p => new PortItem(p));
      }
    }

    return [];
  }

  private async scanUnmappedPorts(): Promise<DetectedPort[]> {
    const mappedPorts = new Set(this.mappedDomains.map(d => d.port));

    return new Promise((resolve) => {
      execFile('lsof', ['-iTCP', '-sTCP:LISTEN', '-n', '-P', '-F', 'pcn'], (err, stdout) => {
        if (err || !stdout) {
          resolve([]);
          return;
        }

        const ports: DetectedPort[] = [];
        let currentPID = 0;
        const seen = new Set<number>();

        for (const line of stdout.split('\n')) {
          if (!line) continue;
          if (line[0] === 'p') {
            currentPID = parseInt(line.slice(1), 10);
          } else if (line[0] === 'n' && line.includes('127.0.0.1')) {
            const idx = line.lastIndexOf(':');
            if (idx >= 0) {
              const port = parseInt(line.slice(idx + 1), 10);
              if (port > 0 && !mappedPorts.has(port) && !seen.has(port) && port !== 8443 && port !== 8444 && port !== 15353) {
                seen.add(port);
                ports.push({ port, name: `pid:${currentPID}`, type: '', dir: '' });
              }
            }
          }
        }

        resolve(ports);
      });
    });
  }
}
