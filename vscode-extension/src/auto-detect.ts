import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';
import { DaemonClient } from './daemon-client';

interface DetectedProject {
  name: string;
  type: string;
}

export function detectProject(dir: string): DetectedProject | null {
  const packageJsonPath = path.join(dir, 'package.json');
  if (fs.existsSync(packageJsonPath)) {
    try {
      const pkg = JSON.parse(fs.readFileSync(packageJsonPath, 'utf-8'));
      if (pkg.name) {
        return { name: pkg.name, type: 'node' };
      }
    } catch {}
  }

  const goModPath = path.join(dir, 'go.mod');
  if (fs.existsSync(goModPath)) {
    const content = fs.readFileSync(goModPath, 'utf-8');
    const match = content.match(/^module\s+(.+)/m);
    if (match) {
      const parts = match[1].trim().split('/');
      return { name: parts[parts.length - 1], type: 'go' };
    }
  }

  const cargoPath = path.join(dir, 'Cargo.toml');
  if (fs.existsSync(cargoPath)) {
    const content = fs.readFileSync(cargoPath, 'utf-8');
    const match = content.match(/^name\s*=\s*"([^"]+)"/m);
    if (match) {
      return { name: match[1], type: 'rust' };
    }
  }

  const pyprojectPath = path.join(dir, 'pyproject.toml');
  if (fs.existsSync(pyprojectPath)) {
    const content = fs.readFileSync(pyprojectPath, 'utf-8');
    const match = content.match(/^name\s*=\s*"([^"]+)"/m);
    if (match) {
      return { name: match[1], type: 'python' };
    }
  }

  return null;
}

export async function suggestMapping(client: DaemonClient): Promise<void> {
  const workspaceDir = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
  if (!workspaceDir) {
    return;
  }

  const running = await client.isRunning();
  if (!running) {
    return;
  }

  try {
    const projects = await client.list();
    const alreadyMapped = projects.some((p) => p.dir === workspaceDir);
    if (alreadyMapped) {
      return;
    }
  } catch {
    return;
  }

  const detected = detectProject(workspaceDir);
  if (!detected) {
    return;
  }

  const choice = await vscode.window.showInformationMessage(
    `Map ${detected.name} to a local domain?`,
    'Yes',
    'No'
  );

  if (choice !== 'Yes') {
    return;
  }

  const domain = await vscode.window.showInputBox({
    prompt: 'Domain name',
    value: `${detected.name}.test`,
    placeHolder: 'myapp.test',
  });

  if (!domain) {
    return;
  }

  const portStr = await vscode.window.showInputBox({
    prompt: 'Port number',
    value: '3000',
    placeHolder: '3000',
    validateInput: (v) => {
      const n = parseInt(v, 10);
      return n > 0 && n < 65536 ? null : 'Enter a valid port (1-65535)';
    },
  });

  if (!portStr) {
    return;
  }

  try {
    await client.add(domain, parseInt(portStr, 10), workspaceDir, false);
    vscode.window.showInformationMessage(`Mapped ${domain} to localhost:${portStr}`);
  } catch (err) {
    vscode.window.showErrorMessage(`Failed to add domain: ${err}`);
  }
}
