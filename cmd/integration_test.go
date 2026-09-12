package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestIntegration_Generate exercises `generate` against a real API instance. Skipped unless
// QR_API_KEY is set (translated into this CLI's own QROCODILE_API_KEY/QROCODILE_API_BASE_URL
// env vars for the duration of the test) — QR_API_KEY/QR_API_URL is the convention shared with
// qrocodile-api-go's own integration test, so both repos read the same local .env during
// development. Set QR_API_INTEGRATION_REQUIRED=true to fail instead of skip.
func TestIntegration_Generate(t *testing.T) {
	apiKey := os.Getenv("QR_API_KEY")
	if apiKey == "" {
		if os.Getenv("QR_API_INTEGRATION_REQUIRED") == "true" {
			t.Fatal("QR_API_KEY is required (QR_API_INTEGRATION_REQUIRED=true)")
		}
		t.Skip("QR_API_KEY not set, skipping integration test")
	}
	t.Setenv("QROCODILE_API_KEY", apiKey)
	if baseURL := os.Getenv("QR_API_URL"); baseURL != "" {
		t.Setenv("QROCODILE_API_BASE_URL", baseURL)
	}

	resetGenerateFlags(t)
	must(t, generateCmd.Flags().Set("preset", "ocean"))
	var out bytes.Buffer
	generateCmd.SetOut(&out)
	if err := runGenerate(generateCmd, []string{"https://qrocodile.io"}); err != nil {
		t.Fatalf("runGenerate: %v", err)
	}
	if !strings.Contains(out.String(), "<svg") {
		t.Errorf("output doesn't look like SVG: %.200s", out.String())
	}
}

func TestIntegration_Generate_invalidKeyIsAPIError(t *testing.T) {
	if os.Getenv("QR_API_KEY") == "" && os.Getenv("QR_API_INTEGRATION_REQUIRED") != "true" {
		t.Skip("QR_API_KEY not set, skipping integration test")
	}
	t.Setenv("QROCODILE_API_KEY", "qk_live_definitely-not-a-real-key")
	if baseURL := os.Getenv("QR_API_URL"); baseURL != "" {
		t.Setenv("QROCODILE_API_BASE_URL", baseURL)
	}

	resetGenerateFlags(t)
	generateCmd.SetOut(&bytes.Buffer{})
	err := runGenerate(generateCmd, []string{"https://qrocodile.io"})
	if err == nil {
		t.Fatal("expected an error for an invalid key")
	}
	if !strings.Contains(err.Error(), "UNAUTHORIZED") {
		t.Errorf("err = %v, want it to mention UNAUTHORIZED", err)
	}
}
