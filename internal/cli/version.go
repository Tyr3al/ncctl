package cli

import (
	"runtime/debug"
	"strings"
)

// These are injected at build time via -ldflags.
var (
	version   = "dev"
	commit    = ""
	buildDate = ""
)

// versionInfo returns a multi-line version string for the given binary name,
// including commit and build date when available.
func versionInfo(binary string) string {
	v := version
	c := commit
	d := buildDate

	if info, ok := debug.ReadBuildInfo(); ok {
		if v == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
			v = info.Main.Version
		}
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				if c == "" && len(s.Value) >= 7 {
					c = s.Value[:7]
				}
			case "vcs.time":
				if d == "" {
					d = s.Value
				}
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(binary + " " + v)
	if c != "" {
		sb.WriteString("\ncommit: " + c)
	}
	if d != "" {
		sb.WriteString("\nbuilt:  " + d)
	}
	return sb.String()
}
