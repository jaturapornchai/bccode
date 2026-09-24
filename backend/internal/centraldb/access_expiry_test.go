package centraldb

import (
	"database/sql"
	"testing"
	"time"
)

// The expiry date is usable through its end ("ใช้งานได้ถึงสิ้นวันที่กำหนด"): access ends at 00:00 of
// the NEXT day in the Holding's timezone; an empty or unknown
// timezone setting means Asia/Bangkok, and no date means no expiry.
func TestAccessExpiryInstant(t *testing.T) {
	stored := sql.NullTime{Valid: true, Time: time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)} // how lib/pq returns DATE
	for _, tc := range []struct {
		timezone string
		want     time.Time
	}{
		{"Asia/Bangkok", time.Date(2026, 12, 31, 17, 0, 0, 0, time.UTC)},
		{"", time.Date(2026, 12, 31, 17, 0, 0, 0, time.UTC)},
		{"  Not/AZone ", time.Date(2026, 12, 31, 17, 0, 0, 0, time.UTC)},
		{"Asia/Tokyo", time.Date(2026, 12, 31, 15, 0, 0, 0, time.UTC)},
		{"UTC", time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)},
	} {
		if got := AccessExpiryInstant(stored, tc.timezone); !got.Equal(tc.want) {
			t.Errorf("timezone %q: expiry instant %s, want %s", tc.timezone, got.UTC(), tc.want)
		}
	}
	if got := AccessExpiryInstant(sql.NullTime{}, "Asia/Bangkok"); !got.IsZero() {
		t.Errorf("no expiry date gave %s, want zero (no expiry)", got)
	}
}
