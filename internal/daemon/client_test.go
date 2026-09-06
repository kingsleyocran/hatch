package daemon

import (
	"testing"
)

func TestClientPingFailsWhenNoDaemon(t *testing.T) {
	c := NewClient("/nonexistent/hatch.sock")
	err := c.Ping()
	if err == nil {
		t.Error("Ping() should fail when no daemon is running")
	}
}
