# Hatch — Product Spec

## What It Is

A CLI tool and VSCode extension that maps local domains (e.g. `cayacart.local`, `phamel.local`) to your running apps. No more juggling `localhost:3000` vs `localhost:3001`. No more cookie collisions. No more browser tabs that all say "localhost."

## The Problem

When you're running multiple projects at the same time (Cayacart, Phamel, Orborbit), every app lives on localhost with a different port number. Your browser tabs all say "localhost." Cookies collide between projects. OAuth redirects break because the callback URL is `localhost:3000` but your app is now on `localhost:3001`. You waste time remembering which port is which.

Developers working on microservices or multiple projects hit this daily. The current solutions are either manual `/etc/hosts` editing (tedious, easy to forget) or tools like Laravel Valet that are framework-specific.

## Core Features

### CLI

```
hatch add cayacart.local 3000
```
Maps `cayacart.local` to `localhost:3000`. Done.

```
hatch add phamel.local 8000
hatch add orborbit.local 3001
```

```
hatch ls
```
Shows all active mappings with status (running / stopped).

```
hatch rm cayacart.local
```
Removes the mapping.

### What Happens Under the Hood

1. Updates `/etc/hosts` to point the domain to `127.0.0.1`
2. Generates and manages nginx reverse proxy configs
3. Routes `cayacart.local` traffic to `localhost:3000` through nginx

### Two Modes

**Native mode:**
- nginx runs on the host machine
- Configs written to `/etc/nginx/conf.d/`
- Best for developers who already have nginx installed

**Docker mode:**
- Hatch manages an nginx container
- Mounts generated configs into the container
- No nginx installation required on the host
- Works with existing docker-compose setups

### Docker Compose Integration

```
hatch scan
```
Reads `docker-compose.yml` in the current directory, detects services and their ports, and suggests `.local` domain mappings:

```
Found 3 services:
  cayacart-api (port 3000) → cayacart-api.local? [Y/n]
  cayacart-web (port 3001) → cayacart-web.local? [Y/n]
  postgres (port 5432) → skip (database)
```

### Local HTTPS (Optional)

```
hatch add cayacart.local 3000 --https
```
Integrates with mkcert to generate and install self-signed certificates. Your local domain works over HTTPS with no browser warnings. Useful for testing OAuth flows, secure cookies, and APIs that require HTTPS.

### Additional Commands

```
hatch status          # health check on all mappings
hatch open cayacart   # opens cayacart.local in default browser
hatch stop            # removes all mappings and stops nginx
hatch config          # show/edit config (nginx path, mode, etc.)
```

## VSCode Extension

A sidebar panel that makes Hatch visual:

- **Domain list** — shows all active `.local` domains with status indicators (green = running, red = stopped)
- **Add/remove** — click to add a new mapping or remove an existing one
- **Open in browser** — click a domain to open it
- **Auto-detect** — reads workspace `docker-compose.yml` or common port configs and suggests mappings
- **Status bar** — shows number of active Hatch domains in the VSCode status bar

## Architecture

### Stack
- **CLI:** Go (cobra for commands, single binary, no runtime dependency)
- **VSCode Extension:** TypeScript
- **Communication:** CLI is the source of truth. Extension calls the CLI under the hood.

### Components
1. **Config store** — `~/.hatch/config.yaml` for global settings, `.hatch.yaml` per project for project-specific mappings
2. **Hosts manager** — reads/writes `/etc/hosts` entries (requires sudo on first use, can be configured for passwordless)
3. **Nginx config generator** — templates for server blocks, SSL configs
4. **Docker manager** — manages the nginx container lifecycle in Docker mode
5. **mkcert integration** — generates certs for HTTPS domains
6. **VSCode extension** — TypeScript sidebar panel, communicates with CLI

### File Structure
```
hatch/
  cmd/           — cobra commands (add, rm, ls, scan, status, etc.)
  internal/
    hosts/       — /etc/hosts read/write
    nginx/       — config generation and management
    docker/      — container management for Docker mode
    tls/         — mkcert integration
    config/      — config loading and persistence
    scanner/     — docker-compose.yml parser
  templates/     — nginx config templates
```

## Similar Projects

- **De-Great's DotLocal** — similar concept, study it, differentiate
- **Laravel Valet** — PHP/Laravel specific, macOS only
- **Hotel / Overmind** — process managers with local domain features

### How Hatch Differentiates
- Language/framework agnostic. Works with anything that runs on a port
- Docker-native. First-class docker-compose integration
- VSCode extension. Visual management, not just CLI
- Single binary. No runtime dependencies (Node, Ruby, PHP)
- Cross-platform. macOS and Linux (Windows WSL2 support later)

## MVP Scope

**Week 1 — CLI:**
- `hatch add`, `hatch rm`, `hatch ls`, `hatch open`
- `/etc/hosts` management
- nginx config generation (native mode)
- Basic config file

**Week 2 — Extension + Polish:**
- Docker mode
- `hatch scan` (docker-compose integration)
- HTTPS with mkcert
- VSCode extension (sidebar, status bar)
- README with install instructions and demo GIF

## Success Criteria

- Running `hatch add myapp.local 3000` takes under 5 seconds from install to working domain
- Zero configuration for the common case
- A developer with 3+ projects running can set up all domains in under a minute
