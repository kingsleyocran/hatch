import * as vscode from 'vscode';
import * as os from 'os';
import * as path from 'path';
import { DaemonClient } from './daemon-client';
import { HatchSidebarProvider } from './sidebar-provider';
import { HatchStatusBar } from './status-bar';
import { BinaryManager } from './binary-manager';
import { suggestMapping } from './auto-detect';

const SOCKET_PATH = path.join(os.homedir(), '.hatch', 'hatch.sock');
const POLL_INTERVAL = 5000;

export function activate(context: vscode.ExtensionContext): void {
  const client = new DaemonClient(SOCKET_PATH);
  const sidebarProvider = new HatchSidebarProvider(context.extensionUri, client);
  const statusBar = new HatchStatusBar(client);
  const binaryManager = new BinaryManager();

  const sidebarRegistration = vscode.window.registerWebviewViewProvider(
    HatchSidebarProvider.viewType,
    sidebarProvider,
  );

  const pollTimer = setInterval(async () => {
    sidebarProvider.refresh();
    await statusBar.update();
  }, POLL_INTERVAL);

  context.subscriptions.push(
    sidebarRegistration,
    statusBar,
    { dispose: () => clearInterval(pollTimer) },
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('hatch.refresh', () => {
      sidebarProvider.refresh();
      statusBar.update();
    }),

    vscode.commands.registerCommand('hatch.addDomain', async () => {
      const domain = await vscode.window.showInputBox({
        prompt: 'Domain name (e.g. myapp.test)',
        placeHolder: 'myapp.test',
      });
      if (!domain) { return; }

      const portStr = await vscode.window.showInputBox({
        prompt: 'Port number',
        placeHolder: '3000',
        validateInput: (v) => {
          const n = parseInt(v, 10);
          return n > 0 && n < 65536 ? null : 'Enter a valid port (1-65535)';
        },
      });
      if (!portStr) { return; }

      const dir = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath || '';

      try {
        await client.add(domain, parseInt(portStr, 10), dir, false);
        vscode.window.showInformationMessage(`Mapped ${domain} to localhost:${portStr}`);
        sidebarProvider.refresh();
        statusBar.update();
      } catch (err) {
        vscode.window.showErrorMessage(`Failed: ${err}`);
      }
    }),

    vscode.commands.registerCommand('hatch.removeDomain', async (item) => {
      const domain = item?.project?.domain;
      if (!domain) { return; }

      try {
        await client.remove(domain);
        vscode.window.showInformationMessage(`Removed ${domain}`);
        sidebarProvider.refresh();
        statusBar.update();
      } catch (err) {
        vscode.window.showErrorMessage(`Failed: ${err}`);
      }
    }),

    vscode.commands.registerCommand('hatch.openDomain', async (item) => {
      let domain: string | undefined;

      if (item?.project?.domain) {
        domain = item.project.domain;
      } else {
        const workspaceDir = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
        if (workspaceDir) {
          try {
            const projects = await client.list();
            const match = projects.find((p) => p.dir === workspaceDir);
            domain = match?.domain;
          } catch {}
        }
      }

      if (domain) {
        const proto = item?.project?.https ? 'https' : 'http';
        vscode.env.openExternal(vscode.Uri.parse(`${proto}://${domain}`));
      }
    }),

    vscode.commands.registerCommand('hatch.scanPorts', async () => {
      const running = await client.isRunning();
      if (!running) {
        vscode.window.showWarningMessage('Start the Hatch daemon first.');
        return;
      }

      const { execFile } = require('child_process');
      const ports: Array<{ port: number; label: string }> = await new Promise((resolve) => {
        execFile('lsof', ['-iTCP', '-sTCP:LISTEN', '-n', '-P', '-F', 'pcn'],
          (err: Error | null, stdout: string) => {
            if (err || !stdout) { resolve([]); return; }
            const result: Array<{ port: number; label: string }> = [];
            const seen = new Set<number>();
            let pid = 0;
            for (const line of stdout.split('\n')) {
              if (!line) continue;
              if (line[0] === 'p') pid = parseInt(line.slice(1), 10);
              else if (line[0] === 'n') {
                const idx = line.lastIndexOf(':');
                if (idx >= 0) {
                  const port = parseInt(line.slice(idx + 1), 10);
                  if (port > 0 && !seen.has(port) && port !== 8443 && port !== 8444 && port !== 15353) {
                    seen.add(port);
                    result.push({ port, label: `:${port} (pid:${pid})` });
                  }
                }
              }
            }
            result.sort((a, b) => a.port - b.port);
            resolve(result);
          }
        );
      });

      if (ports.length === 0) {
        vscode.window.showInformationMessage('No unmapped ports detected.');
        return;
      }

      const picks = ports.map(p => ({
        label: p.label,
        port: p.port,
      }));

      const selected = await vscode.window.showQuickPick(picks, {
        placeHolder: 'Select a port to map to a domain',
        canPickMany: false,
      });

      if (!selected) return;

      const domain = await vscode.window.showInputBox({
        prompt: 'Domain name',
        placeHolder: 'myapp.test',
      });
      if (!domain) return;

      const dir = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath || '';

      try {
        await client.add(domain, selected.port, dir, false);
        vscode.window.showInformationMessage(`Mapped ${domain} to localhost:${selected.port}`);
        sidebarProvider.refresh();
        statusBar.update();
      } catch (err) {
        vscode.window.showErrorMessage(`Failed: ${err}`);
      }
    }),

    vscode.commands.registerCommand('hatch.mapPort', async (item: any) => {
      const port = item?.detected?.port;
      if (!port) return;

      const domain = await vscode.window.showInputBox({
        prompt: 'Domain name',
        placeHolder: 'myapp.test',
      });
      if (!domain) return;

      const dir = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath || '';

      try {
        await client.add(domain, port, dir, false);
        vscode.window.showInformationMessage(`Mapped ${domain} to localhost:${port}`);
        sidebarProvider.refresh();
        statusBar.update();
      } catch (err) {
        vscode.window.showErrorMessage(`Failed: ${err}`);
      }
    }),

    vscode.commands.registerCommand('hatch.startDaemon', async () => {
      try {
        const binary = await binaryManager.ensureBinary();
        const { exec } = require('child_process');
        exec(`"${binary}" start`, (err: Error | null) => {
          if (err) {
            vscode.window.showErrorMessage(`Failed to start daemon: ${err.message}`);
            return;
          }
          setTimeout(() => {
            sidebarProvider.refresh();
            statusBar.update();
          }, 2000);
          vscode.window.showInformationMessage('Hatch daemon started');
        });
      } catch (err) {
        vscode.window.showErrorMessage(`${err}`);
      }
    }),

    vscode.commands.registerCommand('hatch.runSetup', async () => {
      try {
        const binary = await binaryManager.ensureBinary();
        const platform = os.platform();
        const { exec } = require('child_process');

        let setupCmd: string;
        if (platform === 'darwin') {
          setupCmd = `osascript -e 'do shell script "${binary} setup" with administrator privileges'`;
        } else if (platform === 'linux') {
          setupCmd = `pkexec "${binary}" setup`;
        } else {
          setupCmd = `powershell -Command "Start-Process '${binary}' -ArgumentList 'setup' -Verb RunAs -Wait"`;
        }

        await vscode.window.withProgress(
          { location: vscode.ProgressLocation.Notification, title: 'Running Hatch setup...' },
          () => new Promise<void>((resolve, reject) => {
            exec(setupCmd, (err: Error | null) => {
              if (err) { reject(err); } else { resolve(); }
            });
          })
        );

        vscode.window.showInformationMessage('Hatch setup complete!');
        setTimeout(() => {
          sidebarProvider.refresh();
          statusBar.update();
        }, 1000);
      } catch (err) {
        vscode.window.showErrorMessage(`Setup failed: ${err}`);
      }
    }),

    vscode.commands.registerCommand('hatch.stopDaemon', async () => {
      try {
        await client.stop();
        vscode.window.showInformationMessage('Hatch daemon stopped');
        sidebarProvider.refresh();
        statusBar.update();
      } catch (err) {
        vscode.window.showErrorMessage(`Failed: ${err}`);
      }
    }),
  );

  binaryManager.findBinary().then(async (binary) => {
    if (!binary) {
      binaryManager.ensureBinary().catch(() => {});
      return;
    }

    const running = await client.isRunning();
    if (!running) {
      const { exec } = require('child_process');
      exec(`"${binary}" start`, () => {
        setTimeout(() => {
          sidebarProvider.refresh();
          statusBar.update();
        }, 2000);
      });
    }
  });

  statusBar.update();
  suggestMapping(client);
}

export function deactivate(): void {}
