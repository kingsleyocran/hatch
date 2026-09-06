package daemon

import (
	"path/filepath"
	"testing"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/kingsleyocran/hatch/internal/project"
)

func TestDaemonRunning(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default()
	cfg.DNSPort = 0
	cfg.DaemonPort = 0

	sockPath := filepath.Join(dir, "hatch.sock")
	store, err := project.Load(filepath.Join(dir, "projects.yaml"))
	if err != nil {
		t.Fatalf("load store: %v", err)
	}

	d := NewDaemon(cfg, store, sockPath)

	if d.Running() {
		t.Error("Running() = true before Start, want false")
	}

	if err := d.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	if !d.Running() {
		t.Error("Running() = false after Start, want true")
	}

	if err := d.Stop(); err != nil {
		t.Fatalf("Stop() error: %v", err)
	}

	if d.Running() {
		t.Error("Running() = true after Stop, want false")
	}
}

func TestDaemonDoubleStart(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default()
	cfg.DNSPort = 0
	cfg.DaemonPort = 0

	sockPath := filepath.Join(dir, "hatch.sock")
	store, err := project.Load(filepath.Join(dir, "projects.yaml"))
	if err != nil {
		t.Fatalf("load store: %v", err)
	}

	d := NewDaemon(cfg, store, sockPath)

	if err := d.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer d.Stop()

	err = d.Start()
	if err == nil {
		t.Error("second Start() should return error")
	}
}

func TestDaemonHandleUnknownAction(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default()
	cfg.DNSPort = 0
	cfg.DaemonPort = 0

	store, err := project.Load(filepath.Join(dir, "projects.yaml"))
	if err != nil {
		t.Fatalf("load store: %v", err)
	}

	d := NewDaemon(cfg, store, filepath.Join(dir, "hatch.sock"))

	resp := d.handleRequest(Request{Action: "bogus"})
	if resp.OK {
		t.Error("handleRequest with unknown action should return OK=false")
	}
}
