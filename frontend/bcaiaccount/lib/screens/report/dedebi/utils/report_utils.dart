import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/model/bi_report/bi_sale_report_data.dart';
import 'package:smlaicloud/model/global_model.dart';
import '../../../../global.dart' as global;

class ReportUtils {
  // Date formatting
  static String formatDate(String dateStr) {
    try {
      DateTime date = DateTime.parse(dateStr);
      return DateFormat('dd/MM/yyyy').format(date);
    } catch (e) {
      return dateStr;
    }
  }

  // Currency formatting
  static String formatCurrency(double amount) {
    return NumberFormat('#,##0.00').format(amount);
  }

  // Number formatting
  static String formatNumber(double number) {
    return NumberFormat('#,##0.##').format(number);
  }

  // Status colors and text
  static Color getStatusColor(String inquiryType) {
    switch (inquiryType) {
      case '0':
        return global.theme.positiveHighlightColor;
      case '1':
        return global.theme.negativeHighlightColor;
      default:
        return global.theme.surfaceColor;
    }
  }

  static Color getStatusTextColor(String inquiryType) {
    switch (inquiryType) {
      case '0':
        return global.theme.positiveHighlightTextColor;
      case '1':
        return global.theme.negativeHighlightTextColor;
      default:
        return global.theme.iconColor;
    }
  }

  static String getStatusText(String inquiryType) {
    switch (inquiryType) {
      case '0':
        return global.language('inquiry_credit');
      case '1':
        return global.language('cash');
      default:
        return global.language('others');
    }
  }

  // Name extraction
  static String getCreditorName(List<SaleCreditorName> creditorNames) {
    if (creditorNames.isEmpty) return '-';
    final thName = creditorNames.firstWhere(
      (name) => name.code == 'th',
      orElse: () => creditorNames.isNotEmpty
          ? creditorNames.first
          : SaleCreditorName(
              code: '', name: '-', isauto: false, isdelete: false),
    );
    return thName.name;
  }

  static String getDisplayName(List<LanguageDataModel> names) {
    if (names.isEmpty) return '-';
    final thName = names.firstWhere(
      (name) => name.code == 'th',
      orElse: () => names.isNotEmpty
          ? names.first
          : LanguageDataModel(code: '', name: '-'),
    );
    return thName.name;
  }

  static String getItemName(List<ItemName> itemNames) {
    if (itemNames.isEmpty) return '-';
    final thName = itemNames.firstWhere(
      (name) => name.code == 'th',
      orElse: () => itemNames.isNotEmpty
          ? itemNames.first
          : ItemName(code: '', name: '-', isauto: false, isdelete: false),
    );
    return thName.name;
  }

  static String getUnitName(List<UnitName> unitNames) {
    if (unitNames.isEmpty) return '-';
    final thName = unitNames.firstWhere(
      (name) => name.code == 'th',
      orElse: () => unitNames.isNotEmpty
          ? unitNames.first
          : UnitName(code: '', name: '-', isauto: false, isdelete: false),
    );
    return thName.name;
  }

  // เพิ่ม methods ที่รองรับ nullable parameters
  static String getDisplayNameSafe(List<LanguageDataModel>? names) {
    if (names == null || names.isEmpty) return '-';

    try {
      final thName = names.firstWhere(
        (name) => name.code == 'th',
        orElse: () => names.isNotEmpty
            ? names.first
            : LanguageDataModel(code: '', name: '-'),
      );
      return thName.name.isNotEmpty ? thName.name : '-';
    } catch (e) {
      return '-';
    }
  }

  static String getCreditorNameSafe(List<SaleCreditorName>? creditorNames) {
    if (creditorNames == null || creditorNames.isEmpty) return '-';

    try {
      final thName = creditorNames.firstWhere(
        (name) => name.code == 'th',
        orElse: () => creditorNames.isNotEmpty
            ? creditorNames.first
            : SaleCreditorName(
                code: '', name: '-', isauto: false, isdelete: false),
      );
      return thName.name.isNotEmpty ? thName.name : '-';
    } catch (e) {
      return '-';
    }
  }

// เพิ่มสำหรับ Item และ Unit ที่ใช้ใน TransactionDetailDialog
  static String getItemNameSafe(List<ItemName>? itemNames) {
    if (itemNames == null || itemNames.isEmpty) return '-';

    try {
      final thName = itemNames.firstWhere(
        (name) => name.code == 'th',
        orElse: () => itemNames.isNotEmpty
            ? itemNames.first
            : ItemName(code: '', name: '-', isauto: false, isdelete: false),
      );
      return thName.name.isNotEmpty ? thName.name : '-';
    } catch (e) {
      return '-';
    }
  }

  static String getUnitNameSafe(List<UnitName>? unitNames) {
    if (unitNames == null || unitNames.isEmpty) return '-';

    try {
      final thName = unitNames.firstWhere(
        (name) => name.code == 'th',
        orElse: () => unitNames.isNotEmpty
            ? unitNames.first
            : UnitName(code: '', name: '-', isauto: false, isdelete: false),
      );
      return thName.name.isNotEmpty ? thName.name : '-';
    } catch (e) {
      return '-';
    }
  }

  // เพิ่ม method ที่รองรับ nullable amount
  static String formatCurrencySafe(double? amount) {
    if (amount == null) return '0.00';
    return NumberFormat('#,##0.00').format(amount);
  }
}
