package cmd

import "testing"

func TestMaskKey(t *testing.T) {
	cases := map[string]string{
		"qk_live_AAAAAAAAAAAAAAAAAAAAAAAAAAAAtest": "qk_live_…test",
		"short": "…",
		"":      "…",
	}
	for key, want := range cases {
		if got := maskKey(key); got != want {
			t.Errorf("maskKey(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestMaskKey_neverContainsTheMiddleOfTheKey(t *testing.T) {
	key := "qk_live_AAAAAAAAAAAAAAAAAAAAAAAAAAAAtest"
	middle := key[10:20]
	if got := maskKey(key); got == key || contains(got, middle) {
		t.Errorf("maskKey(%q) = %q leaks the key's middle segment %q", key, got, middle)
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
