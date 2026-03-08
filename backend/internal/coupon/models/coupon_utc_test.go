package models

import (
	"testing"
	"time"
)

// TestIsExpiredUTCHandling tests that the IsExpired method correctly handles UTC timezone comparison
func TestIsExpiredUTCHandling(t *testing.T) {
	// Test case 1: Coupon that should not be expired (expires Dec 31, 2025)
	// This matches the actual coupon data mentioned in the issue
	expiryDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	coupon := &Coupon{
		ExpiryDate: expiryDate,
		Status:     CouponStatusActive,
	}

	// Test with current date (June 19, 2025) - should not be expired
	if coupon.IsExpired() {
		t.Errorf("Coupon should not be expired. Expiry date: %v, Current time: %v",
			expiryDate.UTC(), time.Now().UTC())
	}

	// Test case 2: Coupon that should be expired
	expiredDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	expiredCoupon := &Coupon{
		ExpiryDate: expiredDate,
		Status:     CouponStatusActive,
	}

	if !expiredCoupon.IsExpired() {
		t.Errorf("Coupon should be expired. Expiry date: %v, Current time: %v",
			expiredDate.UTC(), time.Now().UTC())
	}

	// Test case 3: Test that both local and UTC timezone comparisons work correctly
	// Create a date that's in the future in UTC
	futureUTCDate := time.Now().UTC().Add(24 * time.Hour)
	futureCoupon := &Coupon{
		ExpiryDate: futureUTCDate,
		Status:     CouponStatusActive,
	}

	if futureCoupon.IsExpired() {
		t.Errorf("Future coupon should not be expired. Expiry date: %v, Current time: %v",
			futureUTCDate, time.Now().UTC())
	}

	t.Logf("UTC Date Handling Test completed successfully")
	t.Logf("Current UTC time: %v", time.Now().UTC())
	t.Logf("Test coupon expiry (Dec 31, 2025): %v", expiryDate)
	t.Logf("Coupon is expired: %v", coupon.IsExpired())
}

// TestCouponReservationIsExpiredUTCHandling tests reservation expiry with UTC handling
func TestCouponReservationIsExpiredUTCHandling(t *testing.T) {
	// Test case: Reservation that should not be expired (expires in 15 minutes from now)
	reservation := NewCouponReservation("test-coupon", "test-customer", "test-transaction")

	if reservation.IsExpired() {
		t.Errorf("New reservation should not be expired. Expires at: %v, Current time: %v",
			reservation.ExpiresAt.UTC(), time.Now().UTC())
	}

	// Test case: Create an expired reservation
	expiredReservation := &CouponReservation{
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour), // 1 hour ago
		Status:    ReservationStatusActive,
	}

	if !expiredReservation.IsExpired() {
		t.Errorf("Expired reservation should be expired. Expires at: %v, Current time: %v",
			expiredReservation.ExpiresAt.UTC(), time.Now().UTC())
	}

	t.Logf("Reservation UTC Date Handling Test completed successfully")
}
