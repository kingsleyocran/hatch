//go:build windows

package platform

import (
	"strings"
	"testing"
)

func TestWindowsHostsEntry(t *testing.T) {
	w := &Windows{}
	content := w.hostsEntry("cayacart.test")

	if !strings.Contains(content, "127.0.0.1") {
		t.Error("hosts entry should contain 127.0.0.1")
	}
	if !strings.Contains(content, "cayacart.test") {
		t.Error("hosts entry should contain domain")
	}
}

func TestWindowsNetshRule(t *testing.T) {
	w := &Windows{}
	args := w.netshArgs(80, 8443)

	found80 := false
	found8443 := false
	for _, a := range args {
		if strings.Contains(a, "80") {
			found80 = true
		}
		if strings.Contains(a, "8443") {
			found8443 = true
		}
	}
	if !found80 || !found8443 {
		t.Errorf("netsh args should contain both ports, got %v", args)
	}
}

func TestWindowsNeedsSudo(t *testing.T) {
	w := &Windows{}
	if !w.NeedsSudo() {
		t.Error("Windows should need admin for setup")
	}
}
