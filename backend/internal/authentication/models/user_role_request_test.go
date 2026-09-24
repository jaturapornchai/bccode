package models

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestUserRoleRequestAllowsEmptyAccessExpiryDate(t *testing.T) {
	for _, body := range []string{
		`{"username":"uat-user","accessexpirydate":""}`,
		`{"username":"uat-user","accessexpirydate":null}`,
		`{"username":"uat-user","accessexpirydate":"  "}`,
		`{"username":"uat-user"}`,
	} {
		var request UserRoleRequest
		if err := json.Unmarshal([]byte(body), &request); err != nil {
			t.Fatalf("%s: empty access expiry must be accepted: %v", body, err)
		}
		if request.AccessExpiryDate != "" || request.Username != "uat-user" {
			t.Fatalf("%s: got expiry %q username %q", body, request.AccessExpiryDate, request.Username)
		}
	}
}

// The settings screen's <input type="date"> sends YYYY-MM-DD; this used to fail the whole
// user save with a time.Time parse error (save-audit finding 4).
func TestUserRoleRequestAcceptsDateInputFormat(t *testing.T) {
	var request UserRoleRequest
	body := `{"username":"somchai01","position":"พนักงานบัญชี","accessexpirydate":"2026-12-31","addexistinguser":true}`
	if err := json.Unmarshal([]byte(body), &request); err != nil {
		t.Fatalf("date input value rejected: %v", err)
	}
	if request.AccessExpiryDate != "2026-12-31" || request.Position != "พนักงานบัญชี" || !request.AddExistingUser {
		t.Fatalf("decoded = %+v", request)
	}
}

func TestUserRoleRequestKeepsRFC3339DateInItsOwnOffset(t *testing.T) {
	cases := map[string]string{
		`"2026-08-31T00:00:00Z"`:      "2026-08-31",
		`"2026-12-31T00:00:00+07:00"`: "2026-12-31",
	}
	for raw, want := range cases {
		var request UserRoleRequest
		if err := json.Unmarshal([]byte(`{"accessexpirydate":`+raw+`}`), &request); err != nil {
			t.Fatalf("%s rejected: %v", raw, err)
		}
		if request.AccessExpiryDate != want {
			t.Fatalf("%s → %q, want %q", raw, request.AccessExpiryDate, want)
		}
	}
}

func TestUserRoleRequestRejectsUnusableExpiryDates(t *testing.T) {
	for _, raw := range []string{`"31/12/2026"`, `"2026-02-30"`, `"2569-12-31"`, `"1999-12-31"`, `20261231`, `"tomorrow"`} {
		var request UserRoleRequest
		err := json.Unmarshal([]byte(`{"username":"u","accessexpirydate":`+raw+`}`), &request)
		if !errors.Is(err, ErrInvalidAccessExpiryDate) {
			t.Fatalf("%s: err = %v, want ErrInvalidAccessExpiryDate", raw, err)
		}
	}
}

func TestAccessExpiryDayUsesInstantLocation(t *testing.T) {
	bangkok := time.FixedZone("ICT", 7*60*60)
	// the instant is 00:00 of the day after the last usable day (AccessEndsAt)
	if got := AccessExpiryDay(time.Date(2027, time.January, 1, 0, 0, 0, 0, bangkok)); got != "2026-12-31" {
		t.Fatalf("AccessExpiryDay = %q", got)
	}
	stored := time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC) // lib/pq DATE
	if got := AccessExpiryDay(AccessEndsAt(stored, "Asia/Bangkok")); got != "2026-12-31" {
		t.Fatalf("round trip = %q", got)
	}
	if got := AccessExpiryDay(time.Time{}); got != "" {
		t.Fatalf("zero instant = %q, want empty", got)
	}
}

// "ใช้งานได้ถึงสิ้นวันที่กำหนด": the whole expiry date is usable in the Holding's timezone,
// access ends at 00:00 of the next day there (review 2026-09-24: was closed ON the date).
func TestAccessEndsAtInclusiveInHoldingTimezone(t *testing.T) {
	stored := time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC) // lib/pq DATE
	ends := AccessEndsAt(stored, "Asia/Bangkok")
	if want := time.Date(2026, time.December, 31, 17, 0, 0, 0, time.UTC); !ends.Equal(want) {
		t.Fatalf("ends %s, want %s", ends.UTC(), want)
	}
	for _, tc := range []struct {
		now     time.Time
		expired bool
	}{
		{time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC), false},    // 07:00 on the expiry date (Bangkok)
		{time.Date(2026, time.December, 31, 16, 59, 59, 0, time.UTC), false}, // 23:59:59 on the expiry date (Bangkok)
		{time.Date(2026, time.December, 31, 17, 0, 0, 0, time.UTC), true},    // 00:00 the next day (Bangkok)
	} {
		if got := AccessExpired(ends, tc.now); got != tc.expired {
			t.Errorf("now %s: expired %v, want %v", tc.now, got, tc.expired)
		}
	}
	if AccessExpired(AccessEndsAt(time.Time{}, "Asia/Bangkok"), time.Now()) {
		t.Fatal("no expiry date must never expire")
	}
	if got := AccessEndsAt(stored, "Asia/Tokyo"); !got.Equal(time.Date(2026, time.December, 31, 15, 0, 0, 0, time.UTC)) {
		t.Fatalf("Tokyo ends %s", got.UTC())
	}
	if got := HoldingLocation(" Not/AZone ").String(); got != "Asia/Bangkok" {
		t.Fatalf("unknown timezone falls back to %s", got)
	}
}
