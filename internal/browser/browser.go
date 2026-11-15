// Package browser opens a URL in the user's default browser.
package browser

import (
	"os/exec"
	"runtime"
)

// Open starts the platform's URL handler on url and returns without waiting for the browser.
func Open(url string) error {
	return command(runtime.GOOS, url).Start()
}

func command(goos, url string) *exec.Cmd {
	switch goos {
	case "darwin":
		return exec.Command("open", url)
	case "windows":
		// Not `cmd /c start`, which would read the URL's `&` as a command separator.
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return exec.Command("xdg-open", url)
	}
}
