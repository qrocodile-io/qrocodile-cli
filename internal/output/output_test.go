package output

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrite_explicitPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.svg")

	if err := Write(path, []byte("<svg/>"), false, &bytes.Buffer{}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if string(got) != "<svg/>" {
		t.Errorf("file content = %q", got)
	}
}

func TestWrite_dashForcesStdout(t *testing.T) {
	var buf bytes.Buffer
	if err := Write("-", []byte{0x89, 0x50, 0x4e, 0x47}, true, &buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if buf.Len() != 4 {
		t.Errorf("buf.Len() = %d, want 4", buf.Len())
	}
}

func TestWrite_noPathTextGoesToStdout(t *testing.T) {
	var buf bytes.Buffer
	if err := Write("", []byte("<svg/>"), false, &buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if buf.String() != "<svg/>" {
		t.Errorf("buf = %q", buf.String())
	}
}

func TestWrite_noPathBinaryToNonTerminalWritesBytes(t *testing.T) {
	// A *bytes.Buffer is never a terminal, so this exercises the "piped" branch of the
	// no-path binary case, not the refusal branch.
	var buf bytes.Buffer
	png := []byte{0x89, 0x50, 0x4e, 0x47}
	if err := Write("", png, true, &buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), png) {
		t.Errorf("buf = %v, want %v", buf.Bytes(), png)
	}
}

func TestWrite_noPathBinaryToTerminalRefuses(t *testing.T) {
	// /dev/null opened directly isn't a character device in the sense we check, so use a
	// pty-like stand-in: os.Stdin itself is usually not a terminal under `go test`, so
	// instead we open the actual controlling terminal device if available, and skip if not.
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		t.Skip("no controlling terminal available in this environment")
	}
	defer func() { _ = tty.Close() }()

	err = Write("", []byte{0x89, 0x50, 0x4e, 0x47}, true, tty)
	if err == nil {
		t.Fatal("expected a refusal error writing binary data to a terminal")
	}
	if !strings.Contains(err.Error(), "refusing") {
		t.Errorf("error = %v, want it to mention refusing", err)
	}
}
