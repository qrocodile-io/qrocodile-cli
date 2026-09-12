package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// resetGenerateFlags restores every generateCmd flag to its default and clears Changed, so
// tests driving the real package-level command don't leak state between cases.
func resetGenerateFlags(t *testing.T) {
	t.Helper()
	generateCmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
}

func decodeDesign(t *testing.T, raw json.RawMessage) map[string]any {
	t.Helper()
	if raw == nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("decoding design JSON: %v", err)
	}
	return m
}

func TestBuildDesign_noFlagsReturnsNil(t *testing.T) {
	resetGenerateFlags(t)
	design, err := buildDesign(&cobra.Command{}, generateCmd.Flags())
	if err != nil {
		t.Fatalf("buildDesign: %v", err)
	}
	if design != nil {
		t.Errorf("design = %s, want nil", design)
	}
}

func TestBuildDesign_presetOnly(t *testing.T) {
	resetGenerateFlags(t)
	if err := generateCmd.Flags().Set("preset", "ocean"); err != nil {
		t.Fatal(err)
	}
	design, err := buildDesign(&cobra.Command{}, generateCmd.Flags())
	if err != nil {
		t.Fatalf("buildDesign: %v", err)
	}
	got := decodeDesign(t, design)
	if got["preset"] != "ocean" {
		t.Errorf("design = %v, want preset=ocean", got)
	}
}

func TestBuildDesign_moduleStyleAndFinderStyle(t *testing.T) {
	resetGenerateFlags(t)
	if err := generateCmd.Flags().Set("module-style", "woodwork"); err != nil {
		t.Fatal(err)
	}
	if err := generateCmd.Flags().Set("finder-style", "morphing"); err != nil {
		t.Fatal(err)
	}
	design, err := buildDesign(&cobra.Command{}, generateCmd.Flags())
	if err != nil {
		t.Fatalf("buildDesign: %v", err)
	}
	got := decodeDesign(t, design)
	if got["moduleStyleId"] != "woodwork" {
		t.Errorf("moduleStyleId = %v, want woodwork", got["moduleStyleId"])
	}
	if got["finderStyleId"] != "morphing" {
		t.Errorf("finderStyleId = %v, want morphing", got["finderStyleId"])
	}
}

func TestBuildDesign_logoIncludesColor(t *testing.T) {
	resetGenerateFlags(t)
	if err := generateCmd.Flags().Set("logo", "website"); err != nil {
		t.Fatal(err)
	}
	if err := generateCmd.Flags().Set("logo-color", "#000000"); err != nil {
		t.Fatal(err)
	}
	design, err := buildDesign(&cobra.Command{}, generateCmd.Flags())
	if err != nil {
		t.Fatalf("buildDesign: %v", err)
	}
	got := decodeDesign(t, design)
	logo, ok := got["logo"].(map[string]any)
	if !ok {
		t.Fatalf("logo = %v, want an object", got["logo"])
	}
	if logo["id"] != "website" {
		t.Errorf("logo.id = %v, want website", logo["id"])
	}
	if logo["color"] != "#000000" {
		t.Errorf("logo.color = %v, want #000000", logo["color"])
	}
}

func TestBuildDesign_flagsOverrideDesignFile(t *testing.T) {
	resetGenerateFlags(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "design.json")
	if err := os.WriteFile(path, []byte(`{"preset":"ocean","margin":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := generateCmd.Flags().Set("design", path); err != nil {
		t.Fatal(err)
	}
	if err := generateCmd.Flags().Set("module-color", "#ff0000"); err != nil {
		t.Fatal(err)
	}
	if err := generateCmd.Flags().Set("margin", "5"); err != nil {
		t.Fatal(err)
	}

	design, err := buildDesign(&cobra.Command{}, generateCmd.Flags())
	if err != nil {
		t.Fatalf("buildDesign: %v", err)
	}
	got := decodeDesign(t, design)

	if got["preset"] != "ocean" {
		t.Errorf("preset = %v, want ocean (from the file, untouched)", got["preset"])
	}
	if got["moduleColor"] != "#ff0000" {
		t.Errorf("moduleColor = %v, want #ff0000 (from the flag)", got["moduleColor"])
	}
	if got["margin"] != float64(5) {
		t.Errorf("margin = %v, want 5 (the flag overriding the file's 1)", got["margin"])
	}
}

func TestBuildDesign_designFromStdin(t *testing.T) {
	resetGenerateFlags(t)
	if err := generateCmd.Flags().Set("design", "-"); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString(`{"preset":"classic"}`))

	design, err := buildDesign(cmd, generateCmd.Flags())
	if err != nil {
		t.Fatalf("buildDesign: %v", err)
	}
	got := decodeDesign(t, design)
	if got["preset"] != "classic" {
		t.Errorf("design = %v, want preset=classic", got)
	}
}

func TestBuildDesign_designInlineJSON(t *testing.T) {
	resetGenerateFlags(t)
	if err := generateCmd.Flags().Set("design", `{"preset":"ocean","margin":2}`); err != nil {
		t.Fatal(err)
	}
	design, err := buildDesign(&cobra.Command{}, generateCmd.Flags())
	if err != nil {
		t.Fatalf("buildDesign: %v", err)
	}
	got := decodeDesign(t, design)
	if got["preset"] != "ocean" {
		t.Errorf("design = %v, want preset=ocean", got)
	}
}

func TestBuildDesign_designInlineJSONWithLeadingWhitespace(t *testing.T) {
	resetGenerateFlags(t)
	if err := generateCmd.Flags().Set("design", `  {"preset":"classic"}`); err != nil {
		t.Fatal(err)
	}
	design, err := buildDesign(&cobra.Command{}, generateCmd.Flags())
	if err != nil {
		t.Fatalf("buildDesign: %v", err)
	}
	got := decodeDesign(t, design)
	if got["preset"] != "classic" {
		t.Errorf("design = %v, want preset=classic", got)
	}
}

func TestBuildDesign_invalidJSONErrors(t *testing.T) {
	resetGenerateFlags(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := generateCmd.Flags().Set("design", path); err != nil {
		t.Fatal(err)
	}
	if _, err := buildDesign(&cobra.Command{}, generateCmd.Flags()); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

func TestBuildDesign_missingFileErrors(t *testing.T) {
	resetGenerateFlags(t)
	if err := generateCmd.Flags().Set("design", "/does/not/exist.json"); err != nil {
		t.Fatal(err)
	}
	if _, err := buildDesign(&cobra.Command{}, generateCmd.Flags()); err == nil {
		t.Fatal("expected an error for a missing file")
	}
}

func TestBuildDesign_errorCorrection(t *testing.T) {
	resetGenerateFlags(t)
	if err := generateCmd.Flags().Set("error-correction", "H"); err != nil {
		t.Fatal(err)
	}
	design, err := buildDesign(&cobra.Command{}, generateCmd.Flags())
	if err != nil {
		t.Fatalf("buildDesign: %v", err)
	}
	got := decodeDesign(t, design)
	if got["errorCorrection"] != "H" {
		t.Errorf("errorCorrection = %v, want H", got["errorCorrection"])
	}
}

// "auto" isn't a real API enum value — it means omit the field — so --error-correction auto
// must never end up sending the literal string "auto" to the API.
func TestBuildDesign_errorCorrectionAutoOmitsField(t *testing.T) {
	resetGenerateFlags(t)
	if err := generateCmd.Flags().Set("error-correction", "auto"); err != nil {
		t.Fatal(err)
	}
	design, err := buildDesign(&cobra.Command{}, generateCmd.Flags())
	if err != nil {
		t.Fatalf("buildDesign: %v", err)
	}
	if design != nil {
		t.Errorf("design = %s, want nil (auto sets nothing on its own)", design)
	}
}

func TestBuildDesign_errorCorrectionAutoClearsDesignValue(t *testing.T) {
	resetGenerateFlags(t)
	if err := generateCmd.Flags().Set("design", `{"errorCorrection":"L"}`); err != nil {
		t.Fatal(err)
	}
	if err := generateCmd.Flags().Set("error-correction", "auto"); err != nil {
		t.Fatal(err)
	}
	design, err := buildDesign(&cobra.Command{}, generateCmd.Flags())
	if err != nil {
		t.Fatalf("buildDesign: %v", err)
	}
	got := decodeDesign(t, design)
	if _, ok := got["errorCorrection"]; ok {
		t.Errorf("errorCorrection = %v, want cleared by --error-correction auto", got["errorCorrection"])
	}
}

func TestBuildDesign_finderColors(t *testing.T) {
	resetGenerateFlags(t)
	if err := generateCmd.Flags().Set("finder-frame-color", "#1b4332"); err != nil {
		t.Fatal(err)
	}
	if err := generateCmd.Flags().Set("finder-eye-color", "#52b788"); err != nil {
		t.Fatal(err)
	}
	design, err := buildDesign(&cobra.Command{}, generateCmd.Flags())
	if err != nil {
		t.Fatalf("buildDesign: %v", err)
	}
	got := decodeDesign(t, design)
	if got["finderFrameColor"] != "#1b4332" {
		t.Errorf("finderFrameColor = %v, want #1b4332", got["finderFrameColor"])
	}
	if got["finderEyeColor"] != "#52b788" {
		t.Errorf("finderEyeColor = %v, want #52b788", got["finderEyeColor"])
	}
}
