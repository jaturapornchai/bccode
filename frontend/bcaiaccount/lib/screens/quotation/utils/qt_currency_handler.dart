import 'package:flutter/material.dart';
import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/services/currency_api_service.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_currency_utils.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/global.dart' as global;

/// Handler สำหรับจัดการ Multi-Currency ในใบเสนอราคา
///
/// รับผิดชอบ:
/// - โหลดรายการสกุลเงิน
/// - จัดการการเลือกและเปลี่ยนสกุลเงิน
/// - ตรวจสอบและอัพเดทอัตราแลกเปลี่ยน
class QTCurrencyHandler {
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

  QTCurrencyHandler({
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
    AppLogger.debug('[QTCurrencyHandler] loadCurrencies เริ่ม — docCurrency: ${screenData.docCurrency}, exchangerate: ${screenData.exchangerate}');
    try {
      final apiService = CurrencyApiService();
      final shopId = global.prefs.getString("shopid") ?? "";

      final response = await apiService.getCurrencies(
        shopId: shopId,
        page: 1,
        limit: 1000,
      );

      if (response.success && response.data != null) {
        final loadedCurrencies = (response.data as List)
            .map((e) => CurrencyModel.fromJson(e))
            .where((c) => !c.isdisabled)
            .toList();

        // กำหนด Base Currency
        final baseCurrencyFromApi = TransactionCurrencyUtils.findBaseCurrency(loadedCurrencies);
        baseCurrency = baseCurrencyFromApi.code;
        baseCurrencySymbol = baseCurrencyFromApi.symbol;

        if (loadedCurrencies.length <= 1) {
          _handleSingleCurrency(loadedCurrencies, screenData);
          return;
        }

        currencies = loadedCurrencies;
        _initializeCurrencySelection(screenData);
        onStateChanged();

        AppLogger.info('[QTCurrencyHandler] โหลดสำเร็จ: ${currencies.length} สกุลเงิน, base=$baseCurrency');
      } else {
        AppLogger.warn('[QTCurrencyHandler] API response ไม่สำเร็จ หรือ data เป็น null');
      }
    } catch (e) {
      AppLogger.error('[QTCurrencyHandler] โหลดสกุลเงินผิดพลาด: $e');
      baseCurrency = global.getBaseCurrency();
      baseCurrencySymbol = global.getBaseCurrencySymbol();

      final docCur = screenData.docCurrency;
      if (docCur != null && docCur.isNotEmpty && docCur != baseCurrency) {
        docCurrencyCode = docCur;
        docCurrencySymbol = screenData.docCurrencySymbol;
        exchangeRate = screenData.safeExchangeRate;
      }
    }
  }

  void _handleSingleCurrency(List<CurrencyModel> loadedCurrencies, TransactionModel screenData) {
    final hasDocCurrency = screenData.docCurrency != null &&
        screenData.docCurrency!.isNotEmpty &&
        screenData.docCurrency != baseCurrency;
    if (hasDocCurrency) {
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
      return;
    }

    _initializeWithBaseCurrency(screenData);
  }

  void _initializeCurrencySelection(TransactionModel screenData) {
    if (screenData.docCurrency != null && screenData.docCurrency!.isNotEmpty) {
      docCurrencyCode = screenData.docCurrency;
      docCurrencySymbol = screenData.docCurrencySymbol;
      exchangeRate = screenData.safeExchangeRate;
    } else {
      _initializeWithBaseCurrency(screenData);
    }
  }

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
    if (code == baseCurrency) {
      docCurrencyCode = code;
      docCurrencySymbol = symbol;
      exchangeRate = 1.0;

      screenData.currency = baseCurrency;
      screenData.currencysymbol = baseCurrencySymbol;
      screenData.docCurrency = code;
      screenData.docCurrencySymbol = symbol;
      screenData.exchangerate = 1.0;

      _clearDocPrices(screenData);
      onStateChanged();
      onRecalculate();
      return;
    }

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
        } else {
          onWarning('${global.language("exchange_rate_not_found_for")} $code ${global.language("please_check_and_enter_manually")}');
        }
      } catch (e) {
        AppLogger.error('[QTCurrencyHandler] ดึงอัตราแลกเปลี่ยนไม่สำเร็จ: $e');
        onWarning('${global.language("fetch_exchange_rate_failed")} ${global.language("please_check_and_enter_manually")}');
      }
    }

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

    _initializeDocPricesForExistingDetails(screenData);
    onStateChanged();
    onRecalculate();
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
        onWarning(global.language('base_currency_rate_must_be_one'));
      }
    }

    exchangeRate = safeRate;
    screenData.exchangerate = exchangeRate;

    _recalculateBasePricesFromDocPrices(screenData);
    onStateChanged();
    onRecalculate();
  }

  void _recalculateBasePricesFromDocPrices(TransactionModel screenData) {
    if (screenData.details == null) return;
    for (var detail in screenData.details!) {
      if (detail.priceDoc != null) {
        detail.price = detail.priceDoc! * exchangeRate;
      }
    }
  }

  void _initializeDocPricesForExistingDetails(TransactionModel screenData) {
    if (screenData.details == null || exchangeRate <= 0) return;
    for (var detail in screenData.details!) {
      if (detail.priceDoc == null && detail.price > 0) {
        detail.priceDoc = detail.price / exchangeRate;
      }
    }
  }

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
