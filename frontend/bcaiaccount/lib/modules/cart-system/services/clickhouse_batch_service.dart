import 'dart:convert';
import 'package:http/http.dart' as http;
import '../../../utils/logger/app_logger.dart';
import '../../../global.dart' as global;

/// Cache entry สำหรับเก็บข้อมูล
class CacheEntry {
  final List<Map<String, dynamic>> data;
  final DateTime timestamp;
  final Duration duration;

  CacheEntry({
    required this.data,
    required this.timestamp,
    required this.duration,
  });

  bool get isExpired => DateTime.now().difference(timestamp) > duration;
}

/// โครงสร้างสำหรับแต่ละ query
class BatchQuery {
  final String query;
  final Map<String, dynamic>? parameters;
  final Duration? cacheDuration;

  BatchQuery({required this.query, this.parameters, this.cacheDuration});

  Map<String, dynamic> toJson() => {
    'query': query,
    if (parameters != null) 'parameters': parameters,
  };
}

/// โครงสร้างสำหรับ response
class BatchResult {
  final int index;
  final String query;
  final String status;
  final int count;
  final List<Map<String, dynamic>> data;
  final String? error;

  BatchResult({
    required this.index,
    required this.query,
    required this.status,
    required this.count,
    required this.data,
    this.error,
  });

  factory BatchResult.fromJson(Map<String, dynamic> json) => BatchResult(
    index: json['index'] as int,
    query: json['query'] as String,
    status: json['status'] as String,
    count: json['count'] as int,
    data: (json['data'] as List<dynamic>).cast<Map<String, dynamic>>(),
    error: json['error'] as String?,
  );

  BatchResult copyWith({
    int? index,
    String? query,
    String? status,
    int? count,
    List<Map<String, dynamic>>? data,
    String? error,
  }) {
    return BatchResult(
      index: index ?? this.index,
      query: query ?? this.query,
      status: status ?? this.status,
      count: count ?? this.count,
      data: data ?? this.data,
      error: error ?? this.error,
    );
  }
}

class BatchResponse {
  final String status;
  final int code;
  final int totalQueries;
  final int successCount;
  final int errorCount;
  final List<BatchResult> results;

  BatchResponse({
    required this.status,
    required this.code,
    required this.totalQueries,
    required this.successCount,
    required this.errorCount,
    required this.results,
  });

  factory BatchResponse.fromJson(Map<String, dynamic> json) => BatchResponse(
    status: json['status'] as String,
    code: json['code'] as int,
    totalQueries: json['total_queries'] as int,
    successCount: json['success_count'] as int,
    errorCount: json['error_count'] as int,
    results: (json['results'] as List<dynamic>)
        .map((e) => BatchResult.fromJson(e as Map<String, dynamic>))
        .toList(),
  );
}

/// Service สำหรับจัดการ batch queries ใน ClickHouse
/// ช่วยลดจำนวน API calls และเพิ่มประสิทธิภาพ
class ClickHouseBatchService {
  final String baseUrl;

  ClickHouseBatchService({String? baseUrl})
    : baseUrl = baseUrl ?? global.goApiUrlPath('clickhouse');

  Map<String, String> get _headers => {'Content-Type': 'application/json'};

  /// สำหรับเก็บ cache ข้อมูลในหน่วยความจำ
  static final Map<String, CacheEntry> _cache = {};
  static const Duration _defaultCacheDuration = Duration(minutes: 5);

  /// สร้าง cache key สำหรับ query
  String _generateCacheKey(String query, Map<String, dynamic>? parameters) {
    final queryKey = query.replaceAll(RegExp(r'\s+'), ' ').trim();
    final paramKey = parameters?.toString() ?? '';
    return '${queryKey}_$paramKey';
  }

  /// ตรวจสอบว่า cache ยังใช้ได้ไหม
  List<Map<String, dynamic>>? _getFromCache(String key) {
    final entry = _cache[key];
    if (entry != null && !entry.isExpired) {
      AppLogger.debug('📦 [Batch Service] Cache hit for key: $key');
      return entry.data;
    }
    _cache.remove(key); // ลบ cache ที่หมดอายุ
    return null;
  }

  /// เก็บข้อมูลลง cache
  void _setCache(
    String key,
    List<Map<String, dynamic>> data,
    Duration? duration,
  ) {
    _cache[key] = CacheEntry(
      data: data,
      timestamp: DateTime.now(),
      duration: duration ?? _defaultCacheDuration,
    );
  }

  /// เรียกใช้ batch queries
  ///
  /// ใช้เพื่อลดจำนวน API calls สำหรับการค้นหาสินค้า
  Future<List<BatchResult>> executeBatchQueries(
    List<BatchQuery> queries, {
    bool enableCache = true,
  }) async {
    try {
      AppLogger.info(
        '🚀 [Batch Service] Executing ${queries.length} batch queries',
      );

      // ตรวจสอบ cache สำหรับแต่ละ query
      final cachedQueries = <int, List<Map<String, dynamic>>>{};
      final uncachedQueries = <int, BatchQuery>{};

      for (var i = 0; i < queries.length; i++) {
        final query = queries[i];
        if (enableCache) {
          final cacheKey = _generateCacheKey(query.query, query.parameters);
          final cachedData = _getFromCache(cacheKey);

          if (cachedData != null) {
            cachedQueries[i] = cachedData;
            continue;
          }
        }
        uncachedQueries[i] = query;
      }

      // ถ้าทุก query มี cache hit ส่งคืนทันที
      if (cachedQueries.length == queries.length) {
        AppLogger.info('🎯 [Batch Service] All queries hit cache');
        return queries.asMap().entries.map((entry) {
          final index = entry.key;
          final query = entry.value;
          final cachedData = cachedQueries[index]!;

          return BatchResult(
            index: index,
            query: query.query,
            status: 'success',
            count: cachedData.length,
            data: cachedData,
          );
        }).toList();
      }

      // สร้าง request payload
      final requestQueries = uncachedQueries.values
          .map((q) => q.toJson())
          .toList();

      final url = Uri.parse('$baseUrl/querys');
      AppLogger.debug('📡 [Batch Service] Request URL: $url');

      final response = await http.post(
        url,
        headers: _headers,
        body: jsonEncode({'queries': requestQueries}),
      );

      AppLogger.debug('📡 [Batch Service] Response: ${response.statusCode}');

      if (response.statusCode == 200) {
        final responseData = jsonDecode(response.body) as Map<String, dynamic>;
        final batchResponse = BatchResponse.fromJson(responseData);

        AppLogger.info(
          '✅ [Batch Service] Batch completed - Success: ${batchResponse.successCount}/${batchResponse.totalQueries}',
        );

        // รวมผลลัพธ์จาก API และ cache
        final results = <BatchResult>[];

        // เพิ่มผลลัพธ์จาก cache ก่อน
        for (var i = 0; i < queries.length; i++) {
          if (cachedQueries.containsKey(i)) {
            final query = queries[i];
            final cachedData = cachedQueries[i]!;

            results.add(
              BatchResult(
                index: i,
                query: query.query,
                status: 'success',
                count: cachedData.length,
                data: cachedData,
              ),
            );
          }
        }

        // เพิ่มผลลัพธ์จาก API response
        for (final result in batchResponse.results) {
          final originalIndex = uncachedQueries.keys.firstWhere(
            (key) => uncachedQueries[key]!.query == result.query,
          );

          results.add(result.copyWith(index: originalIndex));

          // เก็บข้อมูลลง cache ถ้า enableCache = true
          if (enableCache &&
              result.status == 'success' &&
              result.data.isNotEmpty) {
            final query = uncachedQueries[originalIndex]!;
            final cacheKey = _generateCacheKey(query.query, query.parameters);
            _setCache(cacheKey, result.data, query.cacheDuration);
          }
        }

        // เรียงลำดับตาม index เดิม
        results.sort((a, b) => a.index.compareTo(b.index));

        return results;
      } else {
        AppLogger.error(
          '❌ [Batch Service] Failed - Status: ${response.statusCode}, Body: ${response.body}',
        );
        throw Exception('Batch query failed: ${response.statusCode}');
      }
    } catch (e, stackTrace) {
      AppLogger.error(
        '❌ [Batch Service] Error executing batch queries',
        error: e,
        stackTrace: stackTrace,
      );
      throw Exception('Batch query error: $e');
    }
  }

  /// เคลียร์ cache (สำหรับการพัฒนา)
  void clearCache() {
    _cache.clear();
    AppLogger.info('🧹 [Batch Service] Cache cleared');
  }

  /// ตรวจสอบสถานะ cache
  Map<String, dynamic> getCacheStats() {
    var validCount = 0;
    var expiredCount = 0;

    _cache.forEach((key, entry) {
      if (entry.isExpired) {
        expiredCount++;
      } else {
        validCount++;
      }
    });

    return {
      'total_entries': _cache.length,
      'valid_entries': validCount,
      'expired_entries': expiredCount,
    };
  }

  /// ตั้งค่า cache duration สำหรับ query ประเภทต่างๆ
  static const Map<String, Duration> _cacheDurations = {
    'popular_products': Duration(minutes: 10),
    'product_units': Duration(minutes: 30),
    // 'product_balances': NO CACHE - ต้อง real-time
    'universal_search': Duration(minutes: 2),
  };

  /// สร้าง cache duration สำหรับ query type
  static Duration getCacheDurationForQueryType(String queryType) {
    return _cacheDurations[queryType] ?? _defaultCacheDuration;
  }
}

/// Extension สำหรับ BatchResult (เนื่องจาก copyWith ถูกย้ายไปใน class แล้ว)
/// สามารถลบ extension นี้ออกได้ เพราะ copyWith method ถูกย้ายไปใน BatchResult class แล้ว
/// extension BatchResultExtension on BatchResult {
//   BatchResult copyWith({
//     int? index,
//     String? query,
//     String? status,
//     int? count,
//     List<Map<String, dynamic>>? data,
//     String? error,
//   }) {
//     return BatchResult(
//       index: index ?? this.index,
//       query: query ?? this.query,
//       status: status ?? this.status,
//       count: count ?? this.count,
//       data: data ?? this.data,
//       error: error ?? this.error,
//     );
//   }
// }
