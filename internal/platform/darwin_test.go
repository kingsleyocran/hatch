//go:build darwin

package platform

import (
	"strings"
	"testing"
)

func TestDarwinResolverFileContent(t *testing.T) {
	d := &Darwin{}
	content := d.resolverContent(15353)

	if !strings.Contains(content, "nameserver 127.0.0.1") {
		t.Error("resolver content should contain nameserver 127.0.0.1")
	}
	if !strings.Contains(content, "port 15353") {
		t.Error("resolver content should contain port 15353")
	}
}

func TestDarwinLaunchdPlistContent(t *testing.T) {
	d := &Darwin{}
	content := d.launchdPlist("/usr/local/bin/hatch", "/tmp/hatch.sock")

	if !strings.Contains(content, "/usr/local/bin/hatch") {
		t.Error("plist should contain binary path")
	}
	if !strings.Contains(content, "com.hatch.daemon") {
		t.Error("plist should contain service label")
	}
}

func TestDarwinPfctlRuleContent(t *testing.T) {
	d := &Darwin{}
	content := d.pfctlRule(80, 8443)

	if !strings.Contains(content, "rdr pass on lo0") {
		t.Error("pfctl rule should redirect on lo0")
	}
	if !strings.Contains(content, "port 80") {
		t.Error("pfctl rule should contain from port")
	}
	if !strings.Contains(content, "port 8443") {
		t.Error("pfctl rule should contain to port")
	}
}

func TestDarwinNeedsSudo(t *testing.T) {
	d := &Darwin{}
	if !d.NeedsSudo() {
		t.Error("Darwin should need sudo for setup")
	}
}
