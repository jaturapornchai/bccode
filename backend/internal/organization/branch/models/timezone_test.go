package models

import (
	"errors"
	"testing"
)

func TestNormalizeIANATimezone(t *testing.T) {
	for _, test := range []struct {
		name  string
		input string
		want  string
		err   error
	}{
		{name: "IANA location", input: " Asia/Bangkok ", want: "Asia/Bangkok"},
		{name: "UTC zone", input: "UTC", want: "UTC"},
		{name: "required", err: ErrBranchTimezoneRequired},
		{name: "offset rejected", input: "+07:00", err: ErrBranchTimezoneInvalid},
		{name: "machine local rejected", input: "Local", err: ErrBranchTimezoneInvalid},
		{name: "unknown rejected", input: "Not/AZone", err: ErrBranchTimezoneInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeIANATimezone(test.input)
			if !errors.Is(err, test.err) || (test.err == nil && err != nil) {
				t.Fatalf("error = %v, want %v", err, test.err)
			}
			if got != test.want {
				t.Fatalf("timezone = %q, want %q", got, test.want)
			}
		})
	}
}
