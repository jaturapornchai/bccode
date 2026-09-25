package main

import "testing"

func TestValidReloadConfigSecret(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		supplied string
		want     bool
	}{
		{name: "matching", expected: "server-secret", supplied: "server-secret", want: true},
		{name: "missing deployment secret", supplied: "anything", want: false},
		{name: "missing supplied secret", expected: "server-secret", want: false},
		{name: "wrong secret", expected: "server-secret", supplied: "wrong-secret", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validReloadConfigSecret(test.expected, test.supplied); got != test.want {
				t.Fatalf("validReloadConfigSecret() = %v, want %v", got, test.want)
			}
		})
	}
}

// LINE linking is per user and the workspace header offers it before a holding is picked, so the
// link-code routes must not answer "Shop not selected." (which the BFF turns into "log in again").
func TestExceptShopPathsAllowLineLinkWithoutHolding(t *testing.T) {
	for _, path := range []string{"/profile/link-line", "/profile/link-line/code", "/profile/link-line/code/check"} {
		found := false
		for _, allowed := range exceptShopPaths {
			if allowed == path {
				found = true
			}
		}
		if !found {
			t.Errorf("exceptShopPaths is missing %q", path)
		}
	}
}
