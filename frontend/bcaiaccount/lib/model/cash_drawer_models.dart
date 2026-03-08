import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import '../global.dart' as global;

part 'cash_drawer_models.g.dart';

/// Enum for cash drawer transaction types
enum CashDrawerTransactionType {
  @JsonValue(1)
  openShift,
  @JsonValue(2)
  closeShift,
  @JsonValue(3)
  addCash,
  @JsonValue(4)
  withdrawCash,
}

/// Extension for CashDrawerTransactionType
extension CashDrawerTransactionTypeExtension on CashDrawerTransactionType {
  String get displayName {
    switch (this) {
      case CashDrawerTransactionType.openShift:
        return global.language('open_shift');
      case CashDrawerTransactionType.closeShift:
        return global.language('close_shift');
      case CashDrawerTransactionType.addCash:
        return global.language('add_cash_to_drawer');
      case CashDrawerTransactionType.withdrawCash:
        return global.language('withdraw_cash');
    }
  }

  int get value {
    switch (this) {
      case CashDrawerTransactionType.openShift:
        return 1;
      case CashDrawerTransactionType.closeShift:
        return 2;
      case CashDrawerTransactionType.addCash:
        return 3;
      case CashDrawerTransactionType.withdrawCash:
        return 4;
    }
  }

  bool get isCashIn =>
      this == CashDrawerTransactionType.openShift ||
      this == CashDrawerTransactionType.addCash;
  bool get isCashOut => this == CashDrawerTransactionType.withdrawCash;
  bool get isShiftOperation =>
      this == CashDrawerTransactionType.openShift ||
      this == CashDrawerTransactionType.closeShift;
}

/// Validation result class
@JsonSerializable()
class ValidationResult {
  final bool isValid;
  final List<String> errors;
  final List<String> warnings;

  const ValidationResult({
    this.isValid = true,
    this.errors = const [],
    this.warnings = const [],
  });

  factory ValidationResult.success() => const ValidationResult();

  factory ValidationResult.failure(
    List<String> errors, {
    List<String> warnings = const [],
  }) => ValidationResult(isValid: false, errors: errors, warnings: warnings);

  factory ValidationResult.fromJson(Map<String, dynamic> json) =>
      _$ValidationResultFromJson(json);
  Map<String, dynamic> toJson() => _$ValidationResultToJson(this);
}

/// Enhanced Cash Drawer Transaction Model with validation
@JsonSerializable(explicitToJson: true)
class CashDrawerTransaction {
  final String? guidfixed;
  final String usercode;
  final String username;
  final String posid;
  final String docno;
  final CashDrawerTransactionType doctype;
  final DateTime docdate;
  final String? remark;
  final double amount;
  final PaymentBreakdown paymentBreakdown;
  final List<TransactionModel>? billDetails;

  const CashDrawerTransaction({
    this.guidfixed,
    required this.usercode,
    required this.username,
    required this.posid,
    required this.docno,
    required this.doctype,
    required this.docdate,
    this.remark,
    required this.amount,
    required this.paymentBreakdown,
    this.billDetails,
  });

  /// Validation method
  ValidationResult validate() {
    final errors = <String>[];
    final warnings = <String>[];

    // Required field validations
    if (usercode.trim().isEmpty) {
      errors.add(global.language('user_code_cannot_be_empty'));
    }

    if (username.trim().isEmpty) {
      errors.add('ชื่อผู้ใช้ไม่สามารถเป็นค่าว่างได้');
    }

    if (posid.trim().isEmpty) {
      errors.add(global.language('pos_id_cannot_be_empty'));
    }

    if (docno.trim().isEmpty) {
      errors.add(global.language('doc_no_cannot_be_empty'));
    }

    // Amount validations
    if (amount < 0) {
      errors.add('จำนวนเงินไม่สามารถเป็นค่าติดลบได้');
    }

    if (doctype.isCashOut && amount == 0) {
      errors.add('การถอนเงินต้องมีจำนวนเงินมากกว่า 0');
    }

    // Date validations
    final now = DateTime.now();
    if (docdate.isAfter(now.add(const Duration(hours: 1)))) {
      errors.add('วันที่เอกสารไม่สามารถเป็นอนาคตได้');
    }

    // Payment breakdown validation
    final paymentValidation = paymentBreakdown.validate();
    if (!paymentValidation.isValid) {
      errors.addAll(paymentValidation.errors);
      warnings.addAll(paymentValidation.warnings);
    }

    // Business logic validations
    if (doctype == CashDrawerTransactionType.openShift && amount == 0) {
      warnings.add('การเปิดกะด้วยเงินเริ่มต้น 0 อาจไม่เหมาะสม');
    }

    // Payment type warnings
    if (paymentBreakdown.totalAmount != amount) {
      warnings.add('ยอดรวมการชำระเงินไม่ตรงกับจำนวนเงินรวม');
    }

    return ValidationResult(
      isValid: errors.isEmpty,
      errors: errors,
      warnings: warnings,
    );
  }

  /// Get formatted date string
  String get formattedDate {
    return '${docdate.day.toString().padLeft(2, '0')}/${docdate.month.toString().padLeft(2, '0')}/${(docdate.year + 543).toString()} ${docdate.hour.toString().padLeft(2, '0')}:${docdate.minute.toString().padLeft(2, '0')}';
  }

  /// Get display text for transaction type
  String get doctypeDisplayText => doctype.displayName;

  /// Check if this is a cash-in transaction
  bool get isCashIn => doctype.isCashIn;

  /// Check if this is a cash-out transaction
  bool get isCashOut => doctype.isCashOut;

  /// Check if this is a shift operation
  bool get isShiftOperation => doctype.isShiftOperation;

  factory CashDrawerTransaction.fromJson(Map<String, dynamic> json) =>
      _$CashDrawerTransactionFromJson(json);
  Map<String, dynamic> toJson() => _$CashDrawerTransactionToJson(this);

  /// Copy with method for creating modified copies
  CashDrawerTransaction copyWith({
    String? guidfixed,
    String? usercode,
    String? username,
    String? posid,
    String? docno,
    CashDrawerTransactionType? doctype,
    DateTime? docdate,
    String? remark,
    double? amount,
    PaymentBreakdown? paymentBreakdown,
    List<TransactionModel>? billDetails,
  }) {
    return CashDrawerTransaction(
      guidfixed: guidfixed ?? this.guidfixed,
      usercode: usercode ?? this.usercode,
      username: username ?? this.username,
      posid: posid ?? this.posid,
      docno: docno ?? this.docno,
      doctype: doctype ?? this.doctype,
      docdate: docdate ?? this.docdate,
      remark: remark ?? this.remark,
      amount: amount ?? this.amount,
      paymentBreakdown: paymentBreakdown ?? this.paymentBreakdown,
      billDetails: billDetails ?? this.billDetails,
    );
  }
}

/// Payment breakdown model
@JsonSerializable()
class PaymentBreakdown {
  final double cash;
  final double creditCard;
  final double promptPay;
  final double transfer;
  final double cheque;
  final double coupon;

  const PaymentBreakdown({
    this.cash = 0.0,
    this.creditCard = 0.0,
    this.promptPay = 0.0,
    this.transfer = 0.0,
    this.cheque = 0.0,
    this.coupon = 0.0,
  });

  double get totalAmount =>
      cash + creditCard + promptPay + transfer + cheque + coupon;

  ValidationResult validate() {
    final errors = <String>[];
    final warnings = <String>[];

    // Check for negative values
    if (cash < 0) errors.add('จำนวนเงินสดไม่สามารถเป็นค่าติดลบได้');
    if (creditCard < 0) {
      errors.add('จำนวนเงินบัตรเครดิตไม่สามารถเป็นค่าติดลบได้');
    }
    if (promptPay < 0) {
      errors.add('จำนวนเงิน PromptPay ไม่สามารถเป็นค่าติดลบได้');
    }
    if (transfer < 0) errors.add('จำนวนเงินโอนไม่สามารถเป็นค่าติดลบได้');
    if (cheque < 0) errors.add('จำนวนเงินเช็คไม่สามารถเป็นค่าติดลบได้');
    if (coupon < 0) errors.add('จำนวนเงินคูปองไม่สามารถเป็นค่าติดลบได้');

    // Business warnings
    if (totalAmount == 0) {
      warnings.add('ไม่มีการชำระเงินในรายการนี้');
    }

    return ValidationResult(
      isValid: errors.isEmpty,
      errors: errors,
      warnings: warnings,
    );
  }

  factory PaymentBreakdown.fromJson(Map<String, dynamic> json) =>
      _$PaymentBreakdownFromJson(json);
  Map<String, dynamic> toJson() => _$PaymentBreakdownToJson(this);
}

/// Shift summary model for grouping transactions by shift
@JsonSerializable(explicitToJson: true)
class ShiftSummary {
  final String docno;
  final String usercode;
  final String username;
  final String posid;
  final CashDrawerTransaction? openShift;
  final CashDrawerTransaction? closeShift;
  final List<CashDrawerTransaction> transactions;

  const ShiftSummary({
    required this.docno,
    required this.usercode,
    required this.username,
    required this.posid,
    this.openShift,
    this.closeShift,
    required this.transactions,
  });

  /// Check if shift is closed
  bool get isClosed => closeShift != null;

  /// Get start date from open shift
  DateTime? get startDate => openShift?.docdate;

  /// Get end date from close shift
  DateTime? get endDate => closeShift?.docdate;

  /// Get opening amount
  double get openAmount => openShift?.amount ?? 0.0;

  /// Get closing amount
  double get closeAmount => closeShift?.amount ?? 0.0;

  /// Get total cash added during shift
  double get addAmount => transactions
      .where((t) => t.doctype == CashDrawerTransactionType.addCash)
      .fold(0.0, (sum, t) => sum + t.amount);

  /// Get total cash withdrawn during shift
  double get withdrawAmount => transactions
      .where((t) => t.doctype == CashDrawerTransactionType.withdrawCash)
      .fold(0.0, (sum, t) => sum + t.amount);

  /// Get calculated closing balance
  double get calculatedBalance => openAmount + addAmount - withdrawAmount;

  /// Get actual vs calculated difference
  double get balanceDifference => closeAmount - calculatedBalance;

  /// Check if there's a balance discrepancy
  bool get hasBalanceDiscrepancy => isClosed && balanceDifference.abs() > 0.01;

  /// Validate the shift data
  ValidationResult validate() {
    final errors = <String>[];
    final warnings = <String>[];

    // Required fields
    if (docno.trim().isEmpty) {
      errors.add(global.language('doc_no_cannot_be_empty'));
    }

    if (usercode.trim().isEmpty) {
      errors.add(global.language('user_code_cannot_be_empty'));
    }

    if (posid.trim().isEmpty) {
      errors.add(global.language('pos_id_cannot_be_empty'));
    }

    // Business logic validations
    if (openShift == null) {
      errors.add('ไม่พบข้อมูลการเปิดกะ');
    }

    if (isClosed && hasBalanceDiscrepancy) {
      warnings.add(
        'ยอดปิดกะไม่ตรงกับการคำนวณ (ต่าง ${balanceDifference.toStringAsFixed(2)} บาท)',
      );
    }

    // Validate individual transactions
    for (final transaction in transactions) {
      final transactionValidation = transaction.validate();
      if (!transactionValidation.isValid) {
        errors.addAll(
          transactionValidation.errors.map(
            (e) => 'รายการ ${transaction.docno}: $e',
          ),
        );
      }
      warnings.addAll(
        transactionValidation.warnings.map(
          (w) => 'รายการ ${transaction.docno}: $w',
        ),
      );
    }

    return ValidationResult(
      isValid: errors.isEmpty,
      errors: errors,
      warnings: warnings,
    );
  }

  /// Get shift duration in hours
  double? get shiftDurationHours {
    if (startDate == null || endDate == null) return null;
    return endDate!.difference(startDate!).inMinutes / 60.0;
  }

  /// Get formatted shift time text
  String get shiftTimeText {
    if (startDate == null) return 'ไม่มีข้อมูลเวลา';

    final startFormatted =
        '${startDate!.day.toString().padLeft(2, '0')}/${startDate!.month.toString().padLeft(2, '0')}/${(startDate!.year + 543).toString()} ${startDate!.hour.toString().padLeft(2, '0')}:${startDate!.minute.toString().padLeft(2, '0')}';

    if (isClosed && endDate != null) {
      final endFormatted =
          '${endDate!.day.toString().padLeft(2, '0')}/${endDate!.month.toString().padLeft(2, '0')}/${(endDate!.year + 543).toString()} ${endDate!.hour.toString().padLeft(2, '0')}:${endDate!.minute.toString().padLeft(2, '0')}';
      return 'เปิดกะ: $startFormatted - ปิดกะ: $endFormatted';
    } else {
      return 'เปิดกะ: $startFormatted - ยังไม่ปิดกะ';
    }
  }

  factory ShiftSummary.fromJson(Map<String, dynamic> json) =>
      _$ShiftSummaryFromJson(json);
  Map<String, dynamic> toJson() => _$ShiftSummaryToJson(this);
}

/// Dashboard statistics for cash drawer operations
@JsonSerializable()
class CashDrawerDashboard {
  final int totalShifts;
  final int openShifts;
  final int closedShifts;
  final double totalAmount;
  final double totalCashIn;
  final double totalCashOut;
  final List<ValidationResult> validationIssues;

  const CashDrawerDashboard({
    this.totalShifts = 0,
    this.openShifts = 0,
    this.closedShifts = 0,
    this.totalAmount = 0.0,
    this.totalCashIn = 0.0,
    this.totalCashOut = 0.0,
    this.validationIssues = const [],
  });

  /// Check if there are any validation issues
  bool get hasValidationIssues => validationIssues.any((v) => !v.isValid);

  /// Get total number of errors across all validations
  int get totalErrors =>
      validationIssues.fold(0, (sum, v) => sum + v.errors.length);

  /// Get total number of warnings across all validations
  int get totalWarnings =>
      validationIssues.fold(0, (sum, v) => sum + v.warnings.length);

  factory CashDrawerDashboard.fromJson(Map<String, dynamic> json) =>
      _$CashDrawerDashboardFromJson(json);
  Map<String, dynamic> toJson() => _$CashDrawerDashboardToJson(this);
}
