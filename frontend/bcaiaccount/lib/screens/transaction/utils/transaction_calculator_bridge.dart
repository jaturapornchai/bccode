import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_calculator.dart';
// ignore: library_prefixes
import 'package:smlaicloud/services/transaction_calculator_service.dart'
    as calc_service;

/// TransactionCalculatorBridge - Local-first + Backend validation
///
/// Strategy:
/// 1. calTotalValue() → รัน local calculator ก่อน (synchronous, UI อัพเดททันที)
/// 2. fire async → ส่งข้อมูลไป backend /api/transaction/calculate
/// 3. ถ้า backend return ผลต่าง → log warning (สำหรับ debug)
/// 4. ก่อน save → เรียก backend verifyPayment()
///
/// เหตุผล: calTotalValue() ถูกเรียก synchronous จาก 20+ จุด
/// เปลี่ยนเป็น async ทั้งหมดเสี่ยงเกินไป
class TransactionCalculatorBridge extends TransactionCalculator {
  TransactionCalculatorBridge(super.state);

  /// Flag to enable/disable async backend validation
  bool enableBackendValidation = true;

  /// Last backend result for comparison
  calc_service.TransactionCalculationResult? _lastBackendResult;

  /// Whether backend validation is in progress
  bool _isValidating = false;

  /// คำนวณยอดรวมทั้งหมดของเอกสาร
  /// Local-first: รัน local calculator ทันที, แล้ว fire async ไป backend
  @override
  void calTotalValue() {
    // Step 1: Local calculation (synchronous - UI updates immediately)
    super.calTotalValue();

    // Step 2: Fire async backend validation (non-blocking)
    if (enableBackendValidation) {
      _validateWithBackend();
    }
  }

  /// Fire-and-forget backend validation
  Future<void> _validateWithBackend() async {
    if (_isValidating) return; // ไม่ซ้อน request
    _isValidating = true;

    try {
      final screenData = state.screenData as TransactionModel;

      // สร้าง items จาก details
      final items = (screenData.details ?? []).map((detail) {
        return calc_service.TransactionItem(
          itemCode: detail.itemcode,
          itemName: detail.itemcode,
          qty: detail.qty,
          price: detail.price,
          discount: detail.discount,
          vatCal: detail.vatcal ?? 0,
          unitCode: detail.unitcode,
        );
      }).toList();

      if (items.isEmpty) return;

      // ดึง discount text จาก controllers
      String detailDiscount = "0";
      String discountWord = "0";
      try {
        detailDiscount = state.detaildiscountformulaController.text;
        discountWord = state.discountWordController.text;
      } catch (_) {
        // PO screen อาจไม่มี controllers เหล่านี้
      }

      // Determine transaction type
      String? transactionType;
      try {
        final txType = state.transactionType;
        if (txType == global.TransactionTypeEnum.salereturn) {
          transactionType = 'salereturn';
        } else if (txType == global.TransactionTypeEnum.purchasereturn) {
          transactionType = 'purchasereturn';
        }
      } catch (_) {}

      final result = await calc_service.TransactionCalculatorService.calculate(
        items: items,
        discount: discountWord.isNotEmpty ? discountWord : "0",
        detailDiscount: detailDiscount.isNotEmpty ? detailDiscount : "0",
        vatType: screenData.vattype,
        vatRate: screenData.vatrate,
        depositAmount: screenData.sumdeposit ?? 0,
        isManualAmount: screenData.ismanualamount,
        inquiryType: screenData.inquirytype,
        docCurrency: screenData.docCurrency,
        baseCurrency: screenData.currency,
        exchangeRate: screenData.exchangerate,
        transactionType: transactionType,
        refTotalOriginal: screenData.reftotaloriginal,
      );

      _lastBackendResult = result;

      if (result.isSuccess) {
        // เปรียบเทียบผลลัพธ์กับ local calculation
        _compareResults(screenData, result);
      } else {
        debugPrint(
            '[CalcBridge] Backend validation failed: ${result.errorMessage}');
      }
    } catch (e) {
      debugPrint('[CalcBridge] Backend validation error: $e');
    } finally {
      _isValidating = false;
    }
  }

  /// เปรียบเทียบผลลัพธ์ local vs backend
  void _compareResults(
    TransactionModel localData,
    calc_service.TransactionCalculationResult backendResult,
  ) {
    const tolerance = 0.02; // 2 satang tolerance

    final diffs = <String>[];

    if (_diff(localData.totalbeforevat, backendResult.totalBeforeVat) >
        tolerance) {
      diffs.add(
          'totalbeforevat: local=${localData.totalbeforevat} vs backend=${backendResult.totalBeforeVat}');
    }
    if (_diff(localData.totalvatvalue, backendResult.totalVatValue) >
        tolerance) {
      diffs.add(
          'totalvatvalue: local=${localData.totalvatvalue} vs backend=${backendResult.totalVatValue}');
    }
    if (_diff(localData.totalamount, backendResult.totalAmount) > tolerance) {
      diffs.add(
          'totalamount: local=${localData.totalamount} vs backend=${backendResult.totalAmount}');
    }
    if (_diff(localData.totalexceptvat, backendResult.totalExceptVat) >
        tolerance) {
      diffs.add(
          'totalexceptvat: local=${localData.totalexceptvat} vs backend=${backendResult.totalExceptVat}');
    }

    if (diffs.isNotEmpty) {
      debugPrint('[CalcBridge] WARNING: Local vs Backend mismatch:');
      for (final diff in diffs) {
        debugPrint('  $diff');
      }
    }
  }

  double _diff(double a, double b) => (a - b).abs();

  /// ดึงผลลัพธ์ backend ล่าสุด (สำหรับ debug/testing)
  calc_service.TransactionCalculationResult? get lastBackendResult =>
      _lastBackendResult;
}
