package daemon

import (
	"fmt"
	"sync"
	"time"

	"github.com/kingsleyocran/hatch/internal/config"
	hdns "github.com/kingsleyocran/hatch/internal/dns"
	"github.com/kingsleyocran/hatch/internal/project"
	"github.com/kingsleyocran/hatch/internal/proxy"
	"github.com/kingsleyocran/hatch/internal/watcher"
)

type Daemon struct {
	cfg      *config.Config
	store    *project.Store
	sockPath string

	resolver *hdns.Resolver
	proxy    *proxy.Manager
	watcher  *watcher.Watcher
	server   *Server

	mu      sync.Mutex
	running bool
}

func NewDaemon(cfg *config.Config, store *project.Store, sockPath string) *Daemon {
	d := &Daemon{
		cfg:      cfg,
		store:    store,
		sockPath: sockPath,
		resolver: hdns.New(),
		proxy:    proxy.New(),
	}

	d.watcher = watcher.New(2*time.Second, func(port int, alive bool) {
		d.proxy.SetPortStatus(port, alive)
	})

	return d
}

func (d *Daemon) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return fmt.Errorf("daemon already running")
	}

	dnsAddr := fmt.Sprintf("127.0.0.1:%d", d.cfg.DNSPort)
	if err := d.resolver.Start(dnsAddr); err != nil {
		return fmt.Errorf("start DNS resolver: %w", err)
	}

	proxyAddr := fmt.Sprintf("127.0.0.1:%d", d.cfg.DaemonPort)
	if err := d.proxy.Start(proxyAddr); err != nil {
		d.resolver.Stop()
		return fmt.Errorf("start proxy: %w", err)
	}

	for _, p := range d.store.List() {
		d.resolver.AddDomain(p.Domain)
		d.proxy.AddRoute(p.Domain, p.Port)
		d.watcher.Watch(p.Port)
	}

	d.watcher.Start()

	d.server = NewServer(d, d.sockPath)
	if err := d.server.Start(); err != nil {
		d.watcher.Stop()
		d.proxy.Stop()
		d.resolver.Stop()
		return fmt.Errorf("start API server: %w", err)
	}

	d.running = true
	return nil
}

func (d *Daemon) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return nil
	}

	d.server.Stop()
	d.watcher.Stop()
	d.proxy.Stop()
	d.resolver.Stop()

	d.running = false
	return nil
}

func (d *Daemon) Running() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.running
}

func (d *Daemon) handleRequest(req Request) Response {
	switch req.Action {
	case ActionPing:
		return Response{OK: true, Message: "pong"}

	case ActionAdd:
		p := project.Project{
			Dir:     req.Dir,
			Domain:  req.Domain,
			Port:    req.Port,
			Created: time.Now(),
		}
		if err := d.store.Add(p); err != nil {
			return Response{OK: false, Message: err.Error()}
		}
		if err := d.store.Save(); err != nil {
			return Response{OK: false, Message: err.Error()}
		}
		d.resolver.AddDomain(req.Domain)
		d.proxy.AddRoute(req.Domain, req.Port)
		d.watcher.Watch(req.Port)
		return Response{OK: true, Message: fmt.Sprintf("mapped %s → localhost:%d", req.Domain, req.Port)}

	case ActionRemove:
		p, err := d.store.FindByDomain(req.Domain)
		if err != nil {
			return Response{OK: false, Message: err.Error()}
		}
		d.resolver.RemoveDomain(req.Domain)
		d.proxy.RemoveRoute(req.Domain)
		d.watcher.Unwatch(p.Port)
		if err := d.store.Remove(req.Domain); err != nil {
			return Response{OK: false, Message: err.Error()}
		}
		d.store.Save()
		return Response{OK: true, Message: fmt.Sprintf("removed %s", req.Domain)}

	case ActionList:
		projects := d.store.List()
		var statuses []ProjectStatus
		for _, p := range projects {
			statuses = append(statuses, ProjectStatus{
				Domain: p.Domain,
				Port:   p.Port,
				Dir:    p.Dir,
				Alive:  d.watcher.IsAlive(p.Port),
				HTTPS:  p.HTTPS,
			})
		}
		return Response{OK: true, Projects: statuses}

	case ActionStop:
		go d.Stop()
		return Response{OK: true, Message: "daemon stopping"}

	default:
		return Response{OK: false, Message: fmt.Sprintf("unknown action: %s", req.Action)}
	}
}
