import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:smlaicloud/global.dart' as global;

/// Service สำหรับดึงข้อมูล warehouse/location สำหรับรายงานสต็อก
/// ใช้ Backend API แทนการส่ง raw SQL (ป้องกัน SQL injection)
class StockReportLookupService {
  static String get _baseUrl => global.goApiBaseUrl;

  static Map<String, String> get _headers {
    final headers = <String, String>{'Content-Type': 'application/json'};
    final token = global.appConfig.getString("token") ?? '';
    if (token.isNotEmpty) {
      headers['Authorization'] = 'Bearer $token';
    }
    return headers;
  }

  /// ดึง barcodes จาก item codes
  static Future<List<String>> getBarcodesByItemCodes(
      List<String> itemCodes) async {
    if (itemCodes.isEmpty) return [];

    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/stock-report/barcodes");
      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode({
          "shop_id": global.getShopId(),
          "item_codes": itemCodes,
        }),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          return List<String>.from(json['barcodes'] ?? []);
        }
      }
      return [];
    } catch (e) {
      return [];
    } finally {
      httpClient.close();
    }
  }

  /// ดึง warehouses พร้อม locations ตาม barcodes
  static Future<List<WarehouseWithLocations>> getWarehousesAndLocations(
      List<String> barcodes) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/stock-report/warehouses");
      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode({
          "shop_id": global.getShopId(),
          "barcodes": barcodes,
        }),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          final warehouses = json['warehouses'] as List? ?? [];
          return warehouses
              .map((w) => WarehouseWithLocations(
                    whcode: w['whcode'] ?? '',
                    locations: List<String>.from(w['locations'] ?? []),
                  ))
              .toList();
        }
      }
      return [];
    } catch (e) {
      return [];
    } finally {
      httpClient.close();
    }
  }
}

/// Model สำหรับ warehouse พร้อม locations
class WarehouseWithLocations {
  final String whcode;
  final List<String> locations;

  WarehouseWithLocations({required this.whcode, required this.locations});
}
