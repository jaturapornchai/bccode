package models

import (
	"strings"
	"time"
	_ "time/tzdata" // Holding timezones resolve the same in every image, even without zoneinfo
)

// defaultHoldingTimezone is used when a Holding has no (or an unknown) timezone setting.
const defaultHoldingTimezone = "Asia/Bangkok"

// HoldingLocation is the Holding's timezone (holdings.profile settings.timezone), falling
// back to Asia/Bangkok when the setting is empty or not a known zone.
// Lives here (a leaf package) so organization/access, centraldb, shop and authentication
// share one rule without an import cycle (centraldb → mcptoken → organization/access).
func HoldingLocation(timezone string) *time.Location {
	if name := strings.TrimSpace(timezone); name != "" {
		if location, err := time.LoadLocation(name); err == nil {
			return location
		}
	}
	if location, err := time.LoadLocation(defaultHoldingTimezone); err == nil {
		return location
	}
	return time.FixedZone("ICT", 7*60*60)
}

// AccessEndsAt is the moment a membership's access ends. The expiry date means
// "ใช้งานได้ถึงสิ้นวันที่กำหนด": the whole expiry date is still usable in the Holding's
// timezone, so access ends at 00:00 of the NEXT day there. expiryDate carries only the
// calendar date (lib/pq returns DATE as 00:00 UTC); zero = no expiry.
// Every place that resolves a membership (login, Holding selection, organization/access,
// role permissions, API/MCP tokens, GL scope) must use this one rule via AccessExpired.
func AccessEndsAt(expiryDate time.Time, timezone string) time.Time {
	if expiryDate.IsZero() {
		return time.Time{}
	}
	return time.Date(expiryDate.Year(), expiryDate.Month(), expiryDate.Day()+1, 0, 0, 0, 0, HoldingLocation(timezone))
}

// AccessExpired reports whether now has reached endsAt (AccessEndsAt); zero endsAt never expires.
func AccessExpired(endsAt, now time.Time) bool {
	return !endsAt.IsZero() && !now.Before(endsAt)
}
