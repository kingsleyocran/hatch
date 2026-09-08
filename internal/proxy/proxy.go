package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
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
	rp := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)
			pr.Out.Host = target.Host
		},
		FlushInterval: -1,
	}

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

	m.server.SetKeepAlivesEnabled(false)
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

	if isWebSocketUpgrade(r) {
		proxyWebSocket(w, r, rt.port)
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

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

func proxyWebSocket(w http.ResponseWriter, r *http.Request, port int) {
	upstream, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 5*time.Second)
	if err != nil {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
		return
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		upstream.Close()
		http.Error(w, "hijack not supported", http.StatusInternalServerError)
		return
	}

	client, brw, err := hj.Hijack()
	if err != nil {
		upstream.Close()
		return
	}

	targetHost := fmt.Sprintf("127.0.0.1:%d", port)
	fmt.Fprintf(upstream, "%s %s HTTP/1.1\r\n", r.Method, r.RequestURI)
	fmt.Fprintf(upstream, "Host: %s\r\n", targetHost)
	for k, vs := range r.Header {
		if strings.EqualFold(k, "Host") {
			continue
		}
		for _, v := range vs {
			if strings.EqualFold(k, "Origin") {
				fmt.Fprintf(upstream, "%s: http://%s\r\n", k, targetHost)
			} else {
				fmt.Fprintf(upstream, "%s: %s\r\n", k, v)
			}
		}
	}
	fmt.Fprintf(upstream, "\r\n")

	reader := bufio.NewReader(upstream)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			client.Close()
			upstream.Close()
			return
		}
		brw.WriteString(line)
		if line == "\r\n" {
			brw.Flush()
			break
		}
	}

	done := make(chan bool, 2)
	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				brw.Write(buf[:n])
				brw.Flush()
			}
			if err != nil {
				break
			}
		}
		done <- true
	}()
	go func() {
		if brw.Reader.Buffered() > 0 {
			p, _ := brw.Reader.Peek(brw.Reader.Buffered())
			upstream.Write(p)
		}
		buf := make([]byte, 32*1024)
		for {
			n, err := client.Read(buf)
			if n > 0 {
				upstream.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
		done <- true
	}()
	<-done
	client.Close()
	upstream.Close()
}

