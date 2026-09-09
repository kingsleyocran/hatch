<p align="center">
  <img src="hatch-logo.svg" alt="Hatch" width="80" height="80">
</p>

<h1 align="center">Hatch</h1>

<p align="center">
  <strong>Local domains for your dev servers.</strong><br>
  Map custom domains like <code>myapp.test</code> to your running apps — with HTTPS, zero config, and a single binary.
</p>

<p align="center">
  <a href="https://github.com/kingsleyocran/hatch/releases/latest">Download</a> ·
  <a href="https://marketplace.visualstudio.com/items?itemName=kocranbuild.hatch-vs">VSCode Extension</a> ·
  <a href="https://hatch.kocran.build">Website</a>
</p>

---

## The Problem

Running multiple projects locally means juggling `localhost:3000`, `localhost:3001`, `localhost:8000`. Browser tabs all say "localhost." Cookies collide. OAuth breaks. HTTPS requires self-signed certs with browser warnings.

## The Solution

```bash
hatch add webapp.test 3000 --https
hatch add backend.test 8000 --https
hatch add dashboard.test 5173

hatch ls
  DOMAIN            PORT   STATUS    HTTPS
  webapp.test     3000   ● active  ✓
  backend.test       8000   ● active  ✓
  dashboard.test     5173   ● active
```

Visit `https://webapp.test` — it just works. Green lock. No warnings.

## Features

- **Custom Local Domains** — any `.test` domain to any port
- **Instant HTTPS** — auto-generated trusted certificates
- **Zero Dependencies** — single Go binary, no nginx/Docker/Node
- **Auto-Detection** — watches ports, activates domains when servers start
- **Port Scanner** — detects running services, identifies project types
- **Per-Project Memory** — domains tied to directories, not just ports
- **"Server Stopped" Page** — branded waiting page with auto-refresh
- **WebSocket Proxy** — full HMR support for Next.js, Vite, etc.
- **Cross-Platform** — macOS, Linux, Windows

## Install

### CLI

**macOS (Homebrew)**
```bash
brew install kingsleyocran/tap/hatch
```

**macOS / Linux (direct download)**
```bash
curl -L https://github.com/kingsleyocran/hatch/releases/latest/download/hatch-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') -o hatch
chmod +x hatch && sudo mv hatch /usr/local/bin/
```

**Windows**

Download `hatch-windows-amd64.exe` from [Releases](https://github.com/kingsleyocran/hatch/releases/latest).

**Go Install**
```bash
go install github.com/kingsleyocran/hatch@latest
```

### VSCode Extension

Install [Hatch VSCode](https://marketplace.visualstudio.com/items?itemName=kocranbuild.hatch-vs) from the marketplace, or:

```
ext install kocranbuild.hatch-vs
```

### Desktop App

Download from [Releases](https://github.com/kingsleyocran/hatch/releases/latest):
- **Linux:** `hatch-desktop-linux-amd64`
- **Windows:** `hatch-desktop-windows-amd64.exe`
- **macOS:** Coming soon

## Quick Start

```bash
# One-time setup (configures DNS + certificate authority)
hatch setup

# Map your first domain
hatch add myapp.test 3000 --https

# Start your dev server
npm run dev

# Open in browser
open https://myapp.test
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `hatch setup` | One-time setup (DNS, daemon, CA) |
| `hatch add <domain> <port>` | Map a domain to a port |
| `hatch add <domain> <port> --https` | Map with HTTPS |
| `hatch rm <domain>` | Remove a mapping |
| `hatch ls` | List all domains with status |
| `hatch scan` | Detect running ports and suggest mappings |
| `hatch start` | Start the daemon |
| `hatch stop` | Stop the daemon |
| `hatch status` | Daemon health check |
| `hatch config` | Show configuration |
| `hatch config set <key> <value>` | Update configuration |
| `hatch logs` | View daemon logs |
| `hatch open <domain>` | Open domain in browser |

## How It Works

```
Browser → DNS Resolver (port 15353) → 127.0.0.1
                                        ↓
Browser → Hatch Proxy (port 80/443) → localhost:3000
                                     → localhost:8000
                                     → localhost:5173
```

1. **DNS Resolver** — lightweight DNS server resolves `.test` domains to `127.0.0.1`
2. **Reverse Proxy** — routes requests by domain name, terminates TLS
3. **Port Watcher** — detects when servers start/stop, updates domain status
4. **Certificate Authority** — generates trusted local certs, installs in system trust store

## Configuration

```yaml
# ~/.hatch/config.yaml
default_tld: test
auto_https: true
daemon_port: 80
https_port: 443
dns_port: 15353
```

### TLD Options

| TLD | Notes |
|-----|-------|
| `.test` | Default. IETF-reserved, works everywhere |
| `.local` | Familiar but may conflict with mDNS on macOS |
| `.localhost` | IETF-reserved, browsers treat as secure context |
| `.dev` | Google-owned, browsers force HTTPS |
| `.internal` | IETF-reserved for internal use |

## Platform Support

| Feature | macOS | Linux | Windows |
|---------|-------|-------|---------|
| CLI | ✓ | ✓ | ✓ |
| HTTPS | ✓ | ✓ | ✓ |
| DNS | /etc/resolver/ | systemd-resolved | Hosts file |
| Daemon | LaunchDaemon | systemd | Task Scheduler |
| VSCode Extension | ✓ | ✓ | ✓ |
| Desktop App | Coming soon | ✓ | ✓ |

## Tech Stack

- **CLI:** Go, Cobra
- **Proxy:** Go `net/http/httputil` + `nhooyr.io/websocket`
- **DNS:** `github.com/miekg/dns`
- **TLS:** Go `crypto/x509` (no external mkcert)
- **Extension:** TypeScript, VSCode Extension API
- **Desktop:** Wails (Go + React)
- **CI:** GitHub Actions with cross-compilation

## Contributing

Contributions welcome! Please open an issue first to discuss what you'd like to change.

```bash
# Clone
git clone https://github.com/kingsleyocran/hatch.git
cd hatch

# Build
make build

# Clean test from scratch
make clean

# Run
make dev
```

## License

MIT — [Kingsley Ocran](https://github.com/kingsleyocran)
