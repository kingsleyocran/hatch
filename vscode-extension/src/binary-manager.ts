import * as vscode from 'vscode';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import * as https from 'https';
import { execFile } from 'child_process';

const GITHUB_REPO = 'kingsleyocran/hatch';

function getPlatformBinary(): string {
  const platform = os.platform();
  const arch = os.arch();

  let osName: string;
  switch (platform) {
    case 'darwin': osName = 'darwin'; break;
    case 'linux': osName = 'linux'; break;
    case 'win32': osName = 'windows'; break;
    default: throw new Error(`unsupported platform: ${platform}`);
  }

  let archName: string;
  switch (arch) {
    case 'arm64': archName = 'arm64'; break;
    case 'x64': archName = 'amd64'; break;
    default: throw new Error(`unsupported architecture: ${arch}`);
  }

  const ext = platform === 'win32' ? '.exe' : '';
  return `hatch-${osName}-${archName}${ext}`;
}

function hatchDir(): string {
  return path.join(os.homedir(), '.hatch');
}

function hatchBinDir(): string {
  return path.join(hatchDir(), 'bin');
}

export class BinaryManager {
  async findBinary(): Promise<string | null> {
    const localPath = path.join(hatchBinDir(), 'hatch');
    if (fs.existsSync(localPath)) {
      return localPath;
    }

    return new Promise((resolve) => {
      execFile('which', ['hatch'], (err, stdout) => {
        if (err || !stdout.trim()) {
          resolve(null);
        } else {
          resolve(stdout.trim());
        }
      });
    });
  }

  async downloadBinary(): Promise<string> {
    const binaryName = getPlatformBinary();
    const binDir = hatchBinDir();
    fs.mkdirSync(binDir, { recursive: true });

    const destPath = path.join(binDir, 'hatch');

    const releaseUrl = `https://api.github.com/repos/${GITHUB_REPO}/releases/latest`;

    return vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Downloading Hatch CLI...',
        cancellable: false,
      },
      async () => {
        const release = await this.fetchJSON(releaseUrl);
        const assets = release.assets as Array<{ name: string; browser_download_url: string }> | undefined;
        const asset = assets?.find(
          (a) => a.name === binaryName
        );

        if (!asset) {
          throw new Error(
            `No binary found for your platform (${binaryName}). ` +
            `Install manually: go install github.com/${GITHUB_REPO}@latest`
          );
        }

        await this.downloadFile(asset.browser_download_url, destPath);
        fs.chmodSync(destPath, 0o755);

        return destPath;
      }
    );
  }

  async ensureBinary(): Promise<string> {
    const existing = await this.findBinary();
    if (existing) {
      return existing;
    }

    const choice = await vscode.window.showWarningMessage(
      'Hatch CLI not found. Download it now?',
      'Download',
      'Install Manually'
    );

    if (choice === 'Download') {
      return this.downloadBinary();
    }

    const installUrl = `https://github.com/${GITHUB_REPO}#installation`;
    vscode.env.openExternal(vscode.Uri.parse(installUrl));
    throw new Error('Hatch CLI not installed');
  }

  private fetchJSON(url: string): Promise<Record<string, unknown>> {
    return new Promise((resolve, reject) => {
      const get = (u: string) => {
        https.get(u, { headers: { 'User-Agent': 'hatch-vscode' } }, (res) => {
          if (res.statusCode === 301 || res.statusCode === 302) {
            get(res.headers.location!);
            return;
          }
          let data = '';
          res.on('data', (chunk) => (data += chunk));
          res.on('end', () => {
            try {
              resolve(JSON.parse(data));
            } catch {
              reject(new Error(`invalid JSON from ${u}`));
            }
          });
        }).on('error', reject);
      };
      get(url);
    });
  }

  private downloadFile(url: string, dest: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const get = (u: string) => {
        https.get(u, { headers: { 'User-Agent': 'hatch-vscode' } }, (res) => {
          if (res.statusCode === 301 || res.statusCode === 302) {
            get(res.headers.location!);
            return;
          }
          const file = fs.createWriteStream(dest);
          res.pipe(file);
          file.on('finish', () => {
            file.close();
            resolve();
          });
        }).on('error', reject);
      };
      get(url);
    });
  }
}
