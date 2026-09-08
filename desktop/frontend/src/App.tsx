import { useState, useEffect } from 'react';
import { GetDomains, GetStatus, AddDomain, RemoveDomain, GetConfig, SetConfig, IsDaemonRunning, StartDaemon, StopDaemon, ScanPorts } from '../wailsjs/go/main/App';
import { BrowserOpenURL } from '../wailsjs/runtime/runtime';
import logoSvg from './assets/logo.svg';
import './App.css';

type Domain = { domain: string; port: number; dir: string; alive: boolean; https: boolean };
type Status = { running: boolean; uptime: string; domain_count: number; active_count: number };
type Config = { default_tld: string; auto_https: boolean; daemon_port: number; https_port: number; dns_port: number };
type PortInfo = { port: number; name: string; process: string; type: string; dir: string };
type View = 'domains' | 'settings' | 'quickstart';

function App() {
  const [view, setView] = useState<View>('domains');
  const [domains, setDomains] = useState<Domain[]>([]);
  const [status, setStatus] = useState<Status | null>(null);
  const [config, setConfigState] = useState<Config | null>(null);
  const [daemonRunning, setDaemonRunning] = useState(false);
  const [ports, setPorts] = useState<PortInfo[]>([]);
  const [showAddModal, setShowAddModal] = useState(false);
  const [prefillPort, setPrefillPort] = useState<number | null>(null);

  const refresh = async () => {
    const running = await IsDaemonRunning();
    setDaemonRunning(running);
    if (running) {
      setDomains((await GetDomains()) || []);
      setStatus(await GetStatus());
      setPorts((await ScanPorts()) || []);
    }
    setConfigState(await GetConfig());
  };

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, 5000);
    return () => clearInterval(interval);
  }, []);

  const [error, setError] = useState('');

  const handleAdd = async (domain: string, port: number, https: boolean) => {
    const err = await AddDomain(domain, port, https);
    if (err) {
      setError(err);
      setTimeout(() => setError(''), 3000);
    }
    setShowAddModal(false);
    refresh();
  };

  const handleRemove = async (domain: string) => {
    await RemoveDomain(domain);
    refresh();
  };

  return (
    <div className="app">
      <div className="drag-bar" />
      {error && <div className="toast-error">{error}</div>}

      <header className="top-bar">
        <div className="top-left">
          <img src={logoSvg} alt="Hatch" width="22" height="22" />
          <span className="app-name">Hatch</span>
        </div>
        <div className="top-right">
          <span className={`daemon-status ${daemonRunning ? 'online' : 'offline'}`}>
            <span className="daemon-dot" />
            {daemonRunning ? 'Running' : 'Offline'}
          </span>
        </div>
      </header>

      <div className="divider" />

      <main className="content">
        {view === 'domains' && (
          <DomainsView
            domains={domains}
            ports={ports}
            config={config}
            daemonRunning={daemonRunning}
            onRemove={handleRemove}
            onRefresh={refresh}
            onShowAdd={() => { setPrefillPort(null); setShowAddModal(true); }}
            onShowAddPort={(port: number) => { setPrefillPort(port); setShowAddModal(true); }}
            onStartDaemon={async () => { await StartDaemon(); setTimeout(refresh, 2000); }}
          />
        )}
        {view === 'settings' && (
          <SettingsView
            config={config}
            status={status}
            daemonRunning={daemonRunning}
            onConfigChange={async (k, v) => { await SetConfig(k, v); refresh(); }}
            onStopDaemon={async () => { await StopDaemon(); refresh(); }}
          />
        )}
        {view === 'quickstart' && <QuickStartView />}
      </main>

      {showAddModal && (
        <AddDomainModal
          defaultTLD={config?.default_tld || 'test'}
          autoHTTPS={config?.auto_https ?? true}
          prefillPort={prefillPort}
          onAdd={(domain, port, https) => { handleAdd(domain, port, https); setPrefillPort(null); }}
          onClose={() => { setShowAddModal(false); setPrefillPort(null); }}
        />
      )}

      <nav className="bottom-nav">
        <button className={`bottom-btn ${view === 'domains' ? 'active' : ''}`} onClick={() => setView('domains')}>
          <svg width="18" height="18" viewBox="0 0 16 16" fill="currentColor"><path d="M0 8a8 8 0 1116 0A8 8 0 010 8zm7.5-6.923c-.67.204-1.335.82-1.887 1.855A7.97 7.97 0 005.145 4H7.5V1.077zM4.09 4a9.267 9.267 0 01.64-1.539 6.7 6.7 0 01.597-.933A7.025 7.025 0 002.255 4H4.09zm-.582 3.5c.03-.877.138-1.718.312-2.5H1.674a6.958 6.958 0 00-.656 2.5h2.49zM4.847 5a12.5 12.5 0 00-.338 2.5H7.5V5H4.847zM8.5 5v2.5h2.99a12.495 12.495 0 00-.337-2.5H8.5zM4.51 8.5a12.5 12.5 0 00.337 2.5H7.5V8.5H4.51zm3.99 0V11h2.653c.187-.765.306-1.608.338-2.5H8.5zM5.145 12c.138.386.295.744.468 1.068.552 1.035 1.218 1.65 1.887 1.855V12H5.145zm.182 2.472a6.696 6.696 0 01-.597-.933A9.268 9.268 0 014.09 12H2.255a7.024 7.024 0 003.072 2.472zM3.82 11a13.652 13.652 0 01-.312-2.5h-2.49c.062.89.291 1.733.656 2.5H3.82zm6.853 3.472A7.024 7.024 0 0013.745 12H11.91a9.27 9.27 0 01-.64 1.539 6.688 6.688 0 01-.597.933zM8.5 12v2.923c.67-.204 1.335-.82 1.887-1.855.173-.324.33-.682.468-1.068H8.5zm3.68-1h2.146c.365-.767.594-1.61.656-2.5h-2.49a13.65 13.65 0 01-.312 2.5zm2.802-3.5a6.959 6.959 0 00-.656-2.5H12.18c.174.782.282 1.623.312 2.5h2.49zM11.27 2.461c.247.464.462.98.64 1.539h1.835a7.024 7.024 0 00-3.072-2.472c.218.284.418.598.597.933zM10.855 4a7.966 7.966 0 00-.468-1.068C9.835 1.897 9.17 1.282 8.5 1.077V4h2.355z"/></svg>
          <span>Domains</span>
        </button>
        <button className={`bottom-btn ${view === 'settings' ? 'active' : ''}`} onClick={() => setView('settings')}>
          <svg width="18" height="18" viewBox="0 0 16 16" fill="currentColor"><path d="M8 4.754a3.246 3.246 0 100 6.492 3.246 3.246 0 000-6.492zM5.754 8a2.246 2.246 0 114.492 0 2.246 2.246 0 01-4.492 0z"/><path d="M9.796 1.343c-.527-1.79-3.065-1.79-3.592 0l-.094.319a.873.873 0 01-1.255.52l-.292-.16c-1.64-.892-3.433.902-2.54 2.541l.159.292a.873.873 0 01-.52 1.255l-.319.094c-1.79.527-1.79 3.065 0 3.592l.319.094a.873.873 0 01.52 1.255l-.16.292c-.892 1.64.901 3.434 2.541 2.54l.292-.159a.873.873 0 011.255.52l.094.319c.527 1.79 3.065 1.79 3.592 0l.094-.319a.873.873 0 011.255-.52l.292.16c1.64.893 3.434-.902 2.54-2.541l-.159-.292a.873.873 0 01.52-1.255l.319-.094c1.79-.527 1.79-3.065 0-3.592l-.319-.094a.873.873 0 01-.52-1.255l.16-.292c.893-1.64-.902-3.433-2.541-2.54l-.292.159a.873.873 0 01-1.255-.52l-.094-.319zm-2.633.283c.246-.835 1.428-.835 1.674 0l.094.319a1.873 1.873 0 002.693 1.115l.291-.16c.764-.415 1.6.42 1.184 1.185l-.159.292a1.873 1.873 0 001.116 2.692l.318.094c.835.246.835 1.428 0 1.674l-.319.094a1.873 1.873 0 00-1.115 2.693l.16.291c.415.764-.42 1.6-1.185 1.184l-.291-.159a1.873 1.873 0 00-2.693 1.116l-.094.318c-.246.835-1.428.835-1.674 0l-.094-.319a1.873 1.873 0 00-2.692-1.115l-.292.16c-.764.415-1.6-.42-1.184-1.185l.159-.291A1.873 1.873 0 001.945 8.93l-.319-.094c-.835-.246-.835-1.428 0-1.674l.319-.094A1.873 1.873 0 003.06 4.377l-.16-.292c-.415-.764.42-1.6 1.185-1.184l.292.159a1.873 1.873 0 002.692-1.115l.094-.319z"/></svg>
          <span>Settings</span>
        </button>
        <button className={`bottom-btn ${view === 'quickstart' ? 'active' : ''}`} onClick={() => setView('quickstart')}>
          <svg width="18" height="18" viewBox="0 0 16 16" fill="currentColor"><path d="M8 15A7 7 0 118 1a7 7 0 010 14zm0 1A8 8 0 108 0a8 8 0 000 16z"/><path d="M5.255 5.786a.237.237 0 00.241.247h.825c.138 0 .248-.113.266-.25.09-.656.54-1.134 1.342-1.134.686 0 1.314.343 1.314 1.168 0 .635-.374.927-.965 1.371-.673.489-1.206 1.06-1.168 1.987l.003.217a.25.25 0 00.25.246h.811a.25.25 0 00.25-.25v-.105c0-.718.273-.927 1.01-1.486.609-.463 1.244-.977 1.244-2.056 0-1.511-1.276-2.241-2.673-2.241-1.267 0-2.655.59-2.75 2.286zm1.557 5.763c0 .533.425.927 1.01.927.609 0 1.028-.394 1.028-.927 0-.552-.42-.94-1.029-.94-.584 0-1.009.388-1.009.94z"/></svg>
          <span>Quick Start</span>
        </button>
      </nav>
    </div>
  );
}

function DomainsView({ domains, ports, config, daemonRunning, onRemove, onRefresh, onShowAdd, onShowAddPort, onStartDaemon }: {
  domains: Domain[]; ports: PortInfo[]; config: Config | null; daemonRunning: boolean;
  onRemove: (d: string) => void; onRefresh: () => void; onShowAdd: () => void; onShowAddPort: (port: number) => void; onStartDaemon: () => void;
}) {
  const [tab, setTab] = useState<'mapped' | 'ports'>('mapped');

  if (!daemonRunning) {
    return (
      <div className="center-content">
        <svg viewBox="0 0 117.1 107.32" width="48" height="44"><path fill="#B3D7FC" d="M83.92 84.85l5.77 0c3.45,0 5.94,2.82 5.54,6.27l-0.58 5.01c-0.72,6.18 3.71,11.19 9.88,11.19 6.18,0 11.77,-5.01 12.49,-11.19 0.72,-6.18 -3.71,-11.19 -9.88,-11.19l-5.12 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.59 -5.04 0.11 -0.93 4.54 -39.02c0.71,-6.14 -4.11,-11.17 -10.74,-11.23 -0.13,0.01 -0.26,0.01 -0.39,0.01l-4.49 0 -5.55 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.58 -5.01c0.72,-6.18 -3.71,-11.19 -9.88,-11.19 -6.18,0 -11.77,5.01 -12.49,11.19 -0.72,6.18 3.71,11.19 9.88,11.19l5.12 0c3.45,0 5.94,2.82 5.54,6.27l-0.59 5.04c-0,0 -0,0.01 -0,0.01l-0.11 0.92 -1.88 16.2c-0.64,-5.61 -5.17,-9.79 -11.17,-9.79 -5.99,0 -11.48,4.16 -13.44,9.75l1.99 -17.09c0.71,-6.14 -4.11,-11.17 -10.74,-11.23 -0.13,0.01 -0.26,0.01 -0.39,0.01l-4.49 0 -5.55 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.58 -5.01c0.72,-6.18 -3.71,-11.19 -9.88,-11.19 -6.18,0 -11.77,5.01 -12.49,11.19 -0.72,6.18 3.71,11.19 9.88,11.19l5.12 0c3.45,0 5.94,2.82 5.54,6.27l-0.59 5.04c-0,0 -0,0.01 -0,0.01l-0.11 0.92 -4.54 39.01c-0.72,6.18 4.17,11.23 10.86,11.23l0.04 0c0.08,-0 0.15,-0 0.23,-0l4.26 0 0.01 -0 5.77 0c3.45,0 5.94,2.82 5.54,6.27l-0.58 5.01c-0.72,6.18 3.71,11.19 9.88,11.19 6.18,0 11.77,-5.01 12.49,-11.19 0.72,-6.18 -3.71,-11.19 -9.88,-11.19l-5.12 0c-3.45,0 -5.94,-2.82 -5.54,-6.27l0.59 -5.04c0,-0.01 0,-0.01 0,-0.02l0.11 -0.92 1.88 -16.18c0.65,5.59 5.18,9.75 11.17,9.75 6,0 11.5,-4.18 13.45,-9.78l-1.99 17.12c-0.72,6.18 4.17,11.23 10.86,11.23l0.04 0c0.08,-0 0.15,-0 0.23,-0l4.26 0 0.01 -0z"/></svg>
        <h2>Daemon Offline</h2>
        <p className="muted">Start the daemon to manage your local domains.</p>
        <button className="start-btn" onClick={onStartDaemon}>Start Daemon</button>
      </div>
    );
  }

  return (
    <div className="domains-view">
      <div className="tabs">
        <button className={`tab ${tab === 'mapped' ? 'active' : ''}`} onClick={() => setTab('mapped')}>Mapped Domains</button>
        <button className={`tab ${tab === 'ports' ? 'active' : ''}`} onClick={() => setTab('ports')}>Running Ports</button>
        <div className="tab-spacer" />
        <button className="icon-btn" onClick={onShowAdd} title="Add domain">+</button>
        <button className="icon-btn" onClick={onRefresh} title="Refresh">↻</button>
      </div>

      {tab === 'mapped' && (
        <div className="card-list">
          {domains.length === 0 ? (
            <div className="empty-state">
              <p>No domains mapped yet.</p>
              <button className="add-btn" onClick={onShowAdd}>+ Add your first domain</button>
            </div>
          ) : (
            domains.map(d => (
              <div className="card" key={d.domain}>
                <div className="card-row">
                  <span className={`dot ${d.alive ? 'dot-green' : 'dot-red'}`} />
                  <div className="card-info">
                    <span className="card-title">{d.domain}</span>
                    <span className="card-sub">→ localhost:{d.port}</span>
                  </div>
                  <div className="card-badges">
                    <span className={`status-label ${d.alive ? 'sl-active' : 'sl-stopped'}`}>{d.alive ? 'Active' : 'Stopped'}</span>
                    {d.https && <span className="https-badge">HTTPS</span>}
                  </div>
                  <div className="card-actions">
                    <button className="act-btn" onClick={() => BrowserOpenURL(`${d.https ? 'https' : 'http'}://${d.domain}`)} title="Open">↗</button>
                    <button className="act-btn act-danger" onClick={() => onRemove(d.domain)} title="Remove">✕</button>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {tab === 'ports' && (
        <div className="card-list">
          {ports.length === 0 ? (
            <div className="empty-state">
              <p>No unmapped ports detected.</p>
            </div>
          ) : (
            ports.map(p => (
              <div className="card" key={p.port}>
                <div className="card-row">
                  <span className="dot dot-amber" />
                  <div className="card-info">
                    <span className="card-title">:{p.port} {p.name}</span>
                    <span className="card-sub">{[p.process, p.type].filter(Boolean).join(' · ')}</span>
                    {p.dir && <span className="card-dir">{p.dir}</span>}
                  </div>
                  <div className="card-actions always-show">
                    <button className="act-btn act-add" onClick={() => onShowAddPort(p.port)} title="Map to domain">+</button>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}

function AddDomainModal({ defaultTLD, autoHTTPS, prefillPort, onAdd, onClose }: {
  defaultTLD: string; autoHTTPS: boolean; prefillPort: number | null;
  onAdd: (domain: string, port: number, https: boolean) => void; onClose: () => void;
}) {
  const [name, setName] = useState('');
  const [tld, setTld] = useState(defaultTLD);
  const [port, setPort] = useState(prefillPort ? String(prefillPort) : '3000');
  const [https, setHttps] = useState(autoHTTPS);
  const [tldOpen, setTldOpen] = useState(false);

  const tlds = ['test', 'local', 'localhost', 'dev', 'internal'];
  const domain = name.includes('.') ? name : `${name}.${tld}`;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <h3>{prefillPort ? `Map Port ${prefillPort}` : 'Add Domain'}</h3>
        <div className="form-group">
          <label>Domain name</label>
          <div className="domain-input-row">
            <input type="text" placeholder="myapp" value={name} onChange={e => setName(e.target.value)} autoFocus className="domain-name-input" />
            <div className="tld-picker">
              <button className="tld-btn" onClick={() => setTldOpen(!tldOpen)}>.{tld}</button>
              {tldOpen && (
                <div className="tld-dropdown">
                  {tlds.map(t => (
                    <button key={t} className={`tld-option ${t === tld ? 'active' : ''}`} onClick={() => { setTld(t); setTldOpen(false); }}>.{t}</button>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
        {!prefillPort && (
          <div className="form-group">
            <label>Port</label>
            <input type="number" placeholder="3000" value={port} onChange={e => setPort(e.target.value)} />
          </div>
        )}
        <div className="form-row">
          <span>HTTPS</span>
          <label className="toggle">
            <input type="checkbox" checked={https} onChange={e => setHttps(e.target.checked)} />
            <span className="toggle-slider" />
          </label>
        </div>
        <div className="modal-actions">
          <button className="cancel-btn" onClick={onClose}>Cancel</button>
          <button className="primary-btn" onClick={() => name && onAdd(domain, parseInt(port), https)} disabled={!name}>Add Domain</button>
        </div>
      </div>
    </div>
  );
}

function SettingsView({ config, status, daemonRunning, onConfigChange, onStopDaemon }: {
  config: Config | null; status: Status | null; daemonRunning: boolean;
  onConfigChange: (k: string, v: string) => void; onStopDaemon: () => void;
}) {
  return (
    <div className="settings-view">
      <div className="section-label">General</div>
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

      <div className="section-label">Daemon</div>
      <div className="card">
        <div className="setting-row">
          <span>Status</span>
          <span className={daemonRunning ? 'text-green' : 'text-red'}>
            {daemonRunning ? `● Running (${status?.uptime || ''})` : '○ Offline'}
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
        {daemonRunning && <button className="danger-btn" onClick={onStopDaemon}>Stop Daemon</button>}
      </div>
    </div>
  );
}

function QuickStartView() {
  return (
    <div className="quickstart-view">
      <div className="section-label">Quick Start</div>
      <div className="card">
        <p>Hatch maps custom local domains to your running dev servers.</p>
        <pre>{`# Add a domain\nhatch add myapp.test 3000 --https\n\n# List domains\nhatch ls\n\n# Scan running ports\nhatch scan`}</pre>
      </div>

      <div className="section-label">CLI Commands</div>
      <div className="card">
        <table>
          <tbody>
            <tr><td><code>hatch setup</code></td><td>One-time setup</td></tr>
            <tr><td><code>hatch add</code></td><td>Map a domain</td></tr>
            <tr><td><code>hatch rm</code></td><td>Remove mapping</td></tr>
            <tr><td><code>hatch ls</code></td><td>List all domains</td></tr>
            <tr><td><code>hatch scan</code></td><td>Detect running ports</td></tr>
            <tr><td><code>hatch status</code></td><td>Daemon health</td></tr>
          </tbody>
        </table>
      </div>

      <div className="section-label">Links</div>
      <div className="card">
        <p><a href="https://github.com/kingsleyocran/hatch" target="_blank" rel="noreferrer">GitHub Repository</a></p>
        <p><a href="https://github.com/kingsleyocran/hatch#readme" target="_blank" rel="noreferrer">Documentation</a></p>
        <p><a href="https://github.com/kingsleyocran/hatch/issues" target="_blank" rel="noreferrer">Report an Issue</a></p>
      </div>
    </div>
  );
}

export default App;
