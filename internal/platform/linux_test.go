//go:build linux

package platform

import (
	"strings"
	"testing"
)

func TestLinuxResolvedContent(t *testing.T) {
	l := &Linux{}
	content := l.resolvedConf(15353)

	if !strings.Contains(content, "DNS=127.0.0.1:15353") {
		t.Error("resolved config should contain DNS address")
	}
}

func TestLinuxSystemdUnitContent(t *testing.T) {
	l := &Linux{}
	content := l.systemdUnit("/usr/local/bin/hatch", "/tmp/hatch.sock")

	if !strings.Contains(content, "/usr/local/bin/hatch") {
		t.Error("unit should contain binary path")
	}
	if !strings.Contains(content, "hatch.service") {
		t.Error("unit should reference service name")
	}
}

func TestLinuxIptablesRuleContent(t *testing.T) {
	l := &Linux{}
	content := l.iptablesRule(80, 8443)

	if !strings.Contains(content, "80") {
		t.Error("rule should contain from port")
	}
	if !strings.Contains(content, "8443") {
		t.Error("rule should contain to port")
	}
}

func TestLinuxNeedsSudo(t *testing.T) {
	l := &Linux{}
	if !l.NeedsSudo() {
		t.Error("Linux should need sudo for setup")
	}
}
