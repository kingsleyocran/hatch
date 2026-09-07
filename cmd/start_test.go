package cmd

import (
	"os/exec"
	"strings"
	"testing"
)

func TestStartHelpShowsForegroundFlag(t *testing.T) {
	cmd := exec.Command("go", "run", "..", "start", "--help")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("build failed: %v", err)
	}
	if !strings.Contains(string(out), "--foreground") {
		t.Error("start --help should show --foreground flag")
	}
}
