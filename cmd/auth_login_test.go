package cmd

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/qrocodile-io/qrocodile-cli/internal/config"
)

func withTempConfigDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	switch runtime.GOOS {
	case "windows":
		t.Setenv("AppData", dir)
	case "darwin":
		t.Setenv("HOME", dir)
	default:
		t.Setenv("XDG_CONFIG_HOME", dir)
		t.Setenv("HOME", dir)
	}
}

func TestLoginWithKey_readsOnlyOneLine(t *testing.T) {
	withTempConfigDir(t)

	// Regression test: loginWithKey used to io.ReadAll stdin, which blocks forever on an
	// interactive terminal (EOF there is Ctrl+D, not Enter). It now reads a single line, so
	// anything after the first newline is simply never read, not an error.
	cmd := authLoginCmd
	cmd.SetIn(bytes.NewBufferString("qk_live_first\nqk_live_second\n"))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	if err := loginWithKey(cmd); err != nil {
		t.Fatalf("loginWithKey: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if got := cfg.APIKey(); got != "qk_live_first" {
		t.Errorf("stored key = %q, want qk_live_first", got)
	}
}

func TestLoginWithKey_emptyInputErrors(t *testing.T) {
	withTempConfigDir(t)

	cmd := authLoginCmd
	cmd.SetIn(bytes.NewBufferString("\n"))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	if err := loginWithKey(cmd); err == nil {
		t.Fatal("expected an error for empty input")
	}
}
