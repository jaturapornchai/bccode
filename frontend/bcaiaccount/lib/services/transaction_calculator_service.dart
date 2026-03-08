import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:smlaicloud/global.dart' as global;

/// Transaction Calculator Service
/// ใช้ Backend API สำหรับการคำนวณธุรกรรม (VAT, ส่วนลด, การชำระเงิน)
///
/// Enhanced: รองรับ vatcal, 2-level discount, deposit-adjusted VAT,
/// multi-currency, และ return document
class TransactionCalculatorService {
  static String get _baseUrl => global.goApiBaseUrl;

  /// สร้าง headers สำหรับ API (รวม Authorization token)
  static Map<String, String> get _headers {
    final headers = <String, String>{'Content-Type': 'application/json'};
    final token = global.appConfig.getString("token") ?? '';
    if (token.isNotEmpty) {
      headers['Authorization'] = 'Bearer $token';
    }
    return headers;
  }

  /// คำนวณธุรกรรมแบบเต็ม (Enhanced)
  ///
  /// [items] - รายการสินค้า
  /// [discount] - ส่วนลดท้ายบิล (discountword)
  /// [detailDiscount] - ส่วนลดก่อนชำระเงิน (detaildiscountformula)
  /// [vatType] - ประเภทภาษี (0=แยกนอก, 1=รวมใน, 2=ไม่กระทบภาษี)
  /// [vatRate] - อัตราภาษี (default 7)
  /// [payments] - การชำระเงิน
  /// [depositAmount] - เงินมัดจำ
  /// [isManualAmount] - ไม่คำนวณยอดเอง
  /// [docCurrency] - สกุลเงินเอกสาร
  /// [baseCurrency] - สกุลเงินหลัก
  /// [exchangeRate] - อัตราแลกเปลี่ยน
  /// [transactionType] - ประเภทธุรกรรม
  /// [refTotalOriginal] - ยอดเอกสารอ้างอิง (สำหรับ return)
  static Future<TransactionCalculationResult> calculate({
    required List<TransactionItem> items,
    String discount = "",
    String detailDiscount = "",
    int vatType = 0,
    double vatRate = 7.0,
    List<PaymentMethod>? payments,
    double depositAmount = 0,
    bool isManualAmount = false,
    int inquiryType = 0,
    String? docCurrency,
    String? baseCurrency,
    double? exchangeRate,
    String? transactionType,
    double? refTotalOriginal,
    double roundAmount = 0,
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/transaction/calculate");

      final body = {
        "items": items.map((item) => item.toJson()).toList(),
        "discount": discount,
        "detail_discount": detailDiscount,
        "vattype": vatType,
        "vatrate": vatRate,
        "payments": (payments ?? []).map((p) => p.toJson()).toList(),
        "deposit_amount": depositAmount,
        "is_manual_amount": isManualAmount,
        "inquiry_type": inquiryType,
        if (docCurrency != null) "doc_currency": docCurrency,
        if (baseCurrency != null) "base_currency": baseCurrency,
        if (exchangeRate != null) "exchange_rate": exchangeRate,
        if (transactionType != null) "transaction_type": transactionType,
        if (refTotalOriginal != null) "ref_total_original": refTotalOriginal,
        "round_amount": roundAmount,
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          return TransactionCalculationResult.fromJson(json['data']);
        }
      }

      return TransactionCalculationResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      return TransactionCalculationResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// คำนวณแบบย่อ (สำหรับยอดเดียว)
  static Future<QuickCalcResult> quickCalc({
    required double amount,
    String discount = "",
    int vatType = 0,
    double vatRate = 7.0,
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/transaction/quick-calc");

      final body = {
        "amount": amount,
        "discount": discount,
        "vattype": vatType,
        "vatrate": vatRate,
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          return QuickCalcResult.fromJson(json['data']);
        }
      }

      return QuickCalcResult.error("API call failed: ${response.reasonPhrase}");
    } catch (e) {
      return QuickCalcResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ตรวจสอบการชำระเงิน
  static Future<PaymentValidationResult> validatePayment({
    required double totalAmount,
    double depositAmount = 0,
    required List<PaymentMethod> payments,
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/transaction/validate-payment");

      final body = {
        "total_amount": totalAmount,
        "deposit_amount": depositAmount,
        "payments": payments.map((p) => p.toJson()).toList(),
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          return PaymentValidationResult.fromJson(json['data']);
        }
      }

      return PaymentValidationResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      return PaymentValidationResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }
}

/// รายการสินค้าสำหรับคำนวณ (Enhanced)
class TransactionItem {
  final String itemCode;
  final String itemName;
  final double qty;
  final double price;
  final String discount;
  final int vatType;
  final double vatRate;
  final int vatCal; // 0=taxable, 1=exempt
  final String unitCode;
  final double unitStand;
  final double unitDivide;
  final String whCode;
  final String shelfCode;
  final int calcFlag;

  // Calculated fields (returned from API)
  final double? sumAmount;
  final double? discountAmount;
  final double? priceExcludeVat;
  final double? sumAmountExcludeVat;
  final double? totalValueVat;
  final double? totalQty;

  TransactionItem({
    required this.itemCode,
    this.itemName = "",
    required this.qty,
    required this.price,
    this.discount = "",
    this.vatType = 0,
    this.vatRate = 7.0,
    this.vatCal = 0,
    this.unitCode = "",
    this.unitStand = 1,
    this.unitDivide = 1,
    this.whCode = "",
    this.shelfCode = "",
    this.calcFlag = 0,
    this.sumAmount,
    this.discountAmount,
    this.priceExcludeVat,
    this.sumAmountExcludeVat,
    this.totalValueVat,
    this.totalQty,
  });

  Map<String, dynamic> toJson() => {
    "itemcode": itemCode,
    "itemname": itemName,
    "qty": qty,
    "price": price,
    "discount": discount,
    "vattype": vatType,
    "vatrate": vatRate,
    "vatcal": vatCal,
    "unitcode": unitCode,
    "unitstand": unitStand,
    "unitdivide": unitDivide,
    "whcode": whCode,
    "shelfcode": shelfCode,
    "calcflag": calcFlag,
  };

  factory TransactionItem.fromJson(Map<String, dynamic> json) =>
      TransactionItem(
        itemCode: json['itemcode'] ?? '',
        itemName: json['itemname'] ?? '',
        qty: (json['qty'] ?? 0).toDouble(),
        price: (json['price'] ?? 0).toDouble(),
        discount: json['discount'] ?? '',
        vatType: json['vattype'] ?? 0,
        vatRate: (json['vatrate'] ?? 7).toDouble(),
        vatCal: json['vatcal'] ?? 0,
        unitCode: json['unitcode'] ?? '',
        unitStand: (json['unitstand'] ?? 1).toDouble(),
        unitDivide: (json['unitdivide'] ?? 1).toDouble(),
        whCode: json['whcode'] ?? '',
        shelfCode: json['shelfcode'] ?? '',
        calcFlag: json['calcflag'] ?? 0,
        sumAmount: (json['sumamount'] ?? 0).toDouble(),
        discountAmount: (json['discountamount'] ?? 0).toDouble(),
        priceExcludeVat: (json['priceexcludevat'] ?? 0).toDouble(),
        sumAmountExcludeVat: (json['sumamountexcludevat'] ?? 0).toDouble(),
        totalValueVat: (json['totalvaluevat'] ?? 0).toDouble(),
        totalQty: (json['totalqty'] ?? 0).toDouble(),
      );
}

/// วิธีการชำระเงิน
class PaymentMethod {
  final String payCode;
  final String payName;
  final double amount;
  final String? transferBankCode;
  final String? creditCardType;
  final String? chequeNo;
  final String? chequeDate;
  final String? chequeBankCode;

  PaymentMethod({
    required this.payCode,
    this.payName = "",
    required this.amount,
    this.transferBankCode,
    this.creditCardType,
    this.chequeNo,
    this.chequeDate,
    this.chequeBankCode,
  });

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> json = {
      "paycode": payCode,
      "payname": payName,
      "amount": amount,
    };
    if (transferBankCode != null) json["transfer_bank_code"] = transferBankCode;
    if (creditCardType != null) json["credit_card_type"] = creditCardType;
    if (chequeNo != null) json["cheque_no"] = chequeNo;
    if (chequeDate != null) json["cheque_date"] = chequeDate;
    if (chequeBankCode != null) json["cheque_bank_code"] = chequeBankCode;
    return json;
  }
}

/// ผลลัพธ์การคำนวณธุรกรรม (Enhanced)
class TransactionCalculationResult {
  final bool isSuccess;
  final String? errorMessage;
  final List<TransactionItem> items;

  // Enhanced fields (Flutter-compatible)
  final double totalValue;
  final double totalQty;
  final double detailTotalDiscount;
  final double totalDiscountVatAmount;
  final double totalDiscountExceptVatAmount;
  final double totalBeforeVat;
  final double totalVatValue;
  final double totalAfterVat;
  final double totalExceptVat;
  final double detailTotalAmount;
  final double totalDiscount;
  final double totalAmountAfterDiscount;
  final double totalAmount;
  final double totalAmountDoc;
  final double totalValueDoc;
  final double totalDiscountDoc;
  final double totalVatValueDoc;
  final double totalBeforeVatDoc;
  final double totalAfterVatDoc;
  final double refTotalDiff;
  final double refTotalCorrect;

  // Legacy fields
  final double sumAmount;
  final double discountAmount;
  final double beforeVatAmount;
  final double vatAmount;
  final double netAmount;
  final double depositAmount;
  final double paymentTotal;
  final double changeAmount;
  final double remainingAmount;
  final bool isPaymentValid;
  final Map<String, double> paymentBreakdown;

  TransactionCalculationResult({
    required this.isSuccess,
    this.errorMessage,
    this.items = const [],
    this.totalValue = 0,
    this.totalQty = 0,
    this.detailTotalDiscount = 0,
    this.totalDiscountVatAmount = 0,
    this.totalDiscountExceptVatAmount = 0,
    this.totalBeforeVat = 0,
    this.totalVatValue = 0,
    this.totalAfterVat = 0,
    this.totalExceptVat = 0,
    this.detailTotalAmount = 0,
    this.totalDiscount = 0,
    this.totalAmountAfterDiscount = 0,
    this.totalAmount = 0,
    this.totalAmountDoc = 0,
    this.totalValueDoc = 0,
    this.totalDiscountDoc = 0,
    this.totalVatValueDoc = 0,
    this.totalBeforeVatDoc = 0,
    this.totalAfterVatDoc = 0,
    this.refTotalDiff = 0,
    this.refTotalCorrect = 0,
    this.sumAmount = 0,
    this.discountAmount = 0,
    this.beforeVatAmount = 0,
    this.vatAmount = 0,
    this.netAmount = 0,
    this.depositAmount = 0,
    this.paymentTotal = 0,
    this.changeAmount = 0,
    this.remainingAmount = 0,
    this.isPaymentValid = false,
    this.paymentBreakdown = const {},
  });

  factory TransactionCalculationResult.fromJson(Map<String, dynamic> json) {
    final itemsList = (json['items'] as List?)
        ?.map((item) => TransactionItem.fromJson(item))
        .toList() ?? [];

    final breakdown = <String, double>{};
    if (json['payment_breakdown'] != null) {
      (json['payment_breakdown'] as Map).forEach((key, value) {
        breakdown[key.toString()] = (value ?? 0).toDouble();
      });
    }

    return TransactionCalculationResult(
      isSuccess: true,
      items: itemsList,
      // Enhanced fields
      totalValue: (json['total_value'] ?? 0).toDouble(),
      totalQty: (json['total_qty'] ?? 0).toDouble(),
      detailTotalDiscount: (json['detail_total_discount'] ?? 0).toDouble(),
      totalDiscountVatAmount: (json['total_discount_vat_amount'] ?? 0).toDouble(),
      totalDiscountExceptVatAmount: (json['total_discount_except_vat_amount'] ?? 0).toDouble(),
      totalBeforeVat: (json['total_before_vat'] ?? 0).toDouble(),
      totalVatValue: (json['total_vat_value'] ?? 0).toDouble(),
      totalAfterVat: (json['total_after_vat'] ?? 0).toDouble(),
      totalExceptVat: (json['total_except_vat'] ?? 0).toDouble(),
      detailTotalAmount: (json['detail_total_amount'] ?? 0).toDouble(),
      totalDiscount: (json['total_discount'] ?? 0).toDouble(),
      totalAmountAfterDiscount: (json['total_amount_after_discount'] ?? 0).toDouble(),
      totalAmount: (json['total_amount'] ?? 0).toDouble(),
      totalAmountDoc: (json['total_amount_doc'] ?? 0).toDouble(),
      totalValueDoc: (json['total_value_doc'] ?? 0).toDouble(),
      totalDiscountDoc: (json['total_discount_doc'] ?? 0).toDouble(),
      totalVatValueDoc: (json['total_vat_value_doc'] ?? 0).toDouble(),
      totalBeforeVatDoc: (json['total_before_vat_doc'] ?? 0).toDouble(),
      totalAfterVatDoc: (json['total_after_vat_doc'] ?? 0).toDouble(),
      refTotalDiff: (json['ref_total_diff'] ?? 0).toDouble(),
      refTotalCorrect: (json['ref_total_correct'] ?? 0).toDouble(),
      // Legacy fields
      sumAmount: (json['sum_amount'] ?? 0).toDouble(),
      discountAmount: (json['discount_amount'] ?? 0).toDouble(),
      beforeVatAmount: (json['before_vat_amount'] ?? 0).toDouble(),
      vatAmount: (json['vat_amount'] ?? 0).toDouble(),
      netAmount: (json['net_amount'] ?? 0).toDouble(),
      depositAmount: (json['deposit_amount'] ?? 0).toDouble(),
      paymentTotal: (json['payment_total'] ?? 0).toDouble(),
      changeAmount: (json['change_amount'] ?? 0).toDouble(),
      remainingAmount: (json['remaining_amount'] ?? 0).toDouble(),
      isPaymentValid: json['is_payment_valid'] ?? false,
      paymentBreakdown: breakdown,
    );
  }

  factory TransactionCalculationResult.error(String message) =>
      TransactionCalculationResult(
        isSuccess: false,
        errorMessage: message,
      );
}

/// ผลลัพธ์การคำนวณแบบย่อ
class QuickCalcResult {
  final bool isSuccess;
  final String? errorMessage;
  final double amount;
  final double discountAmount;
  final double afterDiscount;
  final double beforeVat;
  final double vatAmount;
  final double totalAmount;

  QuickCalcResult({
    required this.isSuccess,
    this.errorMessage,
    this.amount = 0,
    this.discountAmount = 0,
    this.afterDiscount = 0,
    this.beforeVat = 0,
    this.vatAmount = 0,
    this.totalAmount = 0,
  });

  factory QuickCalcResult.fromJson(Map<String, dynamic> json) => QuickCalcResult(
    isSuccess: true,
    amount: (json['amount'] ?? 0).toDouble(),
    discountAmount: (json['discount_amount'] ?? 0).toDouble(),
    afterDiscount: (json['after_discount'] ?? 0).toDouble(),
    beforeVat: (json['before_vat'] ?? 0).toDouble(),
    vatAmount: (json['vat_amount'] ?? 0).toDouble(),
    totalAmount: (json['total_amount'] ?? 0).toDouble(),
  );

  factory QuickCalcResult.error(String message) => QuickCalcResult(
    isSuccess: false,
    errorMessage: message,
  );
}

/// ผลลัพธ์การตรวจสอบการชำระเงิน
class PaymentValidationResult {
  final bool isSuccess;
  final String? errorMessage;
  final double netAmount;
  final double paymentTotal;
  final double changeAmount;
  final double remainingAmount;
  final bool isValid;
  final String message;
  final Map<String, double> paymentBreakdown;

  PaymentValidationResult({
    required this.isSuccess,
    this.errorMessage,
    this.netAmount = 0,
    this.paymentTotal = 0,
    this.changeAmount = 0,
    this.remainingAmount = 0,
    this.isValid = false,
    this.message = "",
    this.paymentBreakdown = const {},
  });

  factory PaymentValidationResult.fromJson(Map<String, dynamic> json) {
    final breakdown = <String, double>{};
    if (json['payment_breakdown'] != null) {
      (json['payment_breakdown'] as Map).forEach((key, value) {
        breakdown[key.toString()] = (value ?? 0).toDouble();
      });
    }

    return PaymentValidationResult(
      isSuccess: true,
      netAmount: (json['net_amount'] ?? 0).toDouble(),
      paymentTotal: (json['payment_total'] ?? 0).toDouble(),
      changeAmount: (json['change_amount'] ?? 0).toDouble(),
      remainingAmount: (json['remaining_amount'] ?? 0).toDouble(),
      isValid: json['is_valid'] ?? false,
      message: json['message'] ?? '',
      paymentBreakdown: breakdown,
    );
  }

  factory PaymentValidationResult.error(String message) => PaymentValidationResult(
    isSuccess: false,
    errorMessage: message,
  );
}
