# Hatch — Website Content

## Hero Section

### Headline
**Local domains for your dev servers.**

### Subheadline
Stop juggling `localhost:3000`. Map custom domains like `myapp.test` to your running apps — with HTTPS, zero config, and a single binary.

### CTA
- **Download** (primary)
- **View on GitHub** (secondary)

### Hero Visual
Terminal showing:
```
$ hatch add webapp.test 3000 --https
✓ webapp.test → localhost:3000 (https)

$ hatch ls
  DOMAIN            PORT   STATUS    HTTPS
  webapp.test     3000   ● active  ✓
  backend.test       8000   ● active  ✓
  dashboard.test     5173   ● active  ✓
```

---

## Problem Section

### Headline
**Every developer deals with this.**

### Pain points (3 cards)

**Port confusion**
You're running three projects. Which one is `localhost:3000`? Which is `localhost:3001`? Your browser tabs all say "localhost."

**Cookie collisions**
Different projects on localhost share cookies. Auth tokens leak between apps. Sessions break when you switch projects.

**HTTPS headaches**
OAuth requires HTTPS. Secure cookies need HTTPS. Testing SSL locally means self-signed certs, browser warnings, and wasted time.

---

## Solution Section

### Headline
**Hatch fixes all of this.**

### How it works (3 steps)

**1. Install**
Single binary. No dependencies. One command.
```
brew install kingsleyocran/tap/hatch
```

**2. Setup**
One-time setup configures DNS and installs a trusted certificate authority. Your local domains resolve instantly — no `/etc/hosts` editing.
```
hatch setup
```

**3. Map**
Point any `.test` domain to any port. Add `--https` for a trusted green lock.
```
hatch add myapp.test 3000 --https
```

Visit `https://myapp.test` — it just works.

---

## Features Section

### Core Features (grid layout)

**Custom Local Domains**
Map `webapp.test`, `backend.test`, `api.test` — any name to any port. Your browser tabs finally make sense.

**Instant HTTPS**
Generates trusted TLS certificates automatically. Green lock in your browser. No more self-signed cert warnings.

**Zero Dependencies**
Single Go binary. No nginx, no Docker, no Node. Download and run.

**Auto-Detection**
Hatch watches your ports. Start `npm run dev` on port 3000 — your domain goes live automatically. Stop the server — a branded "waiting" page appears with auto-refresh.

**Per-Project Memory**
Domains are tied to project directories, not just ports. Two projects on port 3000? Hatch knows which is which.

**Port Scanner**
`hatch scan` detects all running services, identifies project types (Node, Go, Rust, Python), and suggests domain mappings.

**VSCode Extension**
Manage domains from your editor. Sidebar shows all mapped domains and running ports. Add domains with one click.

**Cross-Platform**
macOS, Linux, and Windows. Same binary, same commands, same experience.

---

## VSCode Extension Section

### Headline
**Manage domains from your editor.**

### Description
The Hatch VSCode extension gives you a visual dashboard for all your local domains. See what's running, add new mappings, and open domains in your browser — without leaving your editor.

### Features
- Sidebar panel with domain status (active/stopped)
- Running port detection with project names
- One-click domain mapping with HTTPS
- Auto-downloads Hatch CLI if not installed
- System password dialog for one-time setup
- Works on macOS, Linux, and Windows

### CTA
**Install from VS Code Marketplace** →

---

## How It Works (Technical)

### Headline
**What happens under the hood.**

### Architecture diagram description

```
Browser → DNS Resolver → 127.0.0.1
                           ↓
Browser → Hatch Proxy (port 80/443) → localhost:3000
                                    → localhost:8000
                                    → localhost:5173
```

**DNS Resolver**
Hatch runs a lightweight DNS server. Your OS is configured to ask Hatch for `.test` domains. Every registered domain resolves to `127.0.0.1` — instantly, no `/etc/hosts` editing.

**Reverse Proxy**
Hatch is a Go reverse proxy that routes requests by domain name. `webapp.test` goes to port 3000. `backend.test` goes to port 8000. TLS termination happens at the proxy — your dev servers stay on plain HTTP.

**Port Watcher**
Hatch monitors registered ports. When your dev server starts, the domain goes live. When it stops, visitors see a branded "waiting" page that auto-refreshes via WebSocket when the server comes back.

**Certificate Authority**
On first HTTPS domain, Hatch generates a local CA and installs it in your system trust store. Domain certificates are generated on the fly. Zero browser warnings.

---

## CLI Reference

### Commands

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

### TLD Options

| TLD | Notes |
|-----|-------|
| `.test` | Default. IETF-reserved, works everywhere |
| `.local` | Familiar but may conflict with mDNS on macOS |
| `.localhost` | IETF-reserved, browsers treat as secure context |
| `.dev` | Google-owned, browsers force HTTPS |
| `.internal` | IETF-reserved for internal use |

```
hatch config set default_tld test
```

---

## Comparison Section

### Headline
**How Hatch compares.**

| Feature | Hatch | Laravel Valet | Hotel | DotLocal |
|---------|-------|---------------|-------|----------|
| Language agnostic | ✓ | PHP only | ✓ | ✓ |
| Single binary | ✓ | ✗ (PHP) | ✗ (Node) | varies |
| No external deps | ✓ | nginx | ✗ | varies |
| HTTPS | ✓ (auto) | ✓ | ✗ | varies |
| VSCode extension | ✓ | ✗ | ✗ | ✗ |
| Cross-platform | ✓ | macOS only | ✓ | varies |
| Port scanning | ✓ | ✗ | ✓ | ✗ |
| Auto-detect projects | ✓ | ✗ | ✗ | ✗ |

---

## Installation Section

### macOS

```bash
# Homebrew
brew install kingsleyocran/tap/hatch

# Or download directly
curl -L https://github.com/kingsleyocran/hatch/releases/latest/download/hatch-darwin-arm64 -o hatch
chmod +x hatch && sudo mv hatch /usr/local/bin/
```

### Linux

```bash
curl -L https://github.com/kingsleyocran/hatch/releases/latest/download/hatch-linux-amd64 -o hatch
chmod +x hatch && sudo mv hatch /usr/local/bin/
```

### Windows

Download `hatch-windows-amd64.exe` from [GitHub Releases](https://github.com/kingsleyocran/hatch/releases/latest).

### Go Install

```bash
go install github.com/kingsleyocran/hatch@latest
```

### VSCode Extension

Search "Hatch" in the VS Code Extensions marketplace, or:
```
ext install kingsleyocran.hatch
```

---

## Desktop App Download Page

### Headline
**Download Hatch Desktop**

### Subheadline
Manage your local domains visually. Dashboard, settings, port scanning — all in one app.

### Download Cards

**Windows**
Download `hatch-desktop-windows-amd64.exe` from GitHub Releases.
Button: **Download for Windows** →

**Linux**
Download `hatch-desktop-linux-amd64` from GitHub Releases.
Button: **Download for Linux** →

**macOS**
Coming soon. macOS desktop app requires Apple Developer Program signing for a smooth install experience. Use the CLI or VSCode extension in the meantime.
Button: **Use CLI Instead** → (links to installation section)

### Important Notice (shown after clicking download)

**Before you run the app:**

The Hatch desktop app is open source but not yet code-signed. Your operating system may show a security warning when you first run it. This is normal for open-source software.

**Windows:**
1. Download `hatch-desktop-windows-amd64.exe`
2. Windows SmartScreen may say "Windows protected your PC"
3. Click **"More info"** then **"Run anyway"**
4. This only happens once — subsequent launches work normally

**Linux:**
1. Download `hatch-desktop-linux-amd64`
2. Make it executable: `chmod +x hatch-desktop-linux-amd64`
3. Run: `./hatch-desktop-linux-amd64`
4. No security warnings on Linux

**Why isn't it signed?**
Code signing certificates cost $200-400/year. As an open-source project, we prioritize features over certificates. The app is fully open source — you can inspect every line of code on [GitHub](https://github.com/kingsleyocran/hatch).

---

## Quick Start Section

### Headline
**Up and running in 60 seconds.**

```bash
# Install
brew install kingsleyocran/tap/hatch

# One-time setup (configures DNS + certificate authority)
hatch setup

# Map your first domain
hatch add myapp.test 3000 --https

# Start your dev server
npm run dev

# Open in browser
open https://myapp.test
```

---

## Use Cases Section

### Headline
**Built for how developers actually work.**

**Multiple projects**
Running Cayacart, Phamel, and Orborbit simultaneously? Each gets its own domain. No more guessing which port is which.

**OAuth & SSO**
OAuth callbacks require consistent URLs. `https://myapp.test/auth/callback` works every time — no port juggling.

**Microservices**
Frontend on `app.test`, API on `api.test`, admin on `admin.test`. Each resolves to the right service.

**Team consistency**
Share `.hatch.yaml` configs. Everyone on the team uses the same domains. No more "it works on my machine."

**Docker & containers**
Works alongside Docker. Map container ports to readable domains.

---

## Open Source Section

### Headline
**Open source. MIT licensed.**

Hatch is free and open source. Built with Go for the CLI and TypeScript for the VSCode extension. Contributions welcome.

- **GitHub:** github.com/kingsleyocran/hatch
- **License:** MIT
- **Author:** Kingsley Ocran

### Tech Stack
- **CLI:** Go, Cobra
- **Proxy:** Go `net/http/httputil` + `nhooyr.io/websocket`
- **DNS:** `github.com/miekg/dns`
- **TLS:** Go `crypto/x509` (no external mkcert)
- **Extension:** TypeScript, VSCode Extension API
- **CI:** GitHub Actions with cross-compilation

---

## Footer

**Hatch** — Local domains for your dev servers.

Built by [Kingsley Ocran](https://github.com/kingsleyocran).

[GitHub](https://github.com/kingsleyocran/hatch) · [VSCode Extension](#) · [Documentation](#)

---

## SEO & Meta

**Title:** Hatch — Local Domains for Your Dev Servers

**Description:** Map custom local domains like myapp.test to your running development servers. HTTPS, zero config, single binary. Free and open source.

**Keywords:** local development, dev server, localhost, custom domains, HTTPS local, reverse proxy, developer tools, VSCode extension

**OG Image:** Hatch logo on dark navy background with tagline "Local domains for your dev servers."

---

## Brand Guidelines

**Logo:** Stylized "H" made of connected rounded shapes — represents routing/connections

**Colors:**
- Primary: `#B3D7FC` (light blue — logo color)
- Background: `#0A1628` (dark navy)
- Accent: `#3FB950` (green — active status)
- Error: `#F85149` (red — stopped status)
- Amber: `#D29922` (unmapped ports)

**Fonts:**
- Headings: Inter or system font stack
- Code: JetBrains Mono or system monospace

**Tone:** Direct, technical, no fluff. Speak to developers who value tools that work without ceremony.
