package watcher

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type portState struct {
	alive    bool
	reported bool
}

type Watcher struct {
	mu       sync.RWMutex
	ports    map[int]*portState
	interval time.Duration
	onChange func(port int, alive bool)
	quit     chan struct{}
	wg       sync.WaitGroup
}

func New(interval time.Duration, onChange func(port int, alive bool)) *Watcher {
	return &Watcher{
		ports:    make(map[int]*portState),
		interval: interval,
		onChange: onChange,
		quit:     make(chan struct{}),
	}
}

func (w *Watcher) Watch(port int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.ports[port] = &portState{}
}

func (w *Watcher) Unwatch(port int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.ports, port)
}

func (w *Watcher) WatchedPorts() []int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	var ports []int
	for p := range w.ports {
		ports = append(ports, p)
	}
	return ports
}

func (w *Watcher) IsAlive(port int) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if s, ok := w.ports[port]; ok {
		return s.alive
	}
	return false
}

func (w *Watcher) Start() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		w.poll()

		for {
			select {
			case <-ticker.C:
				w.poll()
			case <-w.quit:
				return
			}
		}
	}()
}

func (w *Watcher) Stop() {
	close(w.quit)
	w.wg.Wait()
}

func (w *Watcher) poll() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for port, state := range w.ports {
		alive := checkPort(port)
		if alive != state.alive || !state.reported {
			state.alive = alive
			state.reported = true
			if w.onChange != nil {
				w.onChange(port, alive)
			}
		}
	}
}

func checkPort(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 200*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
