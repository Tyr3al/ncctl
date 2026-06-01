package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfirmRiskyRequiresYesUnlessBypassed(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetIn(strings.NewReader("no\n"))
	err := confirmRisky(cmd, &options{}, "Danger")
	if err == nil {
		t.Fatal("confirmRisky() error = nil, want abort")
	}

	cmd = NewRootCommand()
	cmd.SetIn(strings.NewReader(""))
	if err := confirmRisky(cmd, &options{Yes: true}, "Danger"); err != nil {
		t.Fatalf("confirmRisky() with --yes error = %v", err)
	}
}

func TestServerPowerCommandRequiresConfirmation(t *testing.T) {
	cmd := NewRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetIn(strings.NewReader("no\n"))
	cmd.SetArgs([]string{"--config", "missing", "servers", "power", "1", "OFF"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "aborted") {
		t.Fatalf("error = %v, want aborted before API call", err)
	}
}
