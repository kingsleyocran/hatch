import * as net from 'net';
import {
  HatchRequest,
  HatchResponse,
  ProjectStatus,
  DaemonStatus,
  Actions,
} from './types';

export class DaemonClient {
  constructor(private socketPath: string) {}

  private send(req: HatchRequest): Promise<HatchResponse> {
    return new Promise((resolve, reject) => {
      const conn = net.createConnection(this.socketPath);
      let data = '';

      conn.on('connect', () => {
        conn.write(JSON.stringify(req));
      });

      conn.on('data', (chunk) => {
        data += chunk.toString();
      });

      conn.on('end', () => {
        try {
          resolve(JSON.parse(data));
        } catch (err) {
          reject(new Error(`invalid response: ${data}`));
        }
      });

      conn.on('error', (err) => {
        reject(err);
      });

      conn.setTimeout(5000, () => {
        conn.destroy();
        reject(new Error('connection timeout'));
      });
    });
  }

  private async sendChecked(req: HatchRequest): Promise<HatchResponse> {
    const resp = await this.send(req);
    if (!resp.ok) {
      throw new Error(resp.message || 'unknown error');
    }
    return resp;
  }

  async ping(): Promise<void> {
    await this.sendChecked({ action: Actions.Ping });
  }

  async isRunning(): Promise<boolean> {
    try {
      await this.ping();
      return true;
    } catch {
      return false;
    }
  }

  async add(
    domain: string,
    port: number,
    dir: string,
    https: boolean
  ): Promise<string> {
    const resp = await this.sendChecked({
      action: Actions.Add,
      domain,
      port,
      dir,
      https,
    });
    return resp.message || '';
  }

  async remove(domain: string): Promise<string> {
    const resp = await this.sendChecked({
      action: Actions.Remove,
      domain,
    });
    return resp.message || '';
  }

  async list(): Promise<ProjectStatus[]> {
    const resp = await this.sendChecked({ action: Actions.List });
    return resp.projects || [];
  }

  async status(): Promise<DaemonStatus> {
    const resp = await this.sendChecked({ action: Actions.Status });
    if (!resp.status) {
      throw new Error('no status in response');
    }
    return resp.status;
  }

  async stop(): Promise<void> {
    await this.sendChecked({ action: Actions.Stop });
  }
}
