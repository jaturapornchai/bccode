package handlers

import "testing"

func TestEnvFirstReturnsFirstNonEmptyTrimmedValue(t *testing.T) {
	t.Setenv("BC_TEST_EMPTY", " ")
	t.Setenv("BC_TEST_SECOND", "  atlas-value  ")

	got := envFirst("BC_TEST_EMPTY", "BC_TEST_SECOND")
	if got != "atlas-value" {
		t.Fatalf("envFirst() = %q, want %q", got, "atlas-value")
	}
}
