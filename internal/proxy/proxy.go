package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type route struct {
	domain  string
	port    int
	alive   bool
	https   bool
	proxy   *httputil.ReverseProxy
	stopped http.Handler
}

type Manager struct {
	mu        sync.RWMutex
	routes    map[string]*route
	ports     map[int]bool
	server    *http.Server
	tlsServer *http.Server
	ws        *WSHub
}

func New() *Manager {
	return &Manager{
		routes: make(map[string]*route),
		ports:  make(map[int]bool),
		ws:     NewWSHub(),
	}
}

func (m *Manager) AddRouteHTTPS(domain string, port int, https bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	target, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
	rp := httputil.NewSingleHostReverseProxy(target)
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		m.mu.RLock()
		rt := m.routes[domain]
		m.mu.RUnlock()
		if rt != nil {
			rt.stopped.ServeHTTP(w, r)
		}
	}

	m.routes[domain] = &route{
		domain:  domain,
		port:    port,
		alive:   m.ports[port],
		https:   https,
		proxy:   rp,
		stopped: NewStoppedHandler(domain, port),
	}
}

func (m *Manager) RemoveRoute(domain string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.routes, domain)
}

func (m *Manager) Routes() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var domains []string
	for d := range m.routes {
		domains = append(domains, d)
	}
	return domains
}

func (m *Manager) SetPortStatus(port int, alive bool) {
	m.mu.Lock()
	wasAlive := m.ports[port]
	m.ports[port] = alive
	for _, r := range m.routes {
		if r.port == port {
			r.alive = alive
		}
	}
	m.mu.Unlock()

	if alive && !wasAlive {
		m.ws.Broadcast("ready")
	}
}

func (m *Manager) StoppedHandler(domain string, port int) http.Handler {
	return NewStoppedHandler(domain, port)
}

func (m *Manager) Start(addr string) error {
	m.server = &http.Server{
		Addr:    addr,
		Handler: m,
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go m.server.Serve(ln)
	return nil
}

func (m *Manager) Stop() error {
	if m.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return m.server.Shutdown(ctx)
	}
	return nil
}

func (m *Manager) StartTLS(addr string, getCert func(*tls.ClientHelloInfo) (*tls.Certificate, error)) error {
	m.tlsServer = &http.Server{
		Addr:    addr,
		Handler: m,
		TLSConfig: &tls.Config{
			GetCertificate: getCert,
		},
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	tlsLn := tls.NewListener(ln, m.tlsServer.TLSConfig)
	go m.tlsServer.Serve(tlsLn)
	return nil
}

func (m *Manager) StopTLS() error {
	if m.tlsServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return m.tlsServer.Shutdown(ctx)
	}
	return nil
}

func (m *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/__hatch/ws" {
		m.ws.ServeWS(w, r)
		return
	}

	m.mu.RLock()
	rt := m.routes[r.Host]
	m.mu.RUnlock()

	if rt == nil {
		http.NotFound(w, r)
		return
	}

	if rt.https && r.TLS == nil {
		target := "https://" + r.Host + r.URL.RequestURI()
		http.Redirect(w, r, target, http.StatusMovedPermanently)
		return
	}

	if !rt.alive {
		rt.stopped.ServeHTTP(w, r)
		return
	}

	rt.proxy.ServeHTTP(w, r)
}
