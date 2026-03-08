import 'package:flutter/material.dart';
import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/services/currency_api_service.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_currency_utils.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/global.dart' as global;

/// Handler สำหรับจัดการ Multi-Currency ใน PO
///
/// รับผิดชอบ:
/// - โหลดรายการสกุลเงิน
/// - จัดการการเลือกและเปลี่ยนสกุลเงิน
/// - ตรวจสอบและอัพเดทอัตราแลกเปลี่ยน
class POCurrencyHandler {
  /// สกุลเงินหลัก (Base Currency)
  String? baseCurrency;
  String? baseCurrencySymbol;

  /// สกุลเงินเอกสาร (Document Currency)
  String? docCurrencyCode;
  String? docCurrencySymbol;
  double exchangeRate = 1.0;

  /// รายการสกุลเงินที่มี
  List<CurrencyModel> currencies = [];

  /// Callback เมื่อข้อมูลเปลี่ยน
  final VoidCallback onStateChanged;
  final void Function(String message) onWarning;
  final void Function() onRecalculate;

  POCurrencyHandler({
    required this.onStateChanged,
    required this.onWarning,
    required this.onRecalculate,
  });

  /// ตรวจสอบว่าเป็น Multi-Currency หรือไม่
  bool get isMultiCurrency =>
    docCurrencyCode != null &&
    baseCurrency != null &&
    docCurrencyCode != baseCurrency;

  /// ตรวจสอบว่าควรแสดง Currency Section หรือไม่
  bool get shouldShowCurrencySection => currencies.length > 1;

  /// โหลดรายการสกุลเงิน
  Future<void> loadCurrencies(TransactionModel screenData) async {
    AppLogger.debug('[POCurrencyHandler] loadCurrencies เริ่ม — docCurrency: ${screenData.docCurrency}, exchangerate: ${screenData.exchangerate}');
    try {
      final apiService = CurrencyApiService();
      final shopId = global.prefs.getString("shopid") ?? "";

      final response = await apiService.getCurrencies(
        shopId: shopId,
        page: 1,
        limit: 1000,
      );

      AppLogger.debug('[POCurrencyHandler] API response — success: ${response.success}, data!=null: ${response.data != null}');

      if (response.success && response.data != null) {
        final loadedCurrencies = (response.data as List)
            .map((e) => CurrencyModel.fromJson(e))
            .where((c) => !c.isdisabled)
            .toList();

        AppLogger.debug('[POCurrencyHandler] จำนวนสกุลเงิน: ${loadedCurrencies.length}');

        // STEP 1: กำหนด Base Currency
        final baseCurrencyFromApi = TransactionCurrencyUtils.findBaseCurrency(loadedCurrencies);
        baseCurrency = baseCurrencyFromApi.code;
        baseCurrencySymbol = baseCurrencyFromApi.symbol;

        AppLogger.debug('[POCurrencyHandler] baseCurrency: $baseCurrency, docCurrency: ${screenData.docCurrency}');

        // STEP 2: (ยกเลิกแล้ว) ไม่ migrate legacy documents — เดินหน้าอย่างเดียว

        // STEP 3: ตรวจสอบจำนวนสกุลเงิน
        if (loadedCurrencies.length <= 1) {
          _handleSingleCurrency(loadedCurrencies, screenData);
          return;
        }

        // STEP 4: มีหลายสกุลเงิน - ตั้งค่าเริ่มต้น
        currencies = loadedCurrencies;
        _initializeCurrencySelection(screenData);
        onStateChanged();

        AppLogger.info('[POCurrencyHandler] โหลดสำเร็จ: ${currencies.length} สกุลเงิน, base=$baseCurrency, docCurrencyCode=$docCurrencyCode, isMultiCurrency=$isMultiCurrency');
      } else {
        AppLogger.warn('[POCurrencyHandler] API response ไม่สำเร็จ หรือ data เป็น null');
      }
    } catch (e) {
      AppLogger.error('[POCurrencyHandler] โหลดสกุลเงินผิดพลาด: $e');
      // ใช้ค่าจากสาขาเป็น baseCurrency
      baseCurrency = global.getBaseCurrency();
      baseCurrencySymbol = global.getBaseCurrencySymbol();

      // ถ้าเอกสารมี docCurrency อยู่แล้ว ให้ preserve ไว้
      final docCur = screenData.docCurrency;
      if (docCur != null && docCur.isNotEmpty && docCur != baseCurrency) {
        docCurrencyCode = docCur;
        docCurrencySymbol = screenData.docCurrencySymbol;
        exchangeRate = screenData.safeExchangeRate;
        AppLogger.info('[POCurrencyHandler] API ผิดพลาดแต่ preserve docCurrency: $docCur');
      }
    }
  }

  /// จัดการกรณีมีสกุลเงินเดียว
  void _handleSingleCurrency(List<CurrencyModel> loadedCurrencies, TransactionModel screenData) {
    // ตรวจสอบจาก docCurrency โดยตรง (ไม่ใช้ extension getter เพราะต้องการ currency field ที่อาจเป็น null)
    final hasDocCurrency = screenData.docCurrency != null &&
        screenData.docCurrency!.isNotEmpty &&
        screenData.docCurrency != baseCurrency;
    if (hasDocCurrency) {
      // เอกสารเก่ามี multi-currency
      docCurrencyCode = screenData.docCurrency;
      docCurrencySymbol = screenData.docCurrencySymbol;
      exchangeRate = screenData.safeExchangeRate;

      currencies = TransactionCurrencyUtils.createCurrencyListForExistingDoc(
        baseCurrency!,
        baseCurrencySymbol!,
        screenData.docCurrency!,
        screenData.docCurrencySymbol,
      );
      onStateChanged();

      AppLogger.info('[POCurrencyHandler] เอกสารเก่ามี multi-currency: ${screenData.docCurrency}');
      return;
    }

    // เอกสารใหม่ → ใช้ Base Currency
    _initializeWithBaseCurrency(screenData);
    AppLogger.info('[POCurrencyHandler] มีสกุลเงินแค่ ${loadedCurrencies.length} สกุล');
  }

  /// ตั้งค่าเริ่มต้นสำหรับการเลือกสกุลเงิน
  void _initializeCurrencySelection(TransactionModel screenData) {
    if (screenData.docCurrency != null && screenData.docCurrency!.isNotEmpty) {
      docCurrencyCode = screenData.docCurrency;
      docCurrencySymbol = screenData.docCurrencySymbol;
      exchangeRate = screenData.safeExchangeRate;
    } else {
      _initializeWithBaseCurrency(screenData);
    }
  }

  /// ตั้งค่าเริ่มต้นด้วย Base Currency
  void _initializeWithBaseCurrency(TransactionModel screenData) {
    docCurrencyCode = baseCurrency;
    docCurrencySymbol = baseCurrencySymbol;
    exchangeRate = 1.0;

    TransactionCurrencyUtils.initializeWithBaseCurrency(
      screenData,
      baseCurrency!,
      baseCurrencySymbol!,
    );
  }

  /// เปลี่ยนสกุลเงินที่เลือก
  Future<void> onCurrencyChanged(
    String? code,
    String? symbol,
    double? rate,
    TransactionModel screenData,
  ) async {
    // ถ้าเป็นสกุลเงินหลัก
    if (code == baseCurrency) {
      docCurrencyCode = code;
      docCurrencySymbol = symbol;
      exchangeRate = 1.0;

      screenData.currency = baseCurrency;
      screenData.currencysymbol = baseCurrencySymbol;
      screenData.docCurrency = code;
      screenData.docCurrencySymbol = symbol;
      screenData.exchangerate = 1.0;

      // ล้าง priceDoc/sumAmountDoc เมื่อกลับเป็น base currency
      _clearDocPrices(screenData);

      onStateChanged();
      onRecalculate();

      AppLogger.info('[POCurrencyHandler] เปลี่ยนเป็น base currency: $code');
      return;
    }

    // ดึงอัตราแลกเปลี่ยนล่าสุด
    double newRate = rate ?? 1.0;

    if (code != null) {
      try {
        final today = DateTime.now();
        final dateStr = '${today.year}-${today.month.toString().padLeft(2, '0')}-${today.day.toString().padLeft(2, '0')}';
        final apiService = CurrencyApiService();
        final latestRate = await apiService.getLatestExchangeRate(
          currency: code,
          date: dateStr,
        );
        if (latestRate != null && latestRate.rate > 0) {
          newRate = latestRate.rate;
          AppLogger.info('[POCurrencyHandler] ดึงอัตราแลกเปลี่ยนจาก API: $code = $newRate');
        } else {
          onWarning('ไม่พบอัตราแลกเปลี่ยนสำหรับ $code โปรดตรวจสอบและกรอกเอง');
        }
      } catch (e) {
        AppLogger.error('[POCurrencyHandler] ดึงอัตราแลกเปลี่ยนไม่สำเร็จ: $e');
        onWarning('ดึงอัตราแลกเปลี่ยนไม่สำเร็จ โปรดตรวจสอบและกรอกเอง');
      }
    }

    // Validation
    final validation = TransactionCurrencyUtils.validateExchangeRate(
      newRate,
      code ?? '',
      baseCurrency!,
    );
    if (!validation.isValid) {
      newRate = validation.correctedRate;
    }
    if (validation.hasWarning && validation.message != null) {
      onWarning(validation.message!);
    }

    docCurrencyCode = code;
    docCurrencySymbol = symbol;
    exchangeRate = newRate;

    screenData.currency = baseCurrency;
    screenData.currencysymbol = baseCurrencySymbol;
    screenData.docCurrency = code;
    screenData.docCurrencySymbol = symbol;
    screenData.exchangerate = newRate;

    // ตั้งค่า priceDoc สำหรับสินค้าที่มีอยู่แล้ว (ยังไม่มี priceDoc)
    _initializeDocPricesForExistingDetails(screenData);

    onStateChanged();
    onRecalculate();

    AppLogger.info('[POCurrencyHandler] เปลี่ยนสกุลเงิน: doc=$code, rate=$newRate');
  }

  /// แก้ไขอัตราแลกเปลี่ยนเอง
  void onExchangeRateChanged(double? rate, TransactionModel screenData) {
    double safeRate = rate ?? 1.0;

    if (docCurrencyCode != null && docCurrencyCode != baseCurrency) {
      final validation = TransactionCurrencyUtils.validateExchangeRate(
        safeRate,
        docCurrencyCode!,
        baseCurrency!,
      );
      safeRate = validation.correctedRate;
      if (validation.hasWarning && validation.message != null) {
        onWarning(validation.message!);
      }
    } else {
      if (safeRate != 1.0) {
        safeRate = 1.0;
        onWarning(global.language('base_currency_exchange_rate_must_be_one'));
      }
    }

    exchangeRate = safeRate;
    screenData.exchangerate = exchangeRate;

    // คำนวณ base price ใหม่จาก priceDoc เมื่อเปลี่ยน rate
    _recalculateBasePricesFromDocPrices(screenData);

    onStateChanged();
    onRecalculate();

    AppLogger.info('[POCurrencyHandler] แก้ไขอัตราแลกเปลี่ยน: $safeRate');
  }

  /// คำนวณ base price ใหม่จาก priceDoc เมื่อเปลี่ยน exchange rate
  /// สำหรับ detail ที่มี priceDoc (ป้อนราคาเป็นสกุลเงินเอกสาร)
  void _recalculateBasePricesFromDocPrices(TransactionModel screenData) {
    if (screenData.details == null) return;
    for (var detail in screenData.details!) {
      if (detail.priceDoc != null) {
        detail.price = detail.priceDoc! * exchangeRate;
      }
    }
  }

  /// ตั้งค่า priceDoc สำหรับสินค้าที่มีอยู่แล้วแต่ยังไม่มี priceDoc
  /// เรียกเมื่อเปลี่ยนเป็น doc currency ใหม่ → สินค้าเดิมต้องมี 2 สกุลเงินเก็บไว้
  void _initializeDocPricesForExistingDetails(TransactionModel screenData) {
    if (screenData.details == null || exchangeRate <= 0) return;
    for (var detail in screenData.details!) {
      if (detail.priceDoc == null && detail.price > 0) {
        detail.priceDoc = detail.price / exchangeRate;
      }
    }
  }

  /// ล้าง doc currency fields ทั้งหมดเมื่อเปลี่ยนกลับเป็น base currency
  void _clearDocPrices(TransactionModel screenData) {
    if (screenData.details == null) return;
    for (var detail in screenData.details!) {
      detail.priceDoc = null;
      detail.sumAmountDoc = null;
      detail.discountamountDoc = null;
      detail.priceexcludevatDoc = null;
      detail.sumamountexcludevatDoc = null;
      detail.totalvaluevatDoc = null;
    }
  }

  /// Reset สถานะ
  void reset() {
    currencies = [];
    docCurrencyCode = null;
    docCurrencySymbol = null;
    exchangeRate = 1.0;
  }
}
