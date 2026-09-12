// Package output implements the CLI's -o/no--o semantics: -o is a dumb byte-writer, and the
// no--o default follows curl's own behavior — refuse to write binary data to an interactive
// terminal rather than corrupt it.
package output

import (
	"fmt"
	"io"
	"os"
)

// Write sends data to path, or to stdout if path is empty or "-". binary marks data as
// non-text (PNG bytes): with no path, writing binary data to an interactive terminal is
// refused unless path is explicitly "-", matching curl's own guard. Text data (SVG source)
// always prints to stdout when no path is given — there's nothing to corrupt.
func Write(path string, data []byte, binary bool, stdout io.Writer) error {
	switch path {
	case "-":
		_, err := stdout.Write(data)
		return err
	case "":
		if binary && isTerminal(stdout) {
			return fmt.Errorf("refusing to write binary data to your terminal; use -o <file> to save it, or -o - to force it anyway")
		}
		_, err := stdout.Write(data)
		return err
	default:
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
		return nil
	}
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
