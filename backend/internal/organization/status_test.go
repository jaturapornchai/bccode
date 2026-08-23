package organization

import (
	"errors"
	"testing"
)

func TestResolveRequestedActiveStatus(t *testing.T) {
	for _, test := range []struct {
		name        string
		input       string
		current     bool
		want        bool
		wantChanged bool
	}{
		{name: "omitted preserves active", input: `{}`, current: true, want: true},
		{name: "omitted preserves inactive", input: `{}`, current: false, want: false},
		{name: "same status", input: `{"isactive":true}`, current: true, want: true},
		{name: "status changed", input: `{"isactive":false}`, current: true, want: false, wantChanged: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			active, changed, err := ResolveRequestedActiveStatus(test.input, test.current)
			if err != nil {
				t.Fatal(err)
			}
			if active != test.want || changed != test.wantChanged {
				t.Fatalf("status = %v, changed = %v; want %v, %v", active, changed, test.want, test.wantChanged)
			}
		})
	}
}

func TestStatusChangeReason(t *testing.T) {
	reason, err := StatusChangeReason(`{"statusreason":"  ปิดปรับปรุงระบบ  "}`)
	if err != nil || reason != "ปิดปรับปรุงระบบ" {
		t.Fatalf("reason = %q, err = %v", reason, err)
	}
	if _, err := StatusChangeReason(`{"statusreason":" "}`); !errors.Is(err, ErrStatusReasonRequired) {
		t.Fatalf("expected ErrStatusReasonRequired, got %v", err)
	}
}
