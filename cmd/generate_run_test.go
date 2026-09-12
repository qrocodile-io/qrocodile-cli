package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunGenerate_invalidFormat(t *testing.T) {
	resetGenerateFlags(t)
	must(t, generateCmd.Flags().Set("format", "jpeg"))
	err := runGenerate(generateCmd, []string{"content"})
	if err == nil || !strings.Contains(err.Error(), "--format") {
		t.Fatalf("err = %v, want a --format error", err)
	}
}

func TestRunGenerate_logoRequiresLogoColor(t *testing.T) {
	resetGenerateFlags(t)
	must(t, generateCmd.Flags().Set("logo", "website"))
	err := runGenerate(generateCmd, []string{"content"})
	if err == nil || !strings.Contains(err.Error(), "--logo-color") {
		t.Fatalf("err = %v, want it to mention --logo-color", err)
	}
}

func TestRunGenerate_logoColorRequiresLogo(t *testing.T) {
	resetGenerateFlags(t)
	must(t, generateCmd.Flags().Set("logo-color", "#000000"))
	err := runGenerate(generateCmd, []string{"content"})
	if err == nil || !strings.Contains(err.Error(), "--logo") {
		t.Fatalf("err = %v, want it to mention --logo", err)
	}
}

func TestRunGenerate_svgViaFakeAPI(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<svg>fake</svg>"))
	}))
	defer server.Close()

	t.Setenv("QROCODILE_API_KEY", "qk_live_test")
	t.Setenv("QROCODILE_API_BASE_URL", server.URL)

	resetGenerateFlags(t)
	var out bytes.Buffer
	generateCmd.SetOut(&out)
	if err := runGenerate(generateCmd, []string{"https://qrocodile.io"}); err != nil {
		t.Fatalf("runGenerate: %v", err)
	}
	if out.String() != "<svg>fake</svg>" {
		t.Errorf("out = %q", out.String())
	}
	if gotAuth != "Bearer qk_live_test" {
		t.Errorf("Authorization = %q", gotAuth)
	}
}

func TestRunGenerate_noKeyConfiguredErrors(t *testing.T) {
	t.Setenv("QROCODILE_API_KEY", "")
	t.Setenv("HOME", t.TempDir()) // no config file with a stored key here either

	resetGenerateFlags(t)
	err := runGenerate(generateCmd, []string{"https://qrocodile.io"})
	if err == nil || !strings.Contains(err.Error(), "not logged in") {
		t.Fatalf("err = %v, want a not-logged-in error", err)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
