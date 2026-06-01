package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveWritesConfigWithRestrictedPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "netcupctl", "config.json")
	cfg := &Config{APIBaseURL: "https://api.example.test", AuthBaseURL: "https://auth.example.test", UserID: 12, Refresh: "refresh"}
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Fatalf("permissions = %o, want %o", got, want)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Refresh != "refresh" || loaded.UserID != 12 {
		t.Fatalf("loaded = %#v", loaded)
	}
}
