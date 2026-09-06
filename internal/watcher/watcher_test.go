package watcher

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

func TestWatcherDetectsPortUp(t *testing.T) {
	var mu sync.Mutex
	events := make(map[int]bool)

	w := New(100*time.Millisecond, func(port int, alive bool) {
		mu.Lock()
		events[port] = alive
		mu.Unlock()
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	defer ln.Close()

	w.Watch(port)
	w.Start()
	defer w.Stop()

	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	alive, reported := events[port]
	mu.Unlock()

	if !reported || !alive {
		t.Errorf("port %d: reported=%v, alive=%v — want reported=true, alive=true", port, reported, alive)
	}
}

func TestWatcherDetectsPortDown(t *testing.T) {
	var mu sync.Mutex
	var lastAlive bool
	var callCount int

	w := New(100*time.Millisecond, func(port int, alive bool) {
		mu.Lock()
		lastAlive = alive
		callCount++
		mu.Unlock()
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port

	w.Watch(port)
	w.Start()
	defer w.Stop()

	time.Sleep(300 * time.Millisecond)
	ln.Close()
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if lastAlive {
		t.Error("expected port to be reported as down after close")
	}
}

func TestWatcherIsAlive(t *testing.T) {
	w := New(100*time.Millisecond, func(port int, alive bool) {})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	defer ln.Close()

	w.Watch(port)
	w.Start()
	defer w.Stop()

	time.Sleep(300 * time.Millisecond)

	if !w.IsAlive(port) {
		t.Errorf("IsAlive(%d) = false, want true", port)
	}
}

func TestWatcherUnwatch(t *testing.T) {
	w := New(100*time.Millisecond, func(port int, alive bool) {})

	w.Watch(3000)
	w.Unwatch(3000)

	ports := w.WatchedPorts()
	if len(ports) != 0 {
		t.Errorf("WatchedPorts() = %v, want empty", ports)
	}
}

func TestWatcherOnlyReportsChanges(t *testing.T) {
	var mu sync.Mutex
	var callCount int

	w := New(100*time.Millisecond, func(port int, alive bool) {
		mu.Lock()
		callCount++
		mu.Unlock()
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	defer ln.Close()

	w.Watch(port)
	w.Start()
	defer w.Stop()

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	count := callCount
	mu.Unlock()

	// Should fire exactly once (the initial up transition), not on every poll
	if count != 1 {
		t.Errorf("onChange called %d times, want 1 (only on state change)", count)
	}
}

func findFreeClosedPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 50*time.Millisecond)
	if err == nil {
		conn.Close()
		t.Fatalf("port %d still open", port)
	}
	return port
}
