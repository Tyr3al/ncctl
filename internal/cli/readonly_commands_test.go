package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadOnlyCommandsAreRegistered(t *testing.T) {
	for _, args := range [][]string{
		{"servers", "list", "--help"},
		{"servers", "get", "--help"},
		{"interfaces", "list", "--help"},
		{"interfaces", "get", "--help"},
		{"failover", "list", "--help"},
		{"tasks", "list", "--help"},
		{"tasks", "get", "--help"},
		{"rdns", "get", "--help"},
	} {
		cmd := NewRootCommand()
		out := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(out)
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("%v Execute() error = %v", args, err)
		}
		if !strings.Contains(out.String(), "Usage:") {
			t.Fatalf("%v help output missing usage: %q", args, out.String())
		}
	}
}

func TestWriteTable(t *testing.T) {
	out := &bytes.Buffer{}
	if err := writeTable(out, []string{"ID", "NAME"}, [][]string{{"1", "vps"}}); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); !strings.Contains(got, "ID") || !strings.Contains(got, "vps") {
		t.Fatalf("table output = %q", got)
	}
}
