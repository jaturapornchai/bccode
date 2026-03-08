/// DateTime extension methods for coupon screen
/// Provides utilities for converting between local and UTC times
extension DateTimeExtensions on DateTime {
  /// Convert local DateTime to UTC ISO string
  /// Used for saving dates to backend
  String toUtcIsoString() {
    return toUtc().toIso8601String();
  }

  /// Create DateTime at start of day (00:00:00)
  DateTime get startOfDay {
    return DateTime(year, month, day, 0, 0, 0);
  }

  /// Create DateTime at end of day (23:59:59)
  DateTime get endOfDay {
    return DateTime(year, month, day, 23, 59, 59);
  }

  /// Check if this date is before another date (ignoring time)
  bool isBeforeDate(DateTime other) {
    return startOfDay.isBefore(other.startOfDay);
  }

  /// Check if this date is after another date (ignoring time)
  bool isAfterDate(DateTime other) {
    return startOfDay.isAfter(other.startOfDay);
  }

  /// Check if this date is same as another date (ignoring time)
  bool isSameDate(DateTime other) {
    return year == other.year && month == other.month && day == other.day;
  }

  /// Get first day of current month
  DateTime get firstDayOfMonth {
    return DateTime(year, month, 1);
  }

  /// Get last day of current month
  DateTime get lastDayOfMonth {
    return DateTime(year, month + 1, 0, 23, 59, 59);
  }
}

/// String extension for parsing ISO date strings
extension DateTimeStringExtensions on String {
  /// Parse ISO string to local DateTime
  /// Returns current DateTime if parsing fails
  DateTime parseLocalDate() {
    if (isEmpty) {
      return DateTime.now();
    }
    try {
      final utcDate = DateTime.parse(this);
      return utcDate.toLocal();
    } catch (e) {
      return DateTime.now();
    }
  }

  /// Parse ISO string to local DateTime or return null if invalid
  DateTime? tryParseLocalDate() {
    if (isEmpty) {
      return null;
    }
    try {
      final utcDate = DateTime.parse(this);
      return utcDate.toLocal();
    } catch (e) {
      return null;
    }
  }
}

/// Helper class for common date operations in coupon context
class CouponDateHelper {
  /// Get default issued date (first day of current month)
  static DateTime getDefaultIssuedDate() {
    final now = DateTime.now();
    return DateTime(now.year, now.month, 1);
  }

  /// Get default expiry date (last day of current month at 23:59:59)
  static DateTime getDefaultExpiryDate() {
    final now = DateTime.now();
    return DateTime(now.year, now.month + 1, 0, 23, 59, 59);
  }

  /// Validate date range for coupon
  /// Returns error message if invalid, null if valid
  static String? validateDateRange(DateTime? issuedDate, DateTime? expiryDate) {
    if (issuedDate == null || expiryDate == null) {
      return null; // Will be caught by required field validation
    }

    if (expiryDate.isBefore(issuedDate)) {
      return 'Expiry date must be after issued date';
    }

    return null;
  }

  /// Format date for display (e.g., "31 Oct 2025")
  static String formatDateForDisplay(DateTime? date) {
    if (date == null) return '-';

    const months = [
      'Jan',
      'Feb',
      'Mar',
      'Apr',
      'May',
      'Jun',
      'Jul',
      'Aug',
      'Sep',
      'Oct',
      'Nov',
      'Dec',
    ];

    return '${date.day} ${months[date.month - 1]} ${date.year}';
  }

  /// Format date range for display
  static String formatDateRange(DateTime? start, DateTime? end) {
    if (start == null || end == null) return '-';
    return '${formatDateForDisplay(start)} - ${formatDateForDisplay(end)}';
  }

  /// Check if coupon is currently active based on dates
  static bool isActive(DateTime? issuedDate, DateTime? expiryDate) {
    if (issuedDate == null || expiryDate == null) return false;

    final now = DateTime.now();
    return now.isAfter(issuedDate) && now.isBefore(expiryDate);
  }

  /// Check if coupon is expired
  static bool isExpired(DateTime? expiryDate) {
    if (expiryDate == null) return false;
    return DateTime.now().isAfter(expiryDate);
  }

  /// Check if coupon is not yet active
  static bool isNotYetActive(DateTime? issuedDate) {
    if (issuedDate == null) return false;
    return DateTime.now().isBefore(issuedDate);
  }

  /// Get number of days until expiry
  static int? daysUntilExpiry(DateTime? expiryDate) {
    if (expiryDate == null) return null;
    final now = DateTime.now();
    if (now.isAfter(expiryDate)) return 0;
    return expiryDate.difference(now).inDays;
  }

  /// Get number of days since issued
  static int? daysSinceIssued(DateTime? issuedDate) {
    if (issuedDate == null) return null;
    final now = DateTime.now();
    if (now.isBefore(issuedDate)) return 0;
    return now.difference(issuedDate).inDays;
  }
}
