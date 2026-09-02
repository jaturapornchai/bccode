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
