package utils

import "testing"

func TestNormalizeBusinessCode(t *testing.T) {
	t.Parallel()

	if got := NormalizeBusinessCode(" ab c\t01 "); got != "ABC01" {
		t.Fatalf("NormalizeBusinessCode() = %q, want %q", got, "ABC01")
	}
}
