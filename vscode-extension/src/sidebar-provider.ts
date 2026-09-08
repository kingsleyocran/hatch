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
    <div class="setup-icon"><svg viewBox="0 0 117.1 107.32" width="48" height="44"><path fill="var(--vscode-foreground)" opacity="0.5" d="M83.92 84.85l5.77 0c3.45,0 5.94,2.82 5.54,6.27l-0.58 5.01c-0.72,6.18 3.71,11.19 9.88,11.19 6.18,0 11.77,-5.01 12.49,-11.19 0.72,-6.18 -3.71,-11.19 -9.88,-11.19l-5.12 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.59 -5.04 0.11 -0.93 4.54 -39.02c0.71,-6.14 -4.11,-11.17 -10.74,-11.23 -0.13,0.01 -0.26,0.01 -0.39,0.01l-4.49 0 -5.55 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.58 -5.01c0.72,-6.18 -3.71,-11.19 -9.88,-11.19 -6.18,0 -11.77,5.01 -12.49,11.19 -0.72,6.18 3.71,11.19 9.88,11.19l5.12 0c3.45,0 5.94,2.82 5.54,6.27l-0.59 5.04c-0,0 -0,0.01 -0,0.01l-0.11 0.92 -1.88 16.2c-0.64,-5.61 -5.17,-9.79 -11.17,-9.79 -5.99,0 -11.48,4.16 -13.44,9.75l1.99 -17.09c0.71,-6.14 -4.11,-11.17 -10.74,-11.23 -0.13,0.01 -0.26,0.01 -0.39,0.01l-4.49 0 -5.55 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.58 -5.01c0.72,-6.18 -3.71,-11.19 -9.88,-11.19 -6.18,0 -11.77,5.01 -12.49,11.19 -0.72,6.18 3.71,11.19 9.88,11.19l5.12 0c3.45,0 5.94,2.82 5.54,6.27l-0.59 5.04c-0,0 -0,0.01 -0,0.01l-0.11 0.92 -4.54 39.01c-0.72,6.18 4.17,11.23 10.86,11.23l0.04 0c0.08,-0 0.15,-0 0.23,-0l4.26 0 0.01 -0 5.77 0c3.45,0 5.94,2.82 5.54,6.27l-0.58 5.01c-0.72,6.18 3.71,11.19 9.88,11.19 6.18,0 11.77,-5.01 12.49,-11.19 0.72,-6.18 -3.71,-11.19 -9.88,-11.19l-5.12 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.59 -5.04c0,-0.01 0,-0.01 0,-0.02l0.11 -0.92 1.88 -16.18c0.65,5.59 5.18,9.75 11.17,9.75 6,0 11.5,-4.18 13.45,-9.78l-1.99 17.12c-0.72,6.18 4.17,11.23 10.86,11.23l0.04 0c0.08,-0 0.15,-0 0.23,-0l4.26 0 0.01 -0z"/></svg></div>
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
    <div class="setup-icon"><svg viewBox="0 0 117.1 107.32" width="48" height="44"><path fill="var(--vscode-foreground)" opacity="0.5" d="M83.92 84.85l5.77 0c3.45,0 5.94,2.82 5.54,6.27l-0.58 5.01c-0.72,6.18 3.71,11.19 9.88,11.19 6.18,0 11.77,-5.01 12.49,-11.19 0.72,-6.18 -3.71,-11.19 -9.88,-11.19l-5.12 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.59 -5.04 0.11 -0.93 4.54 -39.02c0.71,-6.14 -4.11,-11.17 -10.74,-11.23 -0.13,0.01 -0.26,0.01 -0.39,0.01l-4.49 0 -5.55 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.58 -5.01c0.72,-6.18 -3.71,-11.19 -9.88,-11.19 -6.18,0 -11.77,5.01 -12.49,11.19 -0.72,6.18 3.71,11.19 9.88,11.19l5.12 0c3.45,0 5.94,2.82 5.54,6.27l-0.59 5.04c-0,0 -0,0.01 -0,0.01l-0.11 0.92 -1.88 16.2c-0.64,-5.61 -5.17,-9.79 -11.17,-9.79 -5.99,0 -11.48,4.16 -13.44,9.75l1.99 -17.09c0.71,-6.14 -4.11,-11.17 -10.74,-11.23 -0.13,0.01 -0.26,0.01 -0.39,0.01l-4.49 0 -5.55 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.58 -5.01c0.72,-6.18 -3.71,-11.19 -9.88,-11.19 -6.18,0 -11.77,5.01 -12.49,11.19 -0.72,6.18 3.71,11.19 9.88,11.19l5.12 0c3.45,0 5.94,2.82 5.54,6.27l-0.59 5.04c-0,0 -0,0.01 -0,0.01l-0.11 0.92 -4.54 39.01c-0.72,6.18 4.17,11.23 10.86,11.23l0.04 0c0.08,-0 0.15,-0 0.23,-0l4.26 0 0.01 -0 5.77 0c3.45,0 5.94,2.82 5.54,6.27l-0.58 5.01c-0.72,6.18 3.71,11.19 9.88,11.19 6.18,0 11.77,-5.01 12.49,-11.19 0.72,-6.18 -3.71,-11.19 -9.88,-11.19l-5.12 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.59 -5.04c0,-0.01 0,-0.01 0,-0.02l0.11 -0.92 1.88 -16.18c0.65,5.59 5.18,9.75 11.17,9.75 6,0 11.5,-4.18 13.45,-9.78l-1.99 17.12c-0.72,6.18 4.17,11.23 10.86,11.23l0.04 0c0.08,-0 0.15,-0 0.23,-0l4.26 0 0.01 -0z"/></svg></div>
    <h2>Daemon Offline</h2>
    <p class="description">The Hatch daemon isn't running. Start it to manage your local domains.</p>
    <button class="primary-btn" onclick="post('startDaemon')">Start Daemon</button>
  </div>
  <script>${this.getScript()}</script>
</body></html>`;
    }

    const domainItems = domains.map(d => {
      const statusDot = d.alive
        ? '<span class="dot dot-green"></span>'
        : '<span class="dot dot-red"></span>';
      const statusLabel = d.alive ? 'Active' : 'Stopped';
      const httpsLabel = d.https ? ' · HTTPS' : '';
      return `<div class="card">
        <div class="card-row">
          ${statusDot}
          <div class="card-info">
            <span class="card-title">${d.domain}</span>
            <span class="card-sub">Port ${d.port} · ${statusLabel}${httpsLabel}</span>
          </div>
          <div class="card-actions">
            <button class="act-btn" title="Open in browser" onclick="post('openDomain', { domain: '${d.domain}', https: ${d.https} })">
              <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor"><path d="M1.5 1h6v1H2v12h12V8.5h1v6.5H1V1h.5zm7 0H14v5.5h-1V2.707L7.354 8.354l-.708-.708L12.293 2H8.5V1z"/></svg>
            </button>
            <button class="act-btn act-danger" title="Remove mapping" onclick="post('removeDomain', { domain: '${d.domain}' })">
              <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor"><path d="M8 8.707l3.646 3.647.708-.708L8.707 8l3.647-3.646-.708-.708L8 7.293 4.354 3.646l-.708.708L7.293 8l-3.647 3.646.708.708L8 8.707z"/></svg>
            </button>
          </div>
        </div>
      </div>`;
    }).join('');

    const portItems = ports.map(p => {
      const label = p.name || `pid:${p.pid}`;
      const meta = [p.process, p.type].filter(Boolean).join(' · ');
      return `<div class="card">
        <div class="card-row">
          <span class="dot dot-amber"></span>
          <div class="card-info">
            <span class="card-title">${label}</span>
            <span class="card-sub">Port ${p.port}${meta ? ' · ' + meta : ''}</span>
          </div>
          <div class="card-actions">
            <button class="act-btn act-add" title="Map to domain" onclick="post('mapPort', { port: ${p.port} })">
              <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor"><path d="M8 1v7H1v1h7v7h1V9h7V8H9V1H8z"/></svg>
            </button>
          </div>
        </div>
      </div>`;
    }).join('');

    const domainCount = domains.length;
    const portCount = ports.length;

    return `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>${this.getStyles()}</style></head>
<body>
  <div class="section">
    <div class="section-header">
      <span>Mapped Domains</span>
      <div class="section-actions">
        <span class="badge">${domainCount}</span>
        <button class="hdr-btn" title="Add domain" onclick="post('addDomain')">
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor"><path d="M8 1v7H1v1h7v7h1V9h7V8H9V1H8z"/></svg>
        </button>
      </div>
    </div>
    <div class="card-list">
      ${domainItems || '<div class="empty">No domains mapped yet. Click + to add one.</div>'}
    </div>
  </div>

  <div class="section">
    <div class="section-header">
      <span>Running Ports</span>
      <div class="section-actions">
        <span class="badge">${portCount}</span>
        <button class="hdr-btn" title="Refresh" onclick="post('refresh')">
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor"><path d="M13 3.1V1h1v4h-4V4h2.5A5.5 5.5 0 003.05 5.1l-.9-.4A6.5 6.5 0 0113 3.1zM3 12.9V15H2v-4h4v1H3.5a5.5 5.5 0 009.45-1.1l.9.4A6.5 6.5 0 013 12.9z"/></svg>
        </button>
      </div>
    </div>
    <div class="card-list">
      ${portItems || '<div class="empty">No unmapped ports detected</div>'}
    </div>
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
        padding: 40px 20px;
        text-align: center;
        min-height: 100vh;
      }
      .setup-icon {
        font-size: 36px;
        margin-bottom: 16px;
        opacity: 0.4;
      }
      h2 {
        font-size: 13px;
        font-weight: 600;
        margin-bottom: 10px;
        color: var(--vscode-foreground);
      }
      .description {
        font-size: 11.5px;
        line-height: 1.6;
        color: var(--vscode-descriptionForeground);
        margin-bottom: 6px;
        max-width: 220px;
      }
      .muted { opacity: 0.6; font-size: 11px; }
      .primary-btn {
        margin-top: 16px;
        padding: 7px 20px;
        background: var(--vscode-button-background);
        color: var(--vscode-button-foreground);
        border: none;
        border-radius: 4px;
        cursor: pointer;
        font-size: 12px;
        font-weight: 500;
        font-family: var(--vscode-font-family);
        transition: background 0.15s;
      }
      .primary-btn:hover { background: var(--vscode-button-hoverBackground); }
      .section { margin-bottom: 4px; }
      .section-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        font-size: 11px;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.5px;
        padding: 10px 12px 6px;
        color: var(--vscode-sideBarSectionHeader-foreground);
      }
      .section-actions {
        display: flex;
        align-items: center;
        gap: 4px;
      }
      .badge {
        font-size: 10px;
        font-weight: 600;
        min-width: 18px;
        height: 18px;
        line-height: 18px;
        text-align: center;
        border-radius: 9px;
        background: var(--vscode-badge-background);
        color: var(--vscode-badge-foreground);
      }
      .hdr-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 22px;
        height: 22px;
        background: none;
        border: none;
        border-radius: 4px;
        color: var(--vscode-foreground);
        cursor: pointer;
        opacity: 0.6;
        transition: opacity 0.15s, background 0.15s;
      }
      .hdr-btn:hover { opacity: 1; background: var(--vscode-toolbar-hoverBackground); }
      .card-list { padding: 0 8px 4px; }
      .card {
        margin-bottom: 4px;
        padding: 8px 10px;
        border-radius: 6px;
        background: var(--vscode-list-hoverBackground, rgba(255,255,255,0.04));
        transition: background 0.15s;
      }
      .card:hover { background: var(--vscode-list-activeSelectionBackground, rgba(255,255,255,0.08)); }
      .card-row {
        display: flex;
        align-items: center;
        gap: 10px;
      }
      .dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        flex-shrink: 0;
      }
      .dot-green { background: #3fb950; }
      .dot-red { background: #f85149; }
      .dot-amber { background: #d29922; }
      .card-info {
        flex: 1;
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 1px;
      }
      .card-title {
        font-size: 12.5px;
        font-weight: 600;
        color: var(--vscode-foreground);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .card-sub {
        font-size: 11px;
        color: var(--vscode-descriptionForeground);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .card-actions {
        display: flex;
        gap: 2px;
        opacity: 0;
        transition: opacity 0.15s;
      }
      .card:hover .card-actions { opacity: 1; }
      .act-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 24px;
        height: 24px;
        background: none;
        border: none;
        border-radius: 4px;
        color: var(--vscode-foreground);
        cursor: pointer;
        opacity: 0.7;
        transition: opacity 0.15s, background 0.15s;
      }
      .act-btn:hover { opacity: 1; background: var(--vscode-toolbar-hoverBackground); }
      .act-danger:hover { color: #f85149; }
      .act-add:hover { color: #3fb950; }
      .empty {
        padding: 16px 12px;
        font-size: 11.5px;
        color: var(--vscode-descriptionForeground);
        text-align: center;
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
