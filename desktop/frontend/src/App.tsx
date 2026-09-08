import { useState, useEffect } from 'react';
import { GetDomains, GetStatus, AddDomain, RemoveDomain, GetConfig, SetConfig, IsDaemonRunning, StartDaemon, StopDaemon } from '../wailsjs/go/main/App';
import logoSvg from './assets/logo.svg';
import './App.css';

type Domain = { domain: string; port: number; dir: string; alive: boolean; https: boolean };
type Status = { running: boolean; uptime: string; domain_count: number; active_count: number };
type Config = { default_tld: string; auto_https: boolean; daemon_port: number; https_port: number; dns_port: number };

type View = 'dashboard' | 'settings' | 'help';

function App() {
  const [view, setView] = useState<View>('dashboard');
  const [domains, setDomains] = useState<Domain[]>([]);
  const [status, setStatus] = useState<Status | null>(null);
  const [config, setConfigState] = useState<Config | null>(null);
  const [daemonRunning, setDaemonRunning] = useState(false);

  const refresh = async () => {
    const running = await IsDaemonRunning();
    setDaemonRunning(running);
    if (running) {
      setDomains(await GetDomains());
      setStatus(await GetStatus());
    }
    setConfigState(await GetConfig());
  };

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, 5000);
    return () => clearInterval(interval);
  }, []);

  const handleAdd = async () => {
    const domain = prompt('Domain name (e.g. myapp.test)');
    if (!domain) return;
    const portStr = prompt('Port number', '3000');
    if (!portStr) return;
    const https = confirm('Enable HTTPS?');
    const err = await AddDomain(domain, parseInt(portStr), https);
    if (err) alert(err);
    refresh();
  };

  const handleRemove = async (domain: string) => {
    if (!confirm(`Remove ${domain}?`)) return;
    const err = await RemoveDomain(domain);
    if (err) alert(err);
    refresh();
  };

  const handleConfigChange = async (key: string, value: string) => {
    await SetConfig(key, value);
    refresh();
  };

  return (
    <div className="app">
      <div className="drag-bar" />
      <nav className="sidebar">
        <div className="logo">
          <img src={logoSvg} alt="Hatch" width="28" height="28" />
        </div>
        <button className={`nav-btn ${view === 'dashboard' ? 'active' : ''}`} onClick={() => setView('dashboard')} title="Dashboard">
          <svg width="20" height="20" viewBox="0 0 16 16" fill="currentColor"><path d="M2 2h5v5H2V2zm7 0h5v5H9V2zm-7 7h5v5H2V9zm7 0h5v5H9V9z"/></svg>
        </button>
        <button className={`nav-btn ${view === 'settings' ? 'active' : ''}`} onClick={() => setView('settings')} title="Settings">
          <svg width="20" height="20" viewBox="0 0 16 16" fill="currentColor"><path d="M8 4.754a3.246 3.246 0 100 6.492 3.246 3.246 0 000-6.492zM5.754 8a2.246 2.246 0 114.492 0 2.246 2.246 0 01-4.492 0z"/><path d="M9.796 1.343c-.527-1.79-3.065-1.79-3.592 0l-.094.319a.873.873 0 01-1.255.52l-.292-.16c-1.64-.892-3.433.902-2.54 2.541l.159.292a.873.873 0 01-.52 1.255l-.319.094c-1.79.527-1.79 3.065 0 3.592l.319.094a.873.873 0 01.52 1.255l-.16.292c-.892 1.64.901 3.434 2.541 2.54l.292-.159a.873.873 0 011.255.52l.094.319c.527 1.79 3.065 1.79 3.592 0l.094-.319a.873.873 0 011.255-.52l.292.16c1.64.893 3.434-.902 2.54-2.541l-.159-.292a.873.873 0 01.52-1.255l.319-.094c1.79-.527 1.79-3.065 0-3.592l-.319-.094a.873.873 0 01-.52-1.255l.16-.292c.893-1.64-.902-3.433-2.541-2.54l-.292.159a.873.873 0 01-1.255-.52l-.094-.319zm-2.633.283c.246-.835 1.428-.835 1.674 0l.094.319a1.873 1.873 0 002.693 1.115l.291-.16c.764-.415 1.6.42 1.184 1.185l-.159.292a1.873 1.873 0 001.116 2.692l.318.094c.835.246.835 1.428 0 1.674l-.319.094a1.873 1.873 0 00-1.115 2.693l.16.291c.415.764-.42 1.6-1.185 1.184l-.291-.159a1.873 1.873 0 00-2.693 1.116l-.094.318c-.246.835-1.428.835-1.674 0l-.094-.319a1.873 1.873 0 00-2.692-1.115l-.292.16c-.764.415-1.6-.42-1.184-1.185l.159-.291A1.873 1.873 0 001.945 8.93l-.319-.094c-.835-.246-.835-1.428 0-1.674l.319-.094A1.873 1.873 0 003.06 4.377l-.16-.292c-.415-.764.42-1.6 1.185-1.184l.292.159a1.873 1.873 0 002.692-1.115l.094-.319z"/></svg>
        </button>
        <button className={`nav-btn ${view === 'help' ? 'active' : ''}`} onClick={() => setView('help')} title="Help">
          <svg width="20" height="20" viewBox="0 0 16 16" fill="currentColor"><path d="M8 15A7 7 0 118 1a7 7 0 010 14zm0 1A8 8 0 108 0a8 8 0 000 16z"/><path d="M5.255 5.786a.237.237 0 00.241.247h.825c.138 0 .248-.113.266-.25.09-.656.54-1.134 1.342-1.134.686 0 1.314.343 1.314 1.168 0 .635-.374.927-.965 1.371-.673.489-1.206 1.06-1.168 1.987l.003.217a.25.25 0 00.25.246h.811a.25.25 0 00.25-.25v-.105c0-.718.273-.927 1.01-1.486.609-.463 1.244-.977 1.244-2.056 0-1.511-1.276-2.241-2.673-2.241-1.267 0-2.655.59-2.75 2.286zm1.557 5.763c0 .533.425.927 1.01.927.609 0 1.028-.394 1.028-.927 0-.552-.42-.94-1.029-.94-.584 0-1.009.388-1.009.94z"/></svg>
        </button>
        <div className="nav-spacer" />
        <div className={`status-dot ${daemonRunning ? 'online' : 'offline'}`} title={daemonRunning ? 'Daemon running' : 'Daemon offline'} />
      </nav>

      <main className="content">
        {view === 'dashboard' && (
          <Dashboard domains={domains} daemonRunning={daemonRunning} onAdd={handleAdd} onRemove={handleRemove} onRefresh={refresh} />
        )}
        {view === 'settings' && (
          <Settings config={config} status={status} daemonRunning={daemonRunning} onConfigChange={handleConfigChange} onStopDaemon={async () => { await StopDaemon(); refresh(); }} />
        )}
        {view === 'help' && <Help />}
      </main>
    </div>
  );
}

function Dashboard({ domains, daemonRunning, onAdd, onRemove, onRefresh }: {
  domains: Domain[]; daemonRunning: boolean; onAdd: () => void; onRemove: (d: string) => void; onRefresh: () => void;
}) {
  const [tab, setTab] = useState<'domains' | 'ports'>('domains');

  if (!daemonRunning) {
    return (
      <div className="center-content">
        <svg viewBox="0 0 117.1 107.32" width="48" height="44"><path fill="#B3D7FC" d="M83.92 84.85l5.77 0c3.45,0 5.94,2.82 5.54,6.27l-0.58 5.01c-0.72,6.18 3.71,11.19 9.88,11.19 6.18,0 11.77,-5.01 12.49,-11.19 0.72,-6.18 -3.71,-11.19 -9.88,-11.19l-5.12 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.59 -5.04 0.11 -0.93 4.54 -39.02c0.71,-6.14 -4.11,-11.17 -10.74,-11.23 -0.13,0.01 -0.26,0.01 -0.39,0.01l-4.49 0 -5.55 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.58 -5.01c0.72,-6.18 -3.71,-11.19 -9.88,-11.19 -6.18,0 -11.77,5.01 -12.49,11.19 -0.72,6.18 3.71,11.19 9.88,11.19l5.12 0c3.45,0 5.94,2.82 5.54,6.27l-0.59 5.04c-0,0 -0,0.01 -0,0.01l-0.11 0.92 -1.88 16.2c-0.64,-5.61 -5.17,-9.79 -11.17,-9.79 -5.99,0 -11.48,4.16 -13.44,9.75l1.99 -17.09c0.71,-6.14 -4.11,-11.17 -10.74,-11.23 -0.13,0.01 -0.26,0.01 -0.39,0.01l-4.49 0 -5.55 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.58 -5.01c0.72,-6.18 -3.71,-11.19 -9.88,-11.19 -6.18,0 -11.77,5.01 -12.49,11.19 -0.72,6.18 3.71,11.19 9.88,11.19l5.12 0c3.45,0 5.94,2.82 5.54,6.27l-0.59 5.04c-0,0 -0,0.01 -0,0.01l-0.11 0.92 -4.54 39.01c-0.72,6.18 4.17,11.23 10.86,11.23l0.04 0c0.08,-0 0.15,-0 0.23,-0l4.26 0 0.01 -0 5.77 0c3.45,0 5.94,2.82 5.54,6.27l-0.58 5.01c-0.72,6.18 3.71,11.19 9.88,11.19 6.18,0 11.77,-5.01 12.49,-11.19 0.72,-6.18 -3.71,-11.19 -9.88,-11.19l-5.12 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.59 -5.04c0,-0.01 0,-0.01 0,-0.02l0.11 -0.92 1.88 -16.18c0.65,5.59 5.18,9.75 11.17,9.75 6,0 11.5,-4.18 13.45,-9.78l-1.99 17.12c-0.72,6.18 4.17,11.23 10.86,11.23l0.04 0c0.08,-0 0.15,-0 0.23,-0l4.26 0 0.01 -0z"/></svg>
        <h2>Daemon Offline</h2>
        <p className="muted">The Hatch daemon isn't running.</p>
        <button className="start-btn" onClick={async () => { await StartDaemon(); setTimeout(onRefresh, 2000); }}>Start Daemon</button>
      </div>
    );
  }

  return (
    <div className="dashboard">
      <div className="tabs">
        <button className={`tab ${tab === 'domains' ? 'active' : ''}`} onClick={() => setTab('domains')}>Domains</button>
        <button className={`tab ${tab === 'ports' ? 'active' : ''}`} onClick={() => setTab('ports')}>Ports</button>
      </div>

      {tab === 'domains' && (
        <>
          <div className="section-header">
            <span>Mapped Domains</span>
            <div className="section-actions">
              <span className="badge">{domains.length}</span>
              <button className="icon-btn" onClick={onAdd} title="Add domain">+</button>
              <button className="icon-btn" onClick={onRefresh} title="Refresh">↻</button>
            </div>
          </div>
          <div className="card-list">
            {domains.length === 0 && <div className="empty">No domains mapped yet. Click + to add one.</div>}
            {domains.map(d => (
              <div className="card" key={d.domain}>
                <div className="card-row">
                  <span className={`dot ${d.alive ? 'dot-green' : 'dot-red'}`} />
                  <div className="card-info">
                    <span className="card-title">{d.domain}</span>
                    <span className="card-sub">
                      {d.https ? 'https' : 'http'}://{d.domain} → localhost:{d.port}
                    </span>
                    {d.dir && <span className="card-dir">{d.dir}</span>}
                  </div>
                  <div className="card-actions">
                    <button className="act-btn" onClick={() => window.open(`${d.https ? 'https' : 'http'}://${d.domain}`, '_blank')} title="Open">↗</button>
                    <button className="act-btn act-danger" onClick={() => onRemove(d.domain)} title="Remove">✕</button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </>
      )}

      {tab === 'ports' && (
        <>
          <div className="section-header">
            <span>Running Ports</span>
            <div className="section-actions">
              <button className="icon-btn" onClick={onRefresh} title="Refresh">↻</button>
            </div>
          </div>
          <div className="card-list">
            <div className="empty">Port scanning available in CLI: <code>hatch scan</code></div>
          </div>
        </>
      )}
    </div>
  );
}

function Settings({ config, status, daemonRunning, onConfigChange, onStopDaemon }: {
  config: Config | null; status: Status | null; daemonRunning: boolean;
  onConfigChange: (k: string, v: string) => void; onStopDaemon: () => void;
}) {
  return (
    <div className="settings">
      <div className="section-header"><span>General</span></div>
      <div className="card">
        <div className="setting-row">
          <span>Default TLD</span>
          <select value={config?.default_tld || 'test'} onChange={e => onConfigChange('default_tld', e.target.value)}>
            <option value="test">.test</option>
            <option value="local">.local</option>
            <option value="localhost">.localhost</option>
            <option value="dev">.dev</option>
            <option value="internal">.internal</option>
          </select>
        </div>
        <div className="setting-row">
          <span>Auto HTTPS</span>
          <label className="toggle">
            <input type="checkbox" checked={config?.auto_https || false} onChange={e => onConfigChange('auto_https', e.target.checked ? 'true' : 'false')} />
            <span className="toggle-slider" />
          </label>
        </div>
      </div>

      <div className="section-header"><span>Daemon</span></div>
      <div className="card">
        <div className="setting-row">
          <span>Status</span>
          <span className={daemonRunning ? 'text-green' : 'text-red'}>
            {daemonRunning ? `Running (${status?.uptime || ''})` : 'Offline'}
          </span>
        </div>
        <div className="setting-row">
          <span>Domains</span>
          <span>{status?.domain_count || 0} ({status?.active_count || 0} active)</span>
        </div>
        <div className="setting-row">
          <span>HTTP Port</span>
          <span>{config?.daemon_port || 80}</span>
        </div>
        <div className="setting-row">
          <span>HTTPS Port</span>
          <span>{config?.https_port || 443}</span>
        </div>
        {daemonRunning && (
          <button className="danger-btn" onClick={onStopDaemon}>Stop Daemon</button>
        )}
      </div>
    </div>
  );
}

function Help() {
  return (
    <div className="help">
      <div className="section-header"><span>Quick Start</span></div>
      <div className="card help-card">
        <p>Hatch maps custom local domains to your running dev servers.</p>
        <pre>{`# Add a domain\nhatch add myapp.test 3000 --https\n\n# List domains\nhatch ls\n\n# Scan running ports\nhatch scan`}</pre>
      </div>

      <div className="section-header"><span>CLI Commands</span></div>
      <div className="card help-card">
        <table>
          <tbody>
            <tr><td><code>hatch setup</code></td><td>One-time setup</td></tr>
            <tr><td><code>hatch add &lt;domain&gt; &lt;port&gt;</code></td><td>Map a domain</td></tr>
            <tr><td><code>hatch rm &lt;domain&gt;</code></td><td>Remove mapping</td></tr>
            <tr><td><code>hatch ls</code></td><td>List all domains</td></tr>
            <tr><td><code>hatch scan</code></td><td>Detect running ports</td></tr>
            <tr><td><code>hatch status</code></td><td>Daemon health</td></tr>
            <tr><td><code>hatch config</code></td><td>Show config</td></tr>
          </tbody>
        </table>
      </div>

      <div className="section-header"><span>Links</span></div>
      <div className="card help-card">
        <p><a href="https://github.com/kingsleyocran/hatch" target="_blank">GitHub Repository</a></p>
        <p><a href="https://github.com/kingsleyocran/hatch/issues" target="_blank">Report an Issue</a></p>
      </div>
    </div>
  );
}

export default App;
