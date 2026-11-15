package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qrocodile-io/qrocodile-cli/internal/browser"
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

// fakeRenderAPI serves a fixed SVG with the given response headers.
func fakeRenderAPI(t *testing.T, headers map[string]string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		_, _ = w.Write([]byte("<svg>fake</svg>"))
	}))
	t.Cleanup(server.Close)
	t.Setenv("QROCODILE_API_KEY", "qk_live_test")
	t.Setenv("QROCODILE_API_BASE_URL", server.URL)
}

// runGenerateCapturing runs generate with the flags already set and returns stdout and stderr.
func runGenerateCapturing(t *testing.T) (string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	generateCmd.SetOut(&out)
	generateCmd.SetErr(&errOut)
	t.Cleanup(func() {
		generateCmd.SetOut(nil)
		generateCmd.SetErr(nil)
	})
	if err := runGenerate(generateCmd, []string{"https://qrocodile.io"}); err != nil {
		t.Fatalf("runGenerate: %v", err)
	}
	return out.String(), errOut.String()
}

const fakeDesignerURL = "https://qrocodile.io/?c=abc&utm_source=api"

func TestRunGenerate_printsDesignerURLToStderr(t *testing.T) {
	fakeRenderAPI(t, map[string]string{"Link": "<" + fakeDesignerURL + `>; rel="edit-form"`})
	resetGenerateFlags(t)

	out, errOut := runGenerateCapturing(t)
	if out != "<svg>fake</svg>" {
		t.Errorf("stdout = %q, want only the image", out)
	}
	if errOut != "Edit in the QR Designer: "+fakeDesignerURL+"\n" {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunGenerate_explainsMissingDesignerURL(t *testing.T) {
	fakeRenderAPI(t, map[string]string{"X-Designer-Url-Unavailable": "logo"})
	resetGenerateFlags(t)

	_, errOut := runGenerateCapturing(t)
	if errOut != "No QR Designer link: the link can’t carry logo\n" {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunGenerate_noDesignerURLFlag(t *testing.T) {
	fakeRenderAPI(t, map[string]string{"Link": "<" + fakeDesignerURL + `>; rel="edit-form"`})
	resetGenerateFlags(t)
	must(t, generateCmd.Flags().Set("no-designer-url", "true"))

	if _, errOut := runGenerateCapturing(t); errOut != "" {
		t.Errorf("stderr = %q, want nothing", errOut)
	}
}

func TestRunGenerate_openOpensDesignerURL(t *testing.T) {
	fakeRenderAPI(t, map[string]string{"Link": "<" + fakeDesignerURL + `>; rel="edit-form"`})
	resetGenerateFlags(t)
	must(t, generateCmd.Flags().Set("open", "true"))
	var opened string
	openBrowser = func(url string) error { opened = url; return nil }
	t.Cleanup(func() { openBrowser = browser.Open })

	if _, errOut := runGenerateCapturing(t); errOut != "" {
		t.Errorf("stderr = %q, want nothing", errOut)
	}
	if opened != fakeDesignerURL {
		t.Errorf("opened %q, want %q", opened, fakeDesignerURL)
	}
}

func TestRunGenerate_marginBelowOne(t *testing.T) {
	resetGenerateFlags(t)
	must(t, generateCmd.Flags().Set("margin", "0"))
	err := runGenerate(generateCmd, []string{"content"})
	if err == nil || !strings.Contains(err.Error(), "--margin") {
		t.Fatalf("err = %v, want a --margin error", err)
	}
}

// Every example in generate's --help must pass the CLI's own flag checks.
func TestRunGenerate_helpExamplesAreValid(t *testing.T) {
	fakeRenderAPI(t, nil)
	examples := 0
	for _, line := range strings.Split(generateCmd.Long, "\n") {
		rest, ok := strings.CutPrefix(strings.TrimSpace(line), "qrocodile generate ")
		if !ok {
			continue
		}
		examples++
		t.Run(rest, func(t *testing.T) {
			resetGenerateFlags(t)
			args := splitExampleArgs(t, rest)
			must(t, generateCmd.ParseFlags(args))
			// Keep the example's -o out of the working directory.
			must(t, generateCmd.Flags().Set("output", t.TempDir()+"/out"))
			runGenerateCapturing(t)
		})
	}
	if examples == 0 {
		t.Fatal("found no examples in generate's --help")
	}
}

// splitExampleArgs splits a help example into arguments, honoring double quotes.
func splitExampleArgs(t *testing.T, s string) []string {
	t.Helper()
	var args []string
	var cur strings.Builder
	inQuotes, inArg := false, false
	for _, r := range s {
		switch {
		case r == '"':
			inQuotes, inArg = !inQuotes, true
		case r == ' ' && !inQuotes:
			if inArg {
				args = append(args, cur.String())
				cur.Reset()
				inArg = false
			}
		default:
			cur.WriteRune(r)
			inArg = true
		}
	}
	if inQuotes {
		t.Fatalf("unbalanced quotes in %q", s)
	}
	if inArg {
		args = append(args, cur.String())
	}
	return args
}
