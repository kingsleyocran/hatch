import * as vscode from 'vscode';
import * as os from 'os';
import * as path from 'path';
import { DaemonClient } from './daemon-client';
import { DomainTreeProvider } from './domain-tree';
import { HatchStatusBar } from './status-bar';
import { BinaryManager } from './binary-manager';
import { suggestMapping } from './auto-detect';

const SOCKET_PATH = path.join(os.homedir(), '.hatch', 'hatch.sock');
const POLL_INTERVAL = 5000;

export function activate(context: vscode.ExtensionContext): void {
  const client = new DaemonClient(SOCKET_PATH);
  const treeProvider = new DomainTreeProvider(client);
  const statusBar = new HatchStatusBar(client);
  const binaryManager = new BinaryManager();

  const treeView = vscode.window.createTreeView('hatchDomains', {
    treeDataProvider: treeProvider,
  });

  const pollTimer = setInterval(async () => {
    treeProvider.refresh();
    await statusBar.update();
  }, POLL_INTERVAL);

  context.subscriptions.push(
    treeView,
    statusBar,
    { dispose: () => clearInterval(pollTimer) },
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('hatch.refresh', () => {
      treeProvider.refresh();
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
        treeProvider.refresh();
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
        treeProvider.refresh();
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
      vscode.window.showInformationMessage(
        'Run "hatch scan" in your terminal for interactive port scanning.'
      );
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
            treeProvider.refresh();
            statusBar.update();
          }, 2000);
          vscode.window.showInformationMessage('Hatch daemon started');
        });
      } catch (err) {
        vscode.window.showErrorMessage(`${err}`);
      }
    }),

    vscode.commands.registerCommand('hatch.stopDaemon', async () => {
      try {
        await client.stop();
        vscode.window.showInformationMessage('Hatch daemon stopped');
        treeProvider.refresh();
        statusBar.update();
      } catch (err) {
        vscode.window.showErrorMessage(`Failed: ${err}`);
      }
    }),
  );

  binaryManager.findBinary().then((binary) => {
    if (!binary) {
      binaryManager.ensureBinary().catch(() => {});
    }
  });

  statusBar.update();
  suggestMapping(client);
}

export function deactivate(): void {}
