import '../../../model/product_condition_model.dart';

/// Model to hold all coupon form state in one place
/// This reduces the number of individual state variables and makes state management clearer
class CouponFormState {
  // Basic form data
  final String couponCode;
  final List<String> couponNames; // One per language
  final double? couponValue;
  final int couponType; // 0 = amount, 1 = percentage
  final String remark;

  // Usage limits
  final int? maxUsageCount;
  final int? maxUsageCountPerCustomer;
  final bool isOneTimeUse;

  // Date range
  final DateTime? issuedDate;
  final DateTime? expiryDate;

  // Status
  final int status; // 0 = active, 1 = inactive

  // Customer restrictions
  final List<String> customerCodes;

  // Product conditions
  final ProductConditionModel? productCondition;

  // Validation errors
  final Map<String, String> fieldErrors;

  // UI state
  final bool isChanged;
  final bool isLoadingTranslation;

  const CouponFormState({
    this.couponCode = '',
    this.couponNames = const [],
    this.couponValue,
    this.couponType = 0,
    this.remark = '',
    this.maxUsageCount,
    this.maxUsageCountPerCustomer,
    this.isOneTimeUse = false,
    this.issuedDate,
    this.expiryDate,
    this.status = 0,
    this.customerCodes = const [],
    this.productCondition,
    this.fieldErrors = const {},
    this.isChanged = false,
    this.isLoadingTranslation = false,
  });

  CouponFormState copyWith({
    String? couponCode,
    List<String>? couponNames,
    double? couponValue,
    int? couponType,
    String? remark,
    int? maxUsageCount,
    int? maxUsageCountPerCustomer,
    bool? isOneTimeUse,
    DateTime? issuedDate,
    DateTime? expiryDate,
    int? status,
    List<String>? customerCodes,
    ProductConditionModel? productCondition,
    Map<String, String>? fieldErrors,
    bool? isChanged,
    bool? isLoadingTranslation,
  }) {
    return CouponFormState(
      couponCode: couponCode ?? this.couponCode,
      couponNames: couponNames ?? this.couponNames,
      couponValue: couponValue ?? this.couponValue,
      couponType: couponType ?? this.couponType,
      remark: remark ?? this.remark,
      maxUsageCount: maxUsageCount ?? this.maxUsageCount,
      maxUsageCountPerCustomer:
          maxUsageCountPerCustomer ?? this.maxUsageCountPerCustomer,
      isOneTimeUse: isOneTimeUse ?? this.isOneTimeUse,
      issuedDate: issuedDate ?? this.issuedDate,
      expiryDate: expiryDate ?? this.expiryDate,
      status: status ?? this.status,
      customerCodes: customerCodes ?? this.customerCodes,
      productCondition: productCondition ?? this.productCondition,
      fieldErrors: fieldErrors ?? this.fieldErrors,
      isChanged: isChanged ?? this.isChanged,
      isLoadingTranslation: isLoadingTranslation ?? this.isLoadingTranslation,
    );
  }

  /// Create initial/empty state
  factory CouponFormState.initial() {
    return CouponFormState(
      issuedDate: DateTime(DateTime.now().year, DateTime.now().month, 1),
      expiryDate: DateTime(
        DateTime.now().year,
        DateTime.now().month + 1,
        0,
        23,
        59,
        59,
      ),
    );
  }

  /// Clear all form data (for new coupon or after save)
  CouponFormState clear() {
    return CouponFormState.initial();
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;

    return other is CouponFormState &&
        other.couponCode == couponCode &&
        _listEquals(other.couponNames, couponNames) &&
        other.couponValue == couponValue &&
        other.couponType == couponType &&
        other.remark == remark &&
        other.maxUsageCount == maxUsageCount &&
        other.maxUsageCountPerCustomer == maxUsageCountPerCustomer &&
        other.isOneTimeUse == isOneTimeUse &&
        other.issuedDate == issuedDate &&
        other.expiryDate == expiryDate &&
        other.status == status &&
        _listEquals(other.customerCodes, customerCodes) &&
        other.productCondition == productCondition &&
        _mapEquals(other.fieldErrors, fieldErrors) &&
        other.isChanged == isChanged &&
        other.isLoadingTranslation == isLoadingTranslation;
  }

  @override
  int get hashCode {
    return Object.hash(
      couponCode,
      Object.hashAll(couponNames),
      couponValue,
      couponType,
      remark,
      maxUsageCount,
      maxUsageCountPerCustomer,
      isOneTimeUse,
      issuedDate,
      expiryDate,
      status,
      Object.hashAll(customerCodes),
      productCondition,
      Object.hashAll(fieldErrors.entries),
      isChanged,
      isLoadingTranslation,
    );
  }

  bool _listEquals<T>(List<T>? a, List<T>? b) {
    if (a == null) return b == null;
    if (b == null || a.length != b.length) return false;
    for (int index = 0; index < a.length; index += 1) {
      if (a[index] != b[index]) return false;
    }
    return true;
  }

  bool _mapEquals<K, V>(Map<K, V>? a, Map<K, V>? b) {
    if (a == null) return b == null;
    if (b == null || a.length != b.length) return false;
    for (final K key in a.keys) {
      if (!b.containsKey(key) || b[key] != a[key]) {
        return false;
      }
    }
    return true;
  }
}
