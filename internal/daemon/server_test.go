package daemon

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/kingsleyocran/hatch/internal/config"
	"github.com/kingsleyocran/hatch/internal/project"
)

func setupTestDaemon(t *testing.T) (*Daemon, string) {
	t.Helper()
	dir := t.TempDir()

	cfg := config.Default()
	cfg.DNSPort = 0 // let OS pick
	cfg.DaemonPort = 0

	sockPath := filepath.Join(dir, "hatch.sock")
	storePath := filepath.Join(dir, "projects.yaml")

	store, err := project.Load(storePath)
	if err != nil {
		t.Fatalf("load store: %v", err)
	}

	d := NewDaemon(cfg, store, sockPath)
	return d, sockPath
}

func TestDaemonPing(t *testing.T) {
	d, sockPath := setupTestDaemon(t)

	if err := d.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer d.Stop()

	time.Sleep(100 * time.Millisecond)

	c := NewClient(sockPath)
	if err := c.Ping(); err != nil {
		t.Fatalf("Ping() error: %v", err)
	}
}

func TestDaemonAddAndList(t *testing.T) {
	d, sockPath := setupTestDaemon(t)

	if err := d.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer d.Stop()

	time.Sleep(100 * time.Millisecond)

	c := NewClient(sockPath)

	if err := c.Add("cayacart.test", 3000, "/tmp/cayacart", false); err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	projects, err := c.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("List() returned %d, want 1", len(projects))
	}
	if projects[0].Domain != "cayacart.test" {
		t.Errorf("Domain = %q, want %q", projects[0].Domain, "cayacart.test")
	}
	if projects[0].Port != 3000 {
		t.Errorf("Port = %d, want %d", projects[0].Port, 3000)
	}
}

func TestDaemonRemove(t *testing.T) {
	d, sockPath := setupTestDaemon(t)

	if err := d.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer d.Stop()

	time.Sleep(100 * time.Millisecond)

	c := NewClient(sockPath)
	c.Add("cayacart.test", 3000, "/tmp/cayacart", false)

	if err := c.Remove("cayacart.test"); err != nil {
		t.Fatalf("Remove() error: %v", err)
	}

	projects, _ := c.List()
	if len(projects) != 0 {
		t.Errorf("List() = %d after Remove, want 0", len(projects))
	}
}
