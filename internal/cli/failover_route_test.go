package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestFailoverRouteAllowsMultipleIPs(t *testing.T) {
	cmd := NewRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{
		"--config", "missing",
		"failover", "route",
		"--server-id", "123",
		"--ip", "192.0.2.10",
		"--ip", "2001:db8::/64",
	})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not logged in") {
		t.Fatalf("error = %v, want config/auth path after flag validation", err)
	}
}

func TestFailoverRouteRejectsIDWithoutFamily(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"failover", "route", "--server-id", "123", "--id", "7"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--family is required") {
		t.Fatalf("error = %v, want family validation", err)
	}
}

func TestFailoverRouteRejectsIDAndIPTogether(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"failover", "route", "--server-id", "123", "--id", "7", "--ip", "192.0.2.10"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "cannot be combined") {
		t.Fatalf("error = %v, want id/ip validation", err)
	}
}
