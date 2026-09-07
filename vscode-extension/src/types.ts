export interface HatchRequest {
  action: string;
  domain?: string;
  port?: number;
  dir?: string;
  https?: boolean;
}

export interface ProjectStatus {
  domain: string;
  port: number;
  dir: string;
  alive: boolean;
  https: boolean;
}

export interface DaemonStatus {
  running: boolean;
  uptime: string;
  domain_count: number;
  active_count: number;
}

export interface HatchResponse {
  ok: boolean;
  message?: string;
  projects?: ProjectStatus[];
  status?: DaemonStatus;
}

export const Actions = {
  Ping: 'ping',
  Add: 'add',
  Remove: 'remove',
  List: 'list',
  Stop: 'stop',
  Status: 'status',
} as const;
