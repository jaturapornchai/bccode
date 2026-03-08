import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:smlaicloud/global.dart' as global;

/// Sales Report API Service
/// ใช้ Backend API สำหรับรายงานขาย (ปลอดภัยจาก SQL injection)
class SalesReportApiService {
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

  /// ดึงรายงานขายตามเอกสาร (header หรือ detail)
  ///
  /// [fromDate] - วันที่เริ่มต้น (YYYY-MM-DD)
  /// [toDate] - วันที่สิ้นสุด (YYYY-MM-DD)
  /// [branchCodes] - รหัสสาขา (optional)
  /// [productCodes] - รหัสสินค้า (optional)
  /// [reportType] - "header" หรือ "detail"
  /// [sortAscending] - เรียงจากน้อยไปมาก
  /// [limit] - จำนวน record สูงสุด
  /// [offset] - เริ่มต้นจาก record ที่
  static Future<SalesReportResult> getSalesReport({
    required String fromDate,
    required String toDate,
    List<String>? branchCodes,
    List<String>? productCodes,
    String reportType = "header",
    bool sortAscending = true,
    int limit = 1000,
    int offset = 0,
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/report/sales/by-document");

      final Map<String, dynamic> body = {
        "shop_id": global.getShopId(),
        "from_date": fromDate,
        "to_date": toDate,
        "report_type": reportType,
        "sort_ascending": sortAscending,
        "limit": limit,
        "offset": offset,
      };

      if (branchCodes != null && branchCodes.isNotEmpty) {
        body["branch_codes"] = branchCodes;
      }
      if (productCodes != null && productCodes.isNotEmpty) {
        body["product_codes"] = productCodes;
      }

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          return SalesReportResult.fromJson(json);
        }
      }

      return SalesReportResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      return SalesReportResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ดึงรายงานขาย Header
  static Future<SalesReportResult> getSalesReportHeader({
    required String fromDate,
    required String toDate,
    List<String>? branchCodes,
    List<String>? productCodes,
    bool sortAscending = true,
    int limit = 1000,
    int offset = 0,
  }) {
    return getSalesReport(
      fromDate: fromDate,
      toDate: toDate,
      branchCodes: branchCodes,
      productCodes: productCodes,
      reportType: "header",
      sortAscending: sortAscending,
      limit: limit,
      offset: offset,
    );
  }

  /// ดึงรายงานขาย Detail
  static Future<SalesReportResult> getSalesReportDetail({
    required String fromDate,
    required String toDate,
    List<String>? branchCodes,
    List<String>? productCodes,
    bool sortAscending = true,
    int limit = 1000,
    int offset = 0,
  }) {
    return getSalesReport(
      fromDate: fromDate,
      toDate: toDate,
      branchCodes: branchCodes,
      productCodes: productCodes,
      reportType: "detail",
      sortAscending: sortAscending,
      limit: limit,
      offset: offset,
    );
  }

  /// ดึงสรุปยอดขาย
  ///
  /// [fromDate] - วันที่เริ่มต้น (YYYY-MM-DD)
  /// [toDate] - วันที่สิ้นสุด (YYYY-MM-DD)
  /// [branchCodes] - รหัสสาขา (optional)
  /// [productCodes] - รหัสสินค้า (optional)
  static Future<SalesSummaryResult> getSalesSummary({
    required String fromDate,
    required String toDate,
    List<String>? branchCodes,
    List<String>? productCodes,
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/report/sales/summary");

      final Map<String, dynamic> body = {
        "shop_id": global.getShopId(),
        "from_date": fromDate,
        "to_date": toDate,
      };

      if (branchCodes != null && branchCodes.isNotEmpty) {
        body["branch_codes"] = branchCodes;
      }
      if (productCodes != null && productCodes.isNotEmpty) {
        body["product_codes"] = productCodes;
      }

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          return SalesSummaryResult.fromJson(json['data']);
        }
      }

      return SalesSummaryResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      return SalesSummaryResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }
}

/// ผลลัพธ์รายงานขาย
class SalesReportResult {
  final bool isSuccess;
  final String? errorMessage;
  final List<Map<String, dynamic>> data;
  final int count;
  final int limit;
  final int offset;
  final String reportType;

  SalesReportResult({
    required this.isSuccess,
    this.errorMessage,
    this.data = const [],
    this.count = 0,
    this.limit = 0,
    this.offset = 0,
    this.reportType = "",
  });

  factory SalesReportResult.fromJson(Map<String, dynamic> json) {
    final dataList = (json['data'] as List?)
            ?.map((item) => Map<String, dynamic>.from(item))
            .toList() ??
        [];

    return SalesReportResult(
      isSuccess: true,
      data: dataList,
      count: json['count'] ?? 0,
      limit: json['limit'] ?? 0,
      offset: json['offset'] ?? 0,
      reportType: json['report_type'] ?? '',
    );
  }

  factory SalesReportResult.error(String message) => SalesReportResult(
        isSuccess: false,
        errorMessage: message,
      );

  bool get hasMore => offset + count < limit;
}

/// ผลลัพธ์สรุปยอดขาย
class SalesSummaryResult {
  final bool isSuccess;
  final String? errorMessage;
  final int totalDocuments;
  final double totalAmount;
  final double totalCost;
  final double totalProfit;
  final double profitMargin;

  SalesSummaryResult({
    required this.isSuccess,
    this.errorMessage,
    this.totalDocuments = 0,
    this.totalAmount = 0,
    this.totalCost = 0,
    this.totalProfit = 0,
    this.profitMargin = 0,
  });

  factory SalesSummaryResult.fromJson(Map<String, dynamic> json) {
    return SalesSummaryResult(
      isSuccess: true,
      totalDocuments: json['total_documents'] ?? 0,
      totalAmount: (json['total_amount'] ?? 0).toDouble(),
      totalCost: (json['total_cost'] ?? 0).toDouble(),
      totalProfit: (json['total_profit'] ?? 0).toDouble(),
      profitMargin: (json['profit_margin'] ?? 0).toDouble(),
    );
  }

  factory SalesSummaryResult.error(String message) => SalesSummaryResult(
        isSuccess: false,
        errorMessage: message,
      );
}
