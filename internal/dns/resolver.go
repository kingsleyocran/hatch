package dns

import (
	"net"
	"strings"
	"sync"

	mdns "github.com/miekg/dns"
)

type Resolver struct {
	mu      sync.RWMutex
	domains map[string]bool
	server  *mdns.Server
}

func New() *Resolver {
	return &Resolver{
		domains: make(map[string]bool),
	}
}

func (r *Resolver) AddDomain(domain string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.domains[mdns.Fqdn(strings.ToLower(domain))] = true
}

func (r *Resolver) RemoveDomain(domain string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.domains, mdns.Fqdn(strings.ToLower(domain)))
}

func (r *Resolver) HasDomain(domain string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.domains[mdns.Fqdn(strings.ToLower(domain))]
}

func (r *Resolver) Start(addr string) error {
	mux := mdns.NewServeMux()
	mux.HandleFunc(".", r.handleQuery)

	r.server = &mdns.Server{
		Addr:    addr,
		Net:     "udp",
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- r.server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func (r *Resolver) Stop() error {
	if r.server != nil {
		return r.server.Shutdown()
	}
	return nil
}

func (r *Resolver) handleQuery(w mdns.ResponseWriter, req *mdns.Msg) {
	msg := new(mdns.Msg)
	msg.SetReply(req)
	msg.Authoritative = true

	for _, q := range req.Question {
		if q.Qtype == mdns.TypeA {
			r.mu.RLock()
			_, found := r.domains[strings.ToLower(q.Name)]
			r.mu.RUnlock()

			if found {
				msg.Answer = append(msg.Answer, &mdns.A{
					Hdr: mdns.RR_Header{
						Name:   q.Name,
						Rrtype: mdns.TypeA,
						Class:  mdns.ClassINET,
						Ttl:    60,
					},
					A: net.IPv4(127, 0, 0, 1),
				})
			}
		}
	}

	if len(msg.Answer) == 0 {
		msg.Rcode = mdns.RcodeNameError
	}

	w.WriteMsg(msg)
}
