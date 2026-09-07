import * as assert from 'assert';
import * as net from 'net';
import * as path from 'path';
import * as os from 'os';
import * as fs from 'fs';
import { DaemonClient } from '../../daemon-client';
import { HatchRequest, HatchResponse } from '../../types';

function createMockServer(
  socketPath: string,
  handler: (req: HatchRequest) => HatchResponse
): net.Server {
  const server = net.createServer((conn) => {
    let data = '';
    conn.on('data', (chunk) => {
      data += chunk.toString();
      try {
        const req: HatchRequest = JSON.parse(data);
        const resp = handler(req);
        conn.write(JSON.stringify(resp));
        conn.end();
      } catch {
        // wait for more data
      }
    });
  });
  server.listen(socketPath);
  return server;
}

suite('DaemonClient', () => {
  let socketPath: string;
  let server: net.Server;

  setup(() => {
    socketPath = path.join(
      os.tmpdir(),
      `hatch-test-${Date.now()}-${Math.random().toString(36).slice(2)}.sock`
    );
  });

  teardown((done) => {
    if (server) {
      server.close(() => {
        try { fs.unlinkSync(socketPath); } catch {}
        done();
      });
    } else {
      done();
    }
  });

  test('ping succeeds when daemon is running', async () => {
    server = createMockServer(socketPath, () => ({
      ok: true,
      message: 'pong',
    }));

    const client = new DaemonClient(socketPath);
    await client.ping();
  });

  test('isRunning returns false when no daemon', async () => {
    const client = new DaemonClient('/nonexistent/hatch.sock');
    const running = await client.isRunning();
    assert.strictEqual(running, false);
  });

  test('list returns projects', async () => {
    server = createMockServer(socketPath, () => ({
      ok: true,
      projects: [
        { domain: 'cayacart.test', port: 3000, dir: '/tmp', alive: true, https: false },
      ],
    }));

    const client = new DaemonClient(socketPath);
    const projects = await client.list();
    assert.strictEqual(projects.length, 1);
    assert.strictEqual(projects[0].domain, 'cayacart.test');
    assert.strictEqual(projects[0].alive, true);
  });

  test('add sends correct request', async () => {
    let receivedReq: HatchRequest | undefined;
    server = createMockServer(socketPath, (req) => {
      receivedReq = req;
      return { ok: true, message: 'mapped' };
    });

    const client = new DaemonClient(socketPath);
    await client.add('myapp.test', 3000, '/tmp/myapp', true);

    assert.strictEqual(receivedReq?.action, 'add');
    assert.strictEqual(receivedReq?.domain, 'myapp.test');
    assert.strictEqual(receivedReq?.port, 3000);
    assert.strictEqual(receivedReq?.dir, '/tmp/myapp');
    assert.strictEqual(receivedReq?.https, true);
  });

  test('remove sends correct request', async () => {
    let receivedReq: HatchRequest | undefined;
    server = createMockServer(socketPath, (req) => {
      receivedReq = req;
      return { ok: true, message: 'removed' };
    });

    const client = new DaemonClient(socketPath);
    await client.remove('myapp.test');
    assert.strictEqual(receivedReq?.action, 'remove');
    assert.strictEqual(receivedReq?.domain, 'myapp.test');
  });

  test('status returns daemon status', async () => {
    server = createMockServer(socketPath, () => ({
      ok: true,
      status: {
        running: true,
        uptime: '5m30s',
        domain_count: 3,
        active_count: 2,
      },
    }));

    const client = new DaemonClient(socketPath);
    const status = await client.status();
    assert.strictEqual(status.running, true);
    assert.strictEqual(status.domain_count, 3);
    assert.strictEqual(status.active_count, 2);
  });

  test('throws on error response', async () => {
    server = createMockServer(socketPath, () => ({
      ok: false,
      message: 'domain not found',
    }));

    const client = new DaemonClient(socketPath);
    await assert.rejects(
      () => client.remove('nonexistent.test'),
      /domain not found/
    );
  });
});
