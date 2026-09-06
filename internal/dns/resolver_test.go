package dns

import (
	"fmt"
	"net"
	"testing"
	"time"

	mdns "github.com/miekg/dns"
)

func findFreePort(t *testing.T) int {
	t.Helper()
	l, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := l.LocalAddr().(*net.UDPAddr).Port
	l.Close()
	return port
}

func queryA(addr, domain string) ([]net.IP, error) {
	c := new(mdns.Client)
	c.Timeout = 2 * time.Second
	m := new(mdns.Msg)
	m.SetQuestion(mdns.Fqdn(domain), mdns.TypeA)

	r, _, err := c.Exchange(m, addr)
	if err != nil {
		return nil, err
	}

	var ips []net.IP
	for _, ans := range r.Answer {
		if a, ok := ans.(*mdns.A); ok {
			ips = append(ips, a.A)
		}
	}
	return ips, nil
}

func TestResolverResolvesRegisteredDomain(t *testing.T) {
	port := findFreePort(t)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	r := New()
	r.AddDomain("cayacart.test")

	if err := r.Start(addr); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer r.Stop()

	time.Sleep(50 * time.Millisecond)

	ips, err := queryA(addr, "cayacart.test")
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if len(ips) != 1 || !ips[0].Equal(net.IPv4(127, 0, 0, 1)) {
		t.Errorf("got IPs %v, want [127.0.0.1]", ips)
	}
}

func TestResolverReturnsNXDOMAINForUnknown(t *testing.T) {
	port := findFreePort(t)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	r := New()
	if err := r.Start(addr); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer r.Stop()

	time.Sleep(50 * time.Millisecond)

	c := new(mdns.Client)
	c.Timeout = 2 * time.Second
	m := new(mdns.Msg)
	m.SetQuestion(mdns.Fqdn("unknown.test"), mdns.TypeA)

	resp, _, err := c.Exchange(m, addr)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if resp.Rcode != mdns.RcodeNameError {
		t.Errorf("Rcode = %d, want NXDOMAIN (%d)", resp.Rcode, mdns.RcodeNameError)
	}
}

func TestResolverAddAndRemoveDomain(t *testing.T) {
	r := New()
	r.AddDomain("cayacart.test")

	if !r.HasDomain("cayacart.test") {
		t.Error("HasDomain() = false after Add, want true")
	}

	r.RemoveDomain("cayacart.test")

	if r.HasDomain("cayacart.test") {
		t.Error("HasDomain() = true after Remove, want false")
	}
}
