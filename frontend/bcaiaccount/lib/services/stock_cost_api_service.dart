import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:smlaicloud/global.dart' as global;

/// Secure API Service for Process Stock Cost queries
/// Uses backend parameterized queries to prevent SQL injection
class StockCostApiService {
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

  /// Query stock cost data with filters
  ///
  /// [fromDate] - Start date (YYYY-MM-DD format)
  /// [toDate] - End date (YYYY-MM-DD format)
  /// [branchCodes] - Optional list of branch codes to filter
  /// [productCodes] - Optional list of product codes to filter
  /// [limit] - Max records to return (default 1000, max 10000)
  /// [offset] - Pagination offset
  static Future<Map<String, dynamic>> queryStockCost({
    required String fromDate,
    required String toDate,
    List<String>? branchCodes,
    List<String>? productCodes,
    int limit = 1000,
    int offset = 0,
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/stockcost/query");

      final Map<String, dynamic> body = {
        "shop_id": global.getShopId(),
        "from_date": fromDate,
        "to_date": toDate,
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
        return jsonDecode(response.body);
      } else {
        return {
          "status": "error",
          "code": response.statusCode,
          "message": "API call failed: ${response.reasonPhrase}",
        };
      }
    } catch (e) {
      return {
        "status": "error",
        "code": 500,
        "message": "Connection failed: $e",
      };
    } finally {
      httpClient.close();
    }
  }

  /// Get summary aggregation of stock cost data
  ///
  /// Returns grouped data by itemcode with totals
  static Future<Map<String, dynamic>> getSummary({
    required String fromDate,
    required String toDate,
    List<String>? branchCodes,
    List<String>? productCodes,
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/stockcost/summary");

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
        return jsonDecode(response.body);
      } else {
        return {
          "status": "error",
          "code": response.statusCode,
          "message": "API call failed: ${response.reasonPhrase}",
        };
      }
    } catch (e) {
      return {
        "status": "error",
        "code": 500,
        "message": "Connection failed: $e",
      };
    } finally {
      httpClient.close();
    }
  }

  /// Check data availability for date range
  ///
  /// Returns counts and date ranges of available data
  static Future<Map<String, dynamic>> checkDataAvailability({
    required String fromDate,
    required String toDate,
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/stockcost/check");

      final body = {
        "shop_id": global.getShopId(),
        "from_date": fromDate,
        "to_date": toDate,
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        return jsonDecode(response.body);
      } else {
        return {
          "status": "error",
          "code": response.statusCode,
          "message": "API call failed: ${response.reasonPhrase}",
        };
      }
    } catch (e) {
      return {
        "status": "error",
        "code": 500,
        "message": "Connection failed: $e",
      };
    } finally {
      httpClient.close();
    }
  }
}
