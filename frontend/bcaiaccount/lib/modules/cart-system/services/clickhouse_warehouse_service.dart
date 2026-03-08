import 'dart:convert';
import 'package:http/http.dart' as http;
import '../../../utils/logger/app_logger.dart';
import '../../../global.dart' as global;
import '../models/clickhouse_warehouse_model.dart';
import '../models/clickhouse_location_model.dart';

/// Service สำหรับดึงข้อมูลคลังและที่เก็บจาก ClickHouse
class ClickHouseWarehouseService {
  final String baseUrl;

  ClickHouseWarehouseService({String? baseUrl})
      : baseUrl = baseUrl ?? global.goApiUrlPath('clickhouse');

  Map<String, String> get _headers => {'Content-Type': 'application/json'};

  /// ค้นหาคลังสินค้าตามคำค้นหา (code, name1, name2)
  Future<List<ClickHouseWarehouseModel>> searchWarehouses({
    required String shopId,
    String keyword = '',
    int limit = 100,
  }) async {
    try {
      AppLogger.info('🔍 Searching warehouses: keyword="$keyword", shop=$shopId');

      String query;
      if (keyword.isEmpty) {
        // ดึงทั้งหมด
        query = '''
          SELECT shopid, guidfixed, code, name1, name2
          FROM ${global.clickHouseDatabaseName}.warehouses
          WHERE shopid = '$shopId'
          ORDER BY code
          LIMIT $limit
        ''';
      } else {
        // ค้นหาตาม code, name1, name2
        query = '''
          SELECT shopid, guidfixed, code, name1, name2
          FROM ${global.clickHouseDatabaseName}.warehouses
          WHERE shopid = '$shopId'
            AND (
              lower(code) LIKE lower('%$keyword%')
              OR lower(name1) LIKE lower('%$keyword%')
              OR lower(name2) LIKE lower('%$keyword%')
            )
          ORDER BY code
          LIMIT $limit
        ''';
      }

      final url = Uri.parse('$baseUrl/query');
      final response = await http.post(
        url,
        headers: _headers,
        body: jsonEncode({'query': query}),
      );

      AppLogger.debug('📥 ClickHouse Response: ${response.statusCode}');

      if (response.statusCode == 200) {
        final responseData = jsonDecode(response.body) as Map<String, dynamic>;
        final status = responseData['status'] as String?;
        final data = responseData['data'] as List<dynamic>?;

        if (status == 'success' && data != null) {
          final warehouses = data
              .map((item) => ClickHouseWarehouseModel.fromJson(item as Map<String, dynamic>))
              .toList();

          AppLogger.info('✅ Found ${warehouses.length} warehouse(s)');
          return warehouses;
        }

        AppLogger.warning('⚠️ No warehouses found');
        return [];
      } else {
        throw Exception('Failed to search warehouses: ${response.statusCode}');
      }
    } catch (e, stackTrace) {
      AppLogger.error(
        '❌ Error searching warehouses',
        error: e,
        stackTrace: stackTrace,
      );
      return [];
    }
  }

  /// ดึงคลังสินค้าตาม guidfixed
  Future<ClickHouseWarehouseModel?> getWarehouseById({
    required String shopId,
    required String guidFixed,
  }) async {
    try {
      AppLogger.info('🔍 Getting warehouse by ID: $guidFixed');

      final query = '''
        SELECT shopid, guidfixed, code, name1, name2
        FROM ${global.clickHouseDatabaseName}.warehouses
        WHERE shopid = '$shopId'
          AND guidfixed = '$guidFixed'
        LIMIT 1
      ''';

      final url = Uri.parse('$baseUrl/query');
      final response = await http.post(
        url,
        headers: _headers,
        body: jsonEncode({'query': query}),
      );

      if (response.statusCode == 200) {
        final responseData = jsonDecode(response.body) as Map<String, dynamic>;
        final status = responseData['status'] as String?;
        final data = responseData['data'] as List<dynamic>?;

        if (status == 'success' && data != null && data.isNotEmpty) {
          return ClickHouseWarehouseModel.fromJson(data[0] as Map<String, dynamic>);
        }
      }

      return null;
    } catch (e, stackTrace) {
      AppLogger.error(
        '❌ Error getting warehouse by ID',
        error: e,
        stackTrace: stackTrace,
      );
      return null;
    }
  }

  /// ดึงที่เก็บทั้งหมดของคลังสินค้า
  Future<List<ClickHouseLocationModel>> getLocationsByWarehouse({
    required String shopId,
    required String warehouseCode,
    int limit = 100,
  }) async {
    try {
      AppLogger.info('🔍 Getting locations for warehouse: $warehouseCode');

      final query = '''
        SELECT shopid, guidfixed, locationcode, warehousecode, name1, name2
        FROM ${global.clickHouseDatabaseName}.locations
        WHERE shopid = '$shopId'
          AND warehousecode = '$warehouseCode'
        ORDER BY locationcode
        LIMIT $limit
      ''';

      final url = Uri.parse('$baseUrl/query');
      final response = await http.post(
        url,
        headers: _headers,
        body: jsonEncode({'query': query}),
      );

      AppLogger.debug('📥 ClickHouse Response: ${response.statusCode}');

      if (response.statusCode == 200) {
        final responseData = jsonDecode(response.body) as Map<String, dynamic>;
        final status = responseData['status'] as String?;
        final data = responseData['data'] as List<dynamic>?;

        if (status == 'success' && data != null) {
          final locations = data
              .map((item) => ClickHouseLocationModel.fromJson(item as Map<String, dynamic>))
              .toList();

          AppLogger.info('✅ Found ${locations.length} location(s)');
          return locations;
        }

        AppLogger.warning('⚠️ No locations found');
        return [];
      } else {
        throw Exception('Failed to get locations: ${response.statusCode}');
      }
    } catch (e, stackTrace) {
      AppLogger.error(
        '❌ Error getting locations',
        error: e,
        stackTrace: stackTrace,
      );
      return [];
    }
  }

  /// ดึงที่เก็บตาม guidfixed
  Future<ClickHouseLocationModel?> getLocationById({
    required String shopId,
    required String guidFixed,
  }) async {
    try {
      AppLogger.info('🔍 Getting location by ID: $guidFixed');

      final query = '''
        SELECT shopid, guidfixed, locationcode, warehousecode, name1, name2
        FROM ${global.clickHouseDatabaseName}.locations
        WHERE shopid = '$shopId'
          AND guidfixed = '$guidFixed'
        LIMIT 1
      ''';

      final url = Uri.parse('$baseUrl/query');
      final response = await http.post(
        url,
        headers: _headers,
        body: jsonEncode({'query': query}),
      );

      if (response.statusCode == 200) {
        final responseData = jsonDecode(response.body) as Map<String, dynamic>;
        final status = responseData['status'] as String?;
        final data = responseData['data'] as List<dynamic>?;

        if (status == 'success' && data != null && data.isNotEmpty) {
          return ClickHouseLocationModel.fromJson(data[0] as Map<String, dynamic>);
        }
      }

      return null;
    } catch (e, stackTrace) {
      AppLogger.error(
        '❌ Error getting location by ID',
        error: e,
        stackTrace: stackTrace,
      );
      return null;
    }
  }

  /// ค้นหาที่เก็บทั้งหมด (ไม่จำกัดคลัง)
  Future<List<ClickHouseLocationModel>> searchLocations({
    required String shopId,
    String keyword = '',
    int limit = 100,
  }) async {
    try {
      AppLogger.info('🔍 Searching locations: keyword="$keyword", shop=$shopId');

      String query;
      if (keyword.isEmpty) {
        // ดึงทั้งหมด
        query = '''
          SELECT shopid, guidfixed, locationcode, warehousecode, name1, name2
          FROM ${global.clickHouseDatabaseName}.locations
          WHERE shopid = '$shopId'
          ORDER BY warehousecode, locationcode
          LIMIT $limit
        ''';
      } else {
        // ค้นหาตาม locationcode, name1, name2
        query = '''
          SELECT shopid, guidfixed, locationcode, warehousecode, name1, name2
          FROM ${global.clickHouseDatabaseName}.locations
          WHERE shopid = '$shopId'
            AND (
              lower(locationcode) LIKE lower('%$keyword%')
              OR lower(name1) LIKE lower('%$keyword%')
              OR lower(name2) LIKE lower('%$keyword%')
            )
          ORDER BY warehousecode, locationcode
          LIMIT $limit
        ''';
      }

      final url = Uri.parse('$baseUrl/query');
      final response = await http.post(
        url,
        headers: _headers,
        body: jsonEncode({'query': query}),
      );

      AppLogger.debug('📥 ClickHouse Response: ${response.statusCode}');

      if (response.statusCode == 200) {
        final responseData = jsonDecode(response.body) as Map<String, dynamic>;
        final status = responseData['status'] as String?;
        final data = responseData['data'] as List<dynamic>?;

        if (status == 'success' && data != null) {
          final locations = data
              .map((item) => ClickHouseLocationModel.fromJson(item as Map<String, dynamic>))
              .toList();

          AppLogger.info('✅ Found ${locations.length} location(s)');
          return locations;
        }

        AppLogger.warning('⚠️ No locations found');
        return [];
      } else {
        throw Exception('Failed to search locations: ${response.statusCode}');
      }
    } catch (e, stackTrace) {
      AppLogger.error(
        '❌ Error searching locations',
        error: e,
        stackTrace: stackTrace,
      );
      return [];
    }
  }
}
