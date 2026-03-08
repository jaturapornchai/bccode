import 'package:flutter/material.dart';
import '../../../global.dart' as global;

/// Validator class for coupon form data
/// Centralizes all validation logic for better maintainability
class CouponValidator {
  /// Validates all coupon form fields
  /// Returns a map of field errors (empty if validation passes)
  static Map<String, String> validate({
    required List<TextEditingController> fieldTextController,
    required TextEditingController couponValueController,
    required TextEditingController maxUsageCountController,
    required DateTime? issuedDate,
    required DateTime? expiryDate,
    required int couponType,
  }) {
    final Map<String, String> errors = {};

    // 1. Validate coupon code (required)
    if (fieldTextController[0].text.trim().isEmpty) {
      errors['coupon_code'] = global.language('coupon_error_code_required');
    }

    // 2. Validate coupon value (required)
    if (couponValueController.text.trim().isEmpty) {
      errors['coupon_value'] = global.language('coupon_error_value_required');
    } else {
      // Validate coupon value is a number and greater than 0
      final double? couponValue = double.tryParse(couponValueController.text);
      if (couponValue == null || couponValue <= 0) {
        errors['coupon_value'] = global.language('coupon_error_value_invalid');
      } else {
        // If percentage type, must not exceed 100
        if (couponType == 1 && couponValue > 100) {
          errors['coupon_value'] = global.language(
            'coupon_error_percentage_max',
          );
        }
      }
    }

    // 3. Validate max usage count (required)
    if (maxUsageCountController.text.trim().isEmpty) {
      errors['max_usage_count'] = global.language(
        'coupon_error_max_usage_required',
      );
    } else {
      // Validate max usage count is a number and greater than 0
      final int? maxUsageCount = int.tryParse(maxUsageCountController.text);
      if (maxUsageCount == null || maxUsageCount <= 0) {
        errors['max_usage_count'] = global.language(
          'coupon_error_max_usage_invalid',
        );
      }
    }

    // 4. Validate issued date (required)
    if (issuedDate == null) {
      errors['issued_date'] = global.language(
        'coupon_error_issued_date_required',
      );
    }

    // 5. Validate expiry date (required)
    if (expiryDate == null) {
      errors['expiry_date'] = global.language(
        'coupon_error_expiry_date_required',
      );
    }

    // 6. Validate date range (expiry must be after issued)
    if (issuedDate != null && expiryDate != null) {
      if (expiryDate.isBefore(issuedDate)) {
        errors['expiry_date'] = global.language(
          'coupon_error_expiry_before_issued',
        );
      }
    }

    // 7. Validate coupon name (at least one language required)
    bool hasAtLeastOneName = false;
    for (int i = 1; i < fieldTextController.length; i++) {
      if (fieldTextController[i].text.trim().isNotEmpty) {
        hasAtLeastOneName = true;
        break;
      }
    }
    if (!hasAtLeastOneName) {
      errors['coupon_name'] = global.language('coupon_error_name_required');
    }

    return errors;
  }

  /// Parse and validate coupon value
  /// Returns null if invalid
  static double? parseCouponValue(String text) {
    final trimmed = text.trim();
    if (trimmed.isEmpty) return null;

    final value = double.tryParse(trimmed);
    if (value == null || value <= 0) return null;

    return value;
  }

  /// Parse and validate max usage count
  /// Returns null if invalid
  static int? parseMaxUsageCount(String text) {
    final trimmed = text.trim();
    if (trimmed.isEmpty) return null;

    final value = int.tryParse(trimmed);
    if (value == null || value <= 0) return null;

    return value;
  }

  /// Parse and validate max usage count per customer (optional field)
  /// Returns null if empty or invalid
  static int? parseMaxUsageCountPerCustomer(String text) {
    final trimmed = text.trim();
    if (trimmed.isEmpty) return null;

    return int.tryParse(trimmed);
  }
}
