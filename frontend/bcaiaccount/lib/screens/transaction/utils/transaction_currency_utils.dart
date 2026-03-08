import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;

/// ============================================
/// Transaction Currency Utilities - รวมศูนย์ Currency Logic
/// ============================================
/// 
/// รวม logic ที่เกี่ยวกับ Multi-Currency ทั้งหมดไว้ที่นี่
/// แทนการกระจายอยู่ทั่ว transaction_edit.dart และไฟล์อื่นๆ

class TransactionCurrencyUtils {
  
  /// ดึง Base Currency จากสาขา → ร้านค้า → Default THB
  static String getBaseCurrencyCode() {
    return global.getBaseCurrency();
  }

  /// ดึง Base Currency Symbol จากสาขา
  static String getBaseCurrencySymbol() {
    return global.getBaseCurrencySymbol();
  }

  /// หา Base Currency จาก List
  static CurrencyModel findBaseCurrency(List<CurrencyModel> currencies) {
    final configBase = getBaseCurrencyCode();
    
    return currencies.firstWhere(
      (c) => c.code == configBase,
      orElse: () => currencies.firstWhere(
        (c) => c.code == 'THB',
        orElse: () => currencies.isNotEmpty 
            ? currencies.first 
            : CurrencyModel(code: 'THB', name: 'Thai Baht', symbol: '฿'),
      ),
    );
  }

  /// ตรวจสอบว่าเป็น Multi-Currency หรือไม่
  static bool isMultiCurrency({
    required String? docCurrency,
    required String? baseCurrency,
    required double? exchangeRate,
  }) {
    if (docCurrency == null || docCurrency.isEmpty) return false;
    if (baseCurrency == null || baseCurrency.isEmpty) return false;
    if (docCurrency == baseCurrency) return false;
    if (exchangeRate == null || exchangeRate == 1.0) return false;
    if (exchangeRate <= 0) return false;
    
    return true;
  }

  /// ตั้งค่าเริ่มต้นสำหรับเอกสารใหม่
  static void initializeWithBaseCurrency(
    TransactionModel screenData,
    String baseCurrency,
    String baseCurrencySymbol,
  ) {
    screenData.currency = baseCurrency;
    screenData.currencysymbol = baseCurrencySymbol;
    screenData.docCurrency = baseCurrency;
    screenData.docCurrencySymbol = baseCurrencySymbol;
    screenData.exchangerate = 1.0;
  }

  /// คำนวณยอดในสกุลเงินเอกสาร
  static double calculateDocumentAmount(double baseAmount, double exchangeRate) {
    if (exchangeRate <= 0 || exchangeRate.isNaN || exchangeRate.isInfinite) {
      return baseAmount;
    }
    final dp = global.getDecimalDocument();
    return double.parse((baseAmount / exchangeRate).toStringAsFixed(dp));
  }

  /// คำนวณยอดในสกุลเงินหลัก (กลับกัน)
  static double calculateBaseAmount(double docAmount, double exchangeRate) {
    if (exchangeRate <= 0 || exchangeRate.isNaN || exchangeRate.isInfinite) {
      return docAmount;
    }
    final dp = global.getDecimalDocument();
    return double.parse((docAmount * exchangeRate).toStringAsFixed(dp));
  }

  /// ตรวจสอบ Exchange Rate ว่าถูกต้องหรือไม่
  static ExchangeRateValidationResult validateExchangeRate(
    double rate, 
    String currencyCode,
    String baseCurrency,
  ) {
    // กรณี rate = 1 แต่ไม่ใช่ base currency
    if (rate == 1.0 && currencyCode != baseCurrency) {
      return ExchangeRateValidationResult(
        isValid: false,
        correctedRate: rate,
        hasWarning: true,
        message: 'อัตราแลกเปลี่ยน $currencyCode = 1.0 $baseCurrency อาจไม่ถูกต้อง โปรดตรวจสอบ',
      );
    }
    
    // กรณี rate <= 0 หรือผิดปกติ
    if (rate <= 0 || rate.isNaN || rate.isInfinite) {
      return ExchangeRateValidationResult(
        isValid: false,
        correctedRate: 1.0,
        hasWarning: true,
        message: 'อัตราแลกเปลี่ยนไม่ถูกต้อง ใช้ค่าเริ่มต้น 1.0 แทน',
      );
    }
    
    // กรณี rate สูง/ต่ำผิดปกติ
    if (rate > 1000 || rate < 0.001) {
      return ExchangeRateValidationResult(
        isValid: true,
        correctedRate: rate,
        hasWarning: true,
        message: 'อัตราแลกเปลี่ยน $rate อาจผิดปกติ (ปกติ 0.001 - 1000) โปรดตรวจสอบ',
      );
    }
    
    return ExchangeRateValidationResult(
      isValid: true,
      correctedRate: rate,
      hasWarning: false,
      message: null,
    );
  }

  /// ปรับปรุง exchange rate ให้ปลอดภัย
  static double sanitizeExchangeRate(double? rate) {
    if (rate == null || rate <= 0 || rate.isNaN || rate.isInfinite) {
      return 1.0;
    }
    return rate;
  }

  /// สร้างรายการสกุลเงินสำหรับเอกสารที่มี Multi-Currency 
  /// (กรณี shop มีแค่ 1 สกุลเงิน แต่เอกสารเก่ามี multi-currency)
  static List<CurrencyModel> createCurrencyListForExistingDoc(
    String baseCurrency,
    String baseCurrencySymbol,
    String docCurrency,
    String? docCurrencySymbol,
  ) {
    return [
      CurrencyModel(
        code: baseCurrency, 
        name: baseCurrency, 
        symbol: baseCurrencySymbol,
      ),
      CurrencyModel(
        code: docCurrency, 
        name: docCurrency, 
        symbol: docCurrencySymbol ?? docCurrency,
      ),
    ];
  }
}

/// ผลลัพธ์การตรวจสอบ exchange rate
class ExchangeRateValidationResult {
  final bool isValid;
  final double correctedRate;
  final bool hasWarning;
  final String? message;

  const ExchangeRateValidationResult({
    required this.isValid,
    required this.correctedRate,
    required this.hasWarning,
    this.message,
  });
}

/// Extension สำหรับ TransactionModel
extension TransactionCurrencyExtension on TransactionModel {
  /// ตรวจสอบว่าเอกสารนี้เป็น Multi-Currency หรือไม่
  bool get isMultiCurrency => TransactionCurrencyUtils.isMultiCurrency(
    docCurrency: docCurrency,
    baseCurrency: currency,
    exchangeRate: exchangerate,
  );

  /// ดึง exchange rate ที่ปลอดภัย
  double get safeExchangeRate => TransactionCurrencyUtils.sanitizeExchangeRate(exchangerate);

  /// คำนวณยอดเอกสารจากยอดหลัก
  double calculateDocAmount(double baseAmount) {
    return TransactionCurrencyUtils.calculateDocumentAmount(baseAmount, safeExchangeRate);
  }
}
