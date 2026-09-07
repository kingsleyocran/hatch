package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProjectNode(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "cayacart-web"}`), 0644)

	name, ptype := DetectProject(dir)
	if name != "cayacart-web" {
		t.Errorf("name = %q, want %q", name, "cayacart-web")
	}
	if ptype != "node" {
		t.Errorf("type = %q, want %q", ptype, "node")
	}
}

func TestDetectProjectGo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/example/myapp\n\ngo 1.26\n"), 0644)

	name, ptype := DetectProject(dir)
	if name != "myapp" {
		t.Errorf("name = %q, want %q", name, "myapp")
	}
	if ptype != "go" {
		t.Errorf("type = %q, want %q", ptype, "go")
	}
}

func TestDetectProjectCargo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"rustapp\"\n"), 0644)

	name, ptype := DetectProject(dir)
	if name != "rustapp" {
		t.Errorf("name = %q, want %q", name, "rustapp")
	}
	if ptype != "rust" {
		t.Errorf("type = %q, want %q", ptype, "rust")
	}
}

func TestDetectProjectPython(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]\nname = \"pyapp\"\n"), 0644)

	name, ptype := DetectProject(dir)
	if name != "pyapp" {
		t.Errorf("name = %q, want %q", name, "pyapp")
	}
	if ptype != "python" {
		t.Errorf("type = %q, want %q", ptype, "python")
	}
}

func TestDetectProjectFallback(t *testing.T) {
	dir := t.TempDir()

	name, ptype := DetectProject(dir)
	expected := filepath.Base(dir)
	if name != expected {
		t.Errorf("name = %q, want folder name %q", name, expected)
	}
	if ptype != "" {
		t.Errorf("type = %q, want empty", ptype)
	}
}
