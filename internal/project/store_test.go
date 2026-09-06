package project

import (
	"path/filepath"
	"testing"
	"time"
)

func TestAddAndList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.yaml")
	store, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	p := Project{
		Dir:     "/Users/test/projects/cayacart",
		Domain:  "cayacart.test",
		Port:    3000,
		HTTPS:   false,
		Created: time.Now(),
	}

	if err := store.Add(p); err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	projects := store.List()
	if len(projects) != 1 {
		t.Fatalf("List() returned %d projects, want 1", len(projects))
	}
	if projects[0].Domain != "cayacart.test" {
		t.Errorf("Domain = %q, want %q", projects[0].Domain, "cayacart.test")
	}
}

func TestAddDuplicateDomainFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.yaml")
	store, _ := Load(path)

	p := Project{Dir: "/a", Domain: "cayacart.test", Port: 3000, Created: time.Now()}
	store.Add(p)

	p2 := Project{Dir: "/b", Domain: "cayacart.test", Port: 4000, Created: time.Now()}
	err := store.Add(p2)
	if err == nil {
		t.Error("Add() should fail for duplicate domain")
	}
}

func TestRemove(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.yaml")
	store, _ := Load(path)

	store.Add(Project{Dir: "/a", Domain: "cayacart.test", Port: 3000, Created: time.Now()})
	store.Add(Project{Dir: "/b", Domain: "phamel.test", Port: 8000, Created: time.Now()})

	if err := store.Remove("cayacart.test"); err != nil {
		t.Fatalf("Remove() error: %v", err)
	}

	projects := store.List()
	if len(projects) != 1 {
		t.Fatalf("List() returned %d projects, want 1", len(projects))
	}
	if projects[0].Domain != "phamel.test" {
		t.Errorf("remaining Domain = %q, want %q", projects[0].Domain, "phamel.test")
	}
}

func TestRemoveNonExistent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.yaml")
	store, _ := Load(path)

	err := store.Remove("nonexistent.test")
	if err == nil {
		t.Error("Remove() should fail for non-existent domain")
	}
}

func TestFindByDomain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.yaml")
	store, _ := Load(path)

	store.Add(Project{Dir: "/a", Domain: "cayacart.test", Port: 3000, Created: time.Now()})

	p, err := store.FindByDomain("cayacart.test")
	if err != nil {
		t.Fatalf("FindByDomain() error: %v", err)
	}
	if p.Port != 3000 {
		t.Errorf("Port = %d, want %d", p.Port, 3000)
	}

	_, err = store.FindByDomain("nonexistent.test")
	if err == nil {
		t.Error("FindByDomain() should fail for non-existent domain")
	}
}

func TestFindByPort(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.yaml")
	store, _ := Load(path)

	store.Add(Project{Dir: "/a", Domain: "cayacart.test", Port: 3000, Created: time.Now()})
	store.Add(Project{Dir: "/b", Domain: "phamel.test", Port: 3000, Created: time.Now()})
	store.Add(Project{Dir: "/c", Domain: "orborbit.test", Port: 5173, Created: time.Now()})

	matches := store.FindByPort(3000)
	if len(matches) != 2 {
		t.Fatalf("FindByPort(3000) returned %d, want 2", len(matches))
	}
}

func TestSaveAndReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.yaml")
	store, _ := Load(path)

	store.Add(Project{Dir: "/a", Domain: "cayacart.test", Port: 3000, Created: time.Now()})
	if err := store.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	projects := reloaded.List()
	if len(projects) != 1 {
		t.Fatalf("reloaded List() = %d, want 1", len(projects))
	}
	if projects[0].Domain != "cayacart.test" {
		t.Errorf("reloaded Domain = %q, want %q", projects[0].Domain, "cayacart.test")
	}
}
