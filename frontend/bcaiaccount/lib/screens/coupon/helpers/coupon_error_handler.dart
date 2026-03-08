import 'package:flutter/material.dart';
import '../../../global.dart' as global;

/// Error handler for coupon-related operations
/// Converts technical errors into user-friendly messages
class CouponErrorHandler {
  /// Handle errors from coupon operations and show user-friendly messages
  static void handleError(
    BuildContext context,
    dynamic error, {
    String? operation,
  }) {
    String userMessage;
    Color messageColor;
    IconData icon;

    // Determine user-friendly message based on error type
    if (error is NetworkException) {
      userMessage = global.language('coupon_error_network');
      icon = Icons.wifi_off;
      messageColor = Colors.orange;
    } else if (error is ValidationException) {
      userMessage = error.message ?? global.language('coupon_error_validation');
      icon = Icons.error_outline;
      messageColor = Colors.orange;
    } else if (error is DuplicateException) {
      userMessage = global.language('coupon_error_duplicate');
      icon = Icons.content_copy;
      messageColor = Colors.orange;
    } else if (error is NotFoundException) {
      userMessage = global.language('coupon_error_not_found');
      icon = Icons.search_off;
      messageColor = Colors.orange;
    } else if (error is PermissionException) {
      userMessage = global.language('coupon_error_permission');
      icon = Icons.lock;
      messageColor = Colors.red;
    } else if (error is TimeoutException) {
      userMessage = global.language('coupon_error_timeout');
      icon = Icons.hourglass_empty;
      messageColor = Colors.orange;
    } else if (error is ServerException) {
      userMessage = global.language('coupon_error_server');
      icon = Icons.cloud_off;
      messageColor = Colors.red;
    } else {
      // Generic error
      userMessage = _getGenericErrorMessage(operation);
      icon = Icons.error;
      messageColor = Colors.red;
    }

    // Log technical error for debugging (can be sent to logging service)
    _logError(error, operation);

    // Show user-friendly message
    global.showSnackBar(
      context,
      Icon(icon, color: Colors.white),
      userMessage,
      messageColor,
    );
  }

  /// Get generic error message based on operation type
  static String _getGenericErrorMessage(String? operation) {
    switch (operation) {
      case 'create':
        return global.language('coupon_error_create_failed');
      case 'update':
        return global.language('coupon_error_update_failed');
      case 'delete':
        return global.language('coupon_error_delete_failed');
      case 'load':
        return global.language('coupon_error_load_failed');
      case 'search':
        return global.language('coupon_error_search_failed');
      default:
        return global.language('coupon_error_general');
    }
  }

  /// Log error for debugging purposes
  /// In production, this could send to a logging service like Firebase Crashlytics
  static void _logError(dynamic error, String? operation) {
    // For now, just print to console
    // In production, use proper logging service
    debugPrint('=== Coupon Error ===');
    debugPrint('Operation: ${operation ?? 'unknown'}');
    debugPrint('Error Type: ${error.runtimeType}');
    debugPrint('Error Message: $error');
    debugPrint('Stack Trace: ${StackTrace.current}');
    debugPrint('==================');

    // TODO: Send to Firebase Crashlytics or other logging service
    // FirebaseCrashlytics.instance.recordError(error, StackTrace.current);
  }

  /// Validate and handle product condition errors
  static String? validateProductConditionOperation(
    String operation,
    String? code,
  ) {
    if (code == null || code.trim().isEmpty) {
      switch (operation) {
        case 'add':
          return global.language('coupon_error_no_product_selected');
        case 'remove':
          return global.language('coupon_error_no_product_to_remove');
        default:
          return global.language('coupon_error_invalid_product');
      }
    }
    return null;
  }

  /// Validate customer selection errors
  static String? validateCustomerOperation(
    String operation,
    String? customerCode,
  ) {
    if (customerCode == null || customerCode.trim().isEmpty) {
      switch (operation) {
        case 'add':
          return global.language('coupon_error_no_customer_selected');
        case 'remove':
          return global.language('coupon_error_no_customer_to_remove');
        default:
          return global.language('coupon_error_invalid_customer');
      }
    }
    return null;
  }

  /// Show success message for operations
  static void showSuccess(
    BuildContext context,
    String operation, {
    String? details,
  }) {
    String message;
    IconData icon;

    switch (operation) {
      case 'create':
        message = global.language('coupon_success_created');
        icon = Icons.check_circle;
        break;
      case 'update':
        message = global.language('coupon_success_updated');
        icon = Icons.check_circle;
        break;
      case 'delete':
        message = global.language('coupon_success_deleted');
        icon = Icons.delete_sweep;
        break;
      case 'add_product':
        message = global.language('coupon_success_product_added');
        icon = Icons.add_shopping_cart;
        break;
      case 'remove_product':
        message = global.language('coupon_success_product_removed');
        icon = Icons.remove_shopping_cart;
        break;
      case 'add_customer':
        message = global.language('coupon_success_customer_added');
        icon = Icons.person_add;
        break;
      case 'remove_customer':
        message = global.language('coupon_success_customer_removed');
        icon = Icons.person_remove;
        break;
      default:
        message = global.language('coupon_success_operation');
        icon = Icons.check;
    }

    if (details != null && details.isNotEmpty) {
      message = '$message: $details';
    }

    global.showSnackBar(
      context,
      Icon(icon, color: Colors.white),
      message,
      Colors.green,
    );
  }

  /// Show warning message
  static void showWarning(BuildContext context, String message) {
    global.showSnackBar(
      context,
      const Icon(Icons.warning, color: Colors.white),
      message,
      Colors.orange,
    );
  }

  /// Show info message
  static void showInfo(BuildContext context, String message) {
    global.showSnackBar(
      context,
      const Icon(Icons.info, color: Colors.white),
      message,
      Colors.blue,
    );
  }
}

// Custom exception classes for better error handling

class NetworkException implements Exception {
  final String? message;
  NetworkException([this.message]);

  @override
  String toString() => message ?? 'Network error occurred';
}

class ValidationException implements Exception {
  final String? message;
  ValidationException([this.message]);

  @override
  String toString() => message ?? 'Validation error occurred';
}

class DuplicateException implements Exception {
  final String? message;
  DuplicateException([this.message]);

  @override
  String toString() => message ?? 'Duplicate entry detected';
}

class NotFoundException implements Exception {
  final String? message;
  NotFoundException([this.message]);

  @override
  String toString() => message ?? 'Resource not found';
}

class PermissionException implements Exception {
  final String? message;
  PermissionException([this.message]);

  @override
  String toString() => message ?? 'Permission denied';
}

class TimeoutException implements Exception {
  final String? message;
  TimeoutException([this.message]);

  @override
  String toString() => message ?? 'Operation timed out';
}

class ServerException implements Exception {
  final String? message;
  ServerException([this.message]);

  @override
  String toString() => message ?? 'Server error occurred';
}
