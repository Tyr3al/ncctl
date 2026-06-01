package cli

import (
	"bytes"
	"testing"
	"time"
)

func TestRootCommandParsesGlobalFlags(t *testing.T) {
	cmd := NewRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{
		"--config", "/tmp/netcupctl.json",
		"--api-base-url", "https://api.example.test/scp-core",
		"--auth-base-url", "https://auth.example.test",
		"--timeout", "5s",
		"--json",
		"version",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got, want := out.String(), "netcupctl dev\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}

	opts, ok := commandOptions(cmd)
	if !ok {
		t.Fatal("root options missing")
	}
	if opts.ConfigPath != "/tmp/netcupctl.json" {
		t.Fatalf("ConfigPath = %q", opts.ConfigPath)
	}
	if opts.APIBaseURL != "https://api.example.test/scp-core" {
		t.Fatalf("APIBaseURL = %q", opts.APIBaseURL)
	}
	if opts.AuthBaseURL != "https://auth.example.test" {
		t.Fatalf("AuthBaseURL = %q", opts.AuthBaseURL)
	}
	if opts.Timeout != 5*time.Second {
		t.Fatalf("Timeout = %s", opts.Timeout)
	}
	if !opts.JSON {
		t.Fatal("JSON = false, want true")
	}
}
