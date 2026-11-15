package browser

import (
	"slices"
	"testing"
)

func TestCommand(t *testing.T) {
	const url = "https://qrocodile.io/?a=1&b=2"
	cases := map[string][]string{
		"darwin":  {"open", url},
		"windows": {"rundll32", "url.dll,FileProtocolHandler", url},
		"linux":   {"xdg-open", url},
		"freebsd": {"xdg-open", url},
	}
	for goos, want := range cases {
		if got := command(goos, url).Args; !slices.Equal(got, want) {
			t.Errorf("command(%q) = %q, want %q", goos, got, want)
		}
	}
}
