package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func findFreePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

func TestProxyForwardsToUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from upstream"))
	}))
	defer upstream.Close()

	upstreamPort := upstream.Listener.Addr().(*net.TCPAddr).Port
	proxyPort := findFreePort(t)

	mgr := New()
	mgr.AddRouteHTTPS("cayacart.test", upstreamPort, false)
	mgr.SetPortStatus(upstreamPort, true)

	if err := mgr.Start(fmt.Sprintf("127.0.0.1:%d", proxyPort)); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer mgr.Stop()

	time.Sleep(50 * time.Millisecond)

	req, _ := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/", proxyPort), nil)
	req.Host = "cayacart.test"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "hello from upstream" {
		t.Errorf("body = %q, want %q", string(body), "hello from upstream")
	}
}

func TestProxyServesStoppedPageWhenPortDown(t *testing.T) {
	proxyPort := findFreePort(t)

	mgr := New()
	mgr.AddRouteHTTPS("cayacart.test", 39999, false)
	mgr.SetPortStatus(39999, false)

	if err := mgr.Start(fmt.Sprintf("127.0.0.1:%d", proxyPort)); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer mgr.Stop()

	time.Sleep(50 * time.Millisecond)

	req, _ := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/", proxyPort), nil)
	req.Host = "cayacart.test"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}
}

func TestProxyReturns404ForUnknownDomain(t *testing.T) {
	proxyPort := findFreePort(t)

	mgr := New()
	if err := mgr.Start(fmt.Sprintf("127.0.0.1:%d", proxyPort)); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer mgr.Stop()

	time.Sleep(50 * time.Millisecond)

	req, _ := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/", proxyPort), nil)
	req.Host = "unknown.test"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestProxyRemoveRoute(t *testing.T) {
	mgr := New()
	mgr.AddRouteHTTPS("cayacart.test", 3000, false)
	mgr.RemoveRoute("cayacart.test")

	routes := mgr.Routes()
	if len(routes) != 0 {
		t.Errorf("Routes() = %d, want 0", len(routes))
	}
}
