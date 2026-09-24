package rolepermission

import "testing"

// review 2026-09-24: the delete needs the __v the screen loaded, like the update — without it a
// stale delete could switch off a permission set someone saved meanwhile.
func TestDeleteVersionRequired(t *testing.T) {
	for _, raw := range []string{"", "  ", "abc", "-1", "1.5"} {
		if _, err := deleteVersion(raw); err == nil {
			t.Errorf("deleteVersion(%q) accepted", raw)
		}
	}
	if version, err := deleteVersion(" 3 "); err != nil || version != 3 {
		t.Fatalf("deleteVersion(3) = %d, %v", version, err)
	}
	if version, err := deleteVersion("0"); err != nil || version != 0 {
		t.Fatalf("deleteVersion(0) = %d, %v", version, err)
	}
}
