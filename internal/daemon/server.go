package daemon

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/kingsleyocran/hatch/internal/config"
)

type Server struct {
	daemon   *Daemon
	sockPath string
	listener net.Listener
	wg       sync.WaitGroup
	quit     chan struct{}
}

func NewServer(d *Daemon, sockPath string) *Server {
	return &Server{
		daemon:   d,
		sockPath: sockPath,
		quit:     make(chan struct{}),
	}
}

func (s *Server) Start() error {
	var ln net.Listener
	var err error

	if config.IsWindows() {
		ln, err = net.Listen("tcp", s.sockPath)
	} else {
		os.Remove(s.sockPath)
		if mkErr := os.MkdirAll(filepath.Dir(s.sockPath), 0755); mkErr != nil {
			return mkErr
		}
		ln, err = net.Listen("unix", s.sockPath)
		if err == nil {
			os.Chmod(s.sockPath, 0666)
		}
	}
	if err != nil {
		return err
	}
	s.listener = ln

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-s.quit:
					return
				default:
					continue
				}
			}
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				s.handleConn(conn)
			}()
		}
	}()

	return nil
}

func (s *Server) Stop() {
	close(s.quit)
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
	os.Remove(s.sockPath)
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	var req Request
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		resp := Response{OK: false, Message: "invalid request"}
		json.NewEncoder(conn).Encode(resp)
		return
	}

	resp := s.daemon.handleRequest(req)
	json.NewEncoder(conn).Encode(resp)
}
