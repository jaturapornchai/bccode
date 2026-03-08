import 'dart:convert';

import 'package:http/http.dart' as http;

import '../../../global.dart' as global;
import '../../../utils/logger/app_logger.dart';
import '../models/product_search_response.dart';

/// Service สำหรับค้นหาสินค้าจาก Backend Unified API
/// ทุกอย่างทำที่ backend: tokenize, search, score, balance, format packing
class ProductSearchApi {
  final String? baseUrl;
  final Duration timeout;

  ProductSearchApi({
    this.baseUrl,
    this.timeout = const Duration(seconds: 30),
  });

  /// URL ของ API
  String get _apiUrl {
    if (baseUrl != null && baseUrl!.isNotEmpty) {
      return baseUrl!;
    }
    return global.goApiBaseUrl;
  }

  /// ค้นหาสินค้าแบบ Unified (ทุกอย่างทำที่ backend)
  /// - Thai tokenization ทำที่ backend
  /// - Search & Score ทำที่ backend
  /// - Stock balance ทำที่ backend
  /// - Multi-level packing format ทำที่ backend
  /// - รองรับ pagination (offset) สำหรับ infinite scroll
  Future<ProductSearchResponse> search({
    required String shopId,
    required String keyword,
    String? whcode,
    String? locationcode,
    int limit = 200,
    int offset = 0, // สำหรับ infinite scroll
    bool includeBalance = true,
  }) async {
    final url = Uri.parse('$_apiUrl/api/product/search/unified');

    AppLogger.info(
      '🔍 [ProductSearchApi] Searching: keyword="$keyword", shopId=$shopId, limit=$limit, offset=$offset',
    );

    try {
      final sw = Stopwatch()..start();

      final response = await http
          .post(
            url,
            headers: {'Content-Type': 'application/json'},
            body: jsonEncode({
              'shopid': shopId,
              'keyword': keyword,
              'whcode': whcode ?? '',
              'locationcode': locationcode ?? '',
              'limit': limit,
              'offset': offset,
              'include_balance': includeBalance,
            }),
          )
          .timeout(timeout);

      final httpMs = sw.elapsedMilliseconds;

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body) as Map<String, dynamic>;
        final parseMs1 = sw.elapsedMilliseconds - httpMs;

        final result = ProductSearchResponse.fromJson(json);
        final parseMs2 = sw.elapsedMilliseconds - httpMs - parseMs1;

        AppLogger.info(
          '⏱️ [ProductSearchApi] HTTP=${httpMs}ms, jsonDecode=${parseMs1}ms, fromJson=${parseMs2}ms, total=${sw.elapsedMilliseconds}ms',
        );
        AppLogger.info(
          '✅ [ProductSearchApi] Found ${result.count} products (total: ${result.total}, hasMore: ${result.hasMore})',
        );

        return result;
      } else {
        AppLogger.error(
          '❌ [ProductSearchApi] HTTP ${response.statusCode}: ${response.body}',
        );
        throw Exception('Search failed: HTTP ${response.statusCode}');
      }
    } catch (e) {
      AppLogger.error('❌ [ProductSearchApi] Error: $e');
      rethrow;
    }
  }

  /// ค้นหาด้วย barcode โดยตรง
  Future<ProductSearchResponse> searchByBarcode({
    required String shopId,
    required String barcode,
    bool includeBalance = true,
  }) async {
    return search(
      shopId: shopId,
      keyword: barcode,
      includeBalance: includeBalance,
      limit: 10, // barcode search ควรได้ผลน้อย
    );
  }

  /// ค้นหาด้วย itemcode โดยตรง
  Future<ProductSearchResponse> searchByItemCode({
    required String shopId,
    required String itemCode,
    bool includeBalance = true,
  }) async {
    return search(
      shopId: shopId,
      keyword: itemCode,
      includeBalance: includeBalance,
      limit: 10,
    );
  }
}
