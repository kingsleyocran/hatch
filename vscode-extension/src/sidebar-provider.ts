import * as vscode from 'vscode';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { execFile } from 'child_process';
import { DaemonClient } from './daemon-client';
import { detectProject } from './auto-detect';
import { ProjectStatus } from './types';

interface DetectedPort {
  port: number;
  pid: number;
  process: string;
  name: string;
  type: string;
  dir: string;
}

export class HatchSidebarProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = 'hatchDomains';
  private _view?: vscode.WebviewView;

  constructor(
    private readonly _extensionUri: vscode.Uri,
    private readonly client: DaemonClient,
  ) {}

  public resolveWebviewView(
    webviewView: vscode.WebviewView,
    _context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken,
  ): void {
    this._view = webviewView;
    webviewView.webview.options = { enableScripts: true };

    webviewView.webview.onDidReceiveMessage(async (msg) => {
      switch (msg.type) {
        case 'runSetup':
          vscode.commands.executeCommand('hatch.runSetup');
          break;
        case 'addDomain':
          vscode.commands.executeCommand('hatch.addDomain');
          break;
        case 'removeDomain':
          vscode.commands.executeCommand('hatch.removeDomain', { project: { domain: msg.domain } });
          break;
        case 'openDomain': {
          const proto = msg.https ? 'https' : 'http';
          vscode.env.openExternal(vscode.Uri.parse(`${proto}://${msg.domain}`));
          break;
        }
        case 'mapPort':
          vscode.commands.executeCommand('hatch.mapPort', { detected: { port: msg.port } });
          break;
        case 'startDaemon':
          vscode.commands.executeCommand('hatch.startDaemon');
          break;
        case 'refresh':
          this.refresh();
          break;
        case 'scanPorts':
          vscode.commands.executeCommand('hatch.scanPorts');
          break;
      }
    });

    this.refresh();
  }

  public async refresh(): Promise<void> {
    if (!this._view) return;

    const setupDone = this.isSetupDone();
    const running = await this.client.isRunning();

    let domains: ProjectStatus[] = [];
    let ports: DetectedPort[] = [];

    if (running) {
      try { domains = await this.client.list(); } catch {}
      ports = await this.scanUnmappedPorts(domains);
    }

    this._view.webview.html = this.getHtml(setupDone, running, domains, ports);
  }

  private isSetupDone(): boolean {
    const platform = os.platform();
    if (platform === 'darwin') return fs.existsSync('/etc/resolver/test');
    if (platform === 'linux') return fs.existsSync('/etc/systemd/resolved.conf.d/hatch.conf');
    return fs.existsSync(path.join(os.homedir(), '.hatch', '.setup-done'));
  }

  private async scanUnmappedPorts(mapped: ProjectStatus[]): Promise<DetectedPort[]> {
    const mappedPorts = new Set(mapped.map(d => d.port));

    const rawPorts = await new Promise<Array<{ port: number; pid: number; process: string }>>((resolve) => {
      execFile('lsof', ['-iTCP', '-sTCP:LISTEN', '-n', '-P', '-F', 'pcn'], (err, stdout) => {
        if (err || !stdout) { resolve([]); return; }
        const result: Array<{ port: number; pid: number; process: string }> = [];
        let pid = 0, cmd = '';
        const seen = new Set<number>();
        for (const line of stdout.split('\n')) {
          if (!line) continue;
          if (line[0] === 'p') { pid = parseInt(line.slice(1), 10); cmd = ''; }
          else if (line[0] === 'c') { cmd = line.slice(1); }
          else if (line[0] === 'n') {
            const idx = line.lastIndexOf(':');
            if (idx >= 0) {
              const port = parseInt(line.slice(idx + 1), 10);
              if (port > 0 && !mappedPorts.has(port) && !seen.has(port) && port !== 8443 && port !== 8444 && port !== 15353) {
                seen.add(port);
                result.push({ port, pid, process: cmd });
              }
            }
          }
        }
        resolve(result);
      });
    });

    const ports: DetectedPort[] = await Promise.all(
      rawPorts.map(async (raw) => {
        const dir = await this.getProcessDir(raw.pid);
        let name = dir ? path.basename(dir) : '';
        let type = '';
        if (dir) {
          const project = detectProject(dir);
          if (project) { name = project.name; type = project.type; }
        }
        return { ...raw, name, type, dir };
      })
    );

    ports.sort((a, b) => a.port - b.port);
    return ports;
  }

  private getProcessDir(pid: number): Promise<string> {
    return new Promise((resolve) => {
      execFile('lsof', ['-a', '-p', String(pid), '-d', 'cwd', '-Fn'], (err, stdout) => {
        if (err || !stdout) { resolve(''); return; }
        for (const line of stdout.split('\n')) {
          if (line.startsWith('n/')) { resolve(line.slice(1)); return; }
        }
        resolve('');
      });
    });
  }

  private getHtml(setupDone: boolean, running: boolean, domains: ProjectStatus[], ports: DetectedPort[]): string {
    if (!setupDone) {
      return `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>${this.getStyles()}</style></head>
<body>
  <div class="center-content">
    <div class="setup-icon">&#x2B21;</div>
    <h2>Welcome to Hatch</h2>
    <p class="description">Hatch maps custom local domains to your running dev servers. One-time setup is needed to configure DNS resolution and port forwarding.</p>
    <p class="description muted">This requires admin access — you'll be prompted for your password.</p>
    <button class="primary-btn" onclick="post('runSetup')">Run Setup</button>
  </div>
  <script>${this.getScript()}</script>
</body></html>`;
    }

    if (!running) {
      return `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>${this.getStyles()}</style></head>
<body>
  <div class="center-content">
    <div class="setup-icon">&#x2B21;</div>
    <h2>Daemon Offline</h2>
    <p class="description">The Hatch daemon isn't running. Start it to manage your local domains.</p>
    <button class="primary-btn" onclick="post('startDaemon')">Start Daemon</button>
  </div>
  <script>${this.getScript()}</script>
</body></html>`;
    }

    const domainItems = domains.map(d => {
      const status = d.alive ? 'active' : 'stopped';
      const statusClass = d.alive ? 'status-active' : 'status-stopped';
      const dot = d.alive ? '&#x25CF;' : '&#x25CB;';
      const proto = d.https ? 'https' : 'http';
      return `<div class="item">
        <div class="item-main">
          <span class="${statusClass}">${dot}</span>
          <span class="item-name">${d.domain}</span>
          <span class="item-port">:${d.port}</span>
        </div>
        <div class="item-meta">${status}${d.https ? ' &middot; https' : ''}</div>
        <div class="item-actions">
          <button class="icon-btn" title="Open in browser" onclick="post('openDomain', { domain: '${d.domain}', https: ${d.https} })">&#x2197;</button>
          <button class="icon-btn danger" title="Remove" onclick="post('removeDomain', { domain: '${d.domain}' })">&#x2715;</button>
        </div>
      </div>`;
    }).join('');

    const portItems = ports.map(p => {
      const label = p.name || `pid:${p.pid}`;
      const typeStr = p.type ? ` &middot; ${p.type}` : '';
      return `<div class="item">
        <div class="item-main">
          <span class="status-port">&#x26A1;</span>
          <span class="item-port-num">:${p.port}</span>
          <span class="item-name">${label}</span>
        </div>
        <div class="item-meta">${p.process}${typeStr}</div>
        <div class="item-actions">
          <button class="icon-btn add" title="Map to domain" onclick="post('mapPort', { port: ${p.port} })">+</button>
        </div>
      </div>`;
    }).join('');

    return `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>${this.getStyles()}</style></head>
<body>
  <div class="toolbar">
    <button class="toolbar-btn" onclick="post('addDomain')" title="Add domain">+</button>
    <button class="toolbar-btn" onclick="post('scanPorts')" title="Scan ports">&#x27F3;</button>
    <button class="toolbar-btn" onclick="post('refresh')" title="Refresh">&#x21BB;</button>
  </div>

  <div class="section">
    <div class="section-header">Mapped Domains</div>
    ${domainItems || '<div class="empty">No domains mapped yet</div>'}
  </div>

  <div class="section">
    <div class="section-header">Running Ports</div>
    ${portItems || '<div class="empty">No unmapped ports detected</div>'}
  </div>

  <script>${this.getScript()}</script>
</body></html>`;
  }

  private getStyles(): string {
    return `
      * { margin: 0; padding: 0; box-sizing: border-box; }
      body {
        font-family: var(--vscode-font-family);
        font-size: var(--vscode-font-size);
        color: var(--vscode-foreground);
        background: var(--vscode-sideBar-background);
        padding: 0;
      }
      .center-content {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: 32px 16px;
        text-align: center;
        min-height: 200px;
      }
      .setup-icon {
        font-size: 32px;
        margin-bottom: 12px;
        opacity: 0.6;
      }
      h2 {
        font-size: 14px;
        font-weight: 600;
        margin-bottom: 8px;
        color: var(--vscode-foreground);
      }
      .description {
        font-size: 12px;
        line-height: 1.5;
        color: var(--vscode-descriptionForeground);
        margin-bottom: 8px;
        max-width: 240px;
      }
      .muted { opacity: 0.7; }
      .primary-btn {
        margin-top: 12px;
        padding: 6px 16px;
        background: var(--vscode-button-background);
        color: var(--vscode-button-foreground);
        border: none;
        border-radius: 3px;
        cursor: pointer;
        font-size: 12px;
        font-family: var(--vscode-font-family);
      }
      .primary-btn:hover { background: var(--vscode-button-hoverBackground); }
      .toolbar {
        display: flex;
        justify-content: flex-end;
        gap: 4px;
        padding: 6px 8px;
        border-bottom: 1px solid var(--vscode-sideBarSectionHeader-border);
      }
      .toolbar-btn {
        background: none;
        border: none;
        color: var(--vscode-foreground);
        cursor: pointer;
        padding: 2px 6px;
        font-size: 14px;
        border-radius: 3px;
        opacity: 0.7;
      }
      .toolbar-btn:hover { opacity: 1; background: var(--vscode-toolbar-hoverBackground); }
      .section { padding: 0; }
      .section-header {
        font-size: 11px;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.5px;
        padding: 8px 12px 4px;
        color: var(--vscode-sideBarSectionHeader-foreground);
        background: var(--vscode-sideBarSectionHeader-background);
      }
      .item {
        padding: 6px 12px;
        display: flex;
        flex-direction: column;
        gap: 2px;
        position: relative;
        border-bottom: 1px solid var(--vscode-sideBarSectionHeader-border, transparent);
      }
      .item:hover { background: var(--vscode-list-hoverBackground); }
      .item-main {
        display: flex;
        align-items: center;
        gap: 6px;
      }
      .item-name {
        font-size: 12px;
        font-weight: 500;
        flex: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .item-port, .item-port-num {
        font-size: 11px;
        opacity: 0.7;
        font-family: var(--vscode-editor-font-family);
      }
      .item-meta {
        font-size: 11px;
        color: var(--vscode-descriptionForeground);
        padding-left: 20px;
      }
      .item-actions {
        position: absolute;
        right: 8px;
        top: 50%;
        transform: translateY(-50%);
        display: none;
        gap: 2px;
      }
      .item:hover .item-actions { display: flex; }
      .icon-btn {
        background: none;
        border: none;
        color: var(--vscode-foreground);
        cursor: pointer;
        padding: 2px 5px;
        font-size: 12px;
        border-radius: 3px;
        opacity: 0.7;
      }
      .icon-btn:hover { opacity: 1; background: var(--vscode-toolbar-hoverBackground); }
      .icon-btn.danger:hover { color: var(--vscode-errorForeground); }
      .icon-btn.add { font-size: 16px; font-weight: bold; }
      .icon-btn.add:hover { color: var(--vscode-terminal-ansiGreen); }
      .status-active { color: var(--vscode-terminal-ansiGreen); font-size: 10px; }
      .status-stopped { color: var(--vscode-errorForeground); font-size: 10px; }
      .status-port { font-size: 10px; color: var(--vscode-terminal-ansiYellow); }
      .empty {
        padding: 12px;
        font-size: 12px;
        color: var(--vscode-descriptionForeground);
        text-align: center;
        font-style: italic;
      }
    `;
  }

  private getScript(): string {
    return `
      const vscode = acquireVsCodeApi();
      function post(type, data) {
        vscode.postMessage({ type, ...data });
      }
    `;
  }
}
