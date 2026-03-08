import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/debtor_model.dart';
import '../models/creditor_model.dart';
import '../../../utils/logger/app_logger.dart';
import '../../../global.dart' as global;

/// Service สำหรับค้นหาลูกหนี้และเจ้าหนี้จาก ClickHouse
class ClickHouseDebtorCreditorService {
  final String baseUrl;

  ClickHouseDebtorCreditorService({String? baseUrl})
      : baseUrl = baseUrl ?? global.goApiUrlPath('clickhouse');

  Map<String, String> get _headers => {'Content-Type': 'application/json'};

  /// ค้นหาลูกหนี้ (Debtor) - ค้นหาจากชื่อหรือรหัส
  ///
  /// [keyword] - คำค้นหา (ค้นใน code, name1, name2)
  /// [shopId] - รหัสร้านค้า
  /// [limit] - จำนวนผลลัพธ์สูงสุด (default: 20)
  Future<List<DebtorModel>> searchDebtors({
    required String keyword,
    required String shopId,
    int limit = 20,
  }) async {
    try {
      AppLogger.info('🔍 Searching debtors: keyword="$keyword", shop=$shopId');

      // สร้าง SQL query ค้นหาจากชื่อหรือรหัส
      final query = '''
        SELECT shopid, guidfixed, code, name1, name2
        FROM ${global.clickHouseDatabaseName}.debtors
        WHERE shopid = '$shopId'
          AND (
            lower(code) LIKE lower('%$keyword%')
            OR lower(name1) LIKE lower('%$keyword%')
            OR lower(name2) LIKE lower('%$keyword%')
          )
        ORDER BY code ASC
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
          final debtors = data
              .map((item) => DebtorModel.fromJson(item as Map<String, dynamic>))
              .toList();

          AppLogger.info('✅ Found ${debtors.length} debtor(s)');
          return debtors;
        }

        AppLogger.warning('⚠️ No debtors found');
        return [];
      } else {
        throw Exception(
          'Failed to search debtors: ${response.statusCode}',
        );
      }
    } catch (e, stackTrace) {
      AppLogger.error(
        '❌ Error searching debtors',
        error: e,
        stackTrace: stackTrace,
      );
      rethrow;
    }
  }

  /// ค้นหาเจ้าหนี้ (Creditor) - ค้นหาจากชื่อหรือรหัส
  ///
  /// [keyword] - คำค้นหา (ค้นใน code, name1, name2)
  /// [shopId] - รหัสร้านค้า
  /// [limit] - จำนวนผลลัพธ์สูงสุด (default: 20)
  Future<List<CreditorModel>> searchCreditors({
    required String keyword,
    required String shopId,
    int limit = 20,
  }) async {
    try {
      AppLogger.info('🔍 Searching creditors: keyword="$keyword", shop=$shopId');

      // สร้าง SQL query ค้นหาจากชื่อหรือรหัส
      final query = '''
        SELECT shopid, guidfixed, code, name1, name2
        FROM ${global.clickHouseDatabaseName}.creditors
        WHERE shopid = '$shopId'
          AND (
            lower(code) LIKE lower('%$keyword%')
            OR lower(name1) LIKE lower('%$keyword%')
            OR lower(name2) LIKE lower('%$keyword%')
          )
        ORDER BY code ASC
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
          final creditors = data
              .map(
                (item) => CreditorModel.fromJson(item as Map<String, dynamic>),
              )
              .toList();

          AppLogger.info('✅ Found ${creditors.length} creditor(s)');
          return creditors;
        }

        AppLogger.warning('⚠️ No creditors found');
        return [];
      } else {
        throw Exception(
          'Failed to search creditors: ${response.statusCode}',
        );
      }
    } catch (e, stackTrace) {
      AppLogger.error(
        '❌ Error searching creditors',
        error: e,
        stackTrace: stackTrace,
      );
      rethrow;
    }
  }

  /// ดึงข้อมูลลูกหนี้ตาม guidfixed
  Future<DebtorModel?> getDebtorById({
    required String guidFixed,
    required String shopId,
  }) async {
    try {
      final query = '''
        SELECT shopid, guidfixed, code, name1, name2
        FROM ${global.clickHouseDatabaseName}.debtors
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
        final data = responseData['data'] as List<dynamic>?;

        if (data != null && data.isNotEmpty) {
          return DebtorModel.fromJson(data.first as Map<String, dynamic>);
        }
      }

      return null;
    } catch (e) {
      AppLogger.error('❌ Error getting debtor by id: $e');
      return null;
    }
  }

  /// ดึงข้อมูลเจ้าหนี้ตาม guidfixed
  Future<CreditorModel?> getCreditorById({
    required String guidFixed,
    required String shopId,
  }) async {
    try {
      final query = '''
        SELECT shopid, guidfixed, code, name1, name2
        FROM ${global.clickHouseDatabaseName}.creditors
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
        final data = responseData['data'] as List<dynamic>?;

        if (data != null && data.isNotEmpty) {
          return CreditorModel.fromJson(data.first as Map<String, dynamic>);
        }
      }

      return null;
    } catch (e) {
      AppLogger.error('❌ Error getting creditor by id: $e');
      return null;
    }
  }
}
