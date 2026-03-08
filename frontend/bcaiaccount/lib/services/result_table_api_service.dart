// API Service สำหรับจัดการ Result Table
// รองรับ 4 endpoints: /resultfromquery, /resultget, /resulttopdf, /get

import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:smlaicloud/model/result_table_models.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/global.dart' as global;

class ResultTableQueryAlias {
  final String value;
  const ResultTableQueryAlias._(this.value);

  static const documentHeader = ResultTableQueryAlias._('document_header');
  static const documentLines = ResultTableQueryAlias._('document_lines');

  static ResultTableQueryAlias of(String alias) =>
      ResultTableQueryAlias._(alias);
}

class LinkedQueryLinkConfig {
  final ResultTableQueryAlias parentAlias;
  final List<String> parentKeys;
  final List<String> childKeys;

  const LinkedQueryLinkConfig({
    required this.parentAlias,
    required this.parentKeys,
    required this.childKeys,
  }) : assert(
         parentKeys.length == childKeys.length,
         'parentKeys and childKeys must have the same length',
       );

  factory LinkedQueryLinkConfig.single({
    required ResultTableQueryAlias parentAlias,
    required String parentKey,
    required String childKey,
  }) => LinkedQueryLinkConfig(
    parentAlias: parentAlias,
    parentKeys: [parentKey],
    childKeys: [childKey],
  );

  Map<String, dynamic> toJson() => {
    'parent_alias': parentAlias.value,
    'parent_keys': parentKeys,
    'child_keys': childKeys,
  };
}

class LinkedQueryDefinition {
  final ResultTableQueryAlias alias;
  final String query;
  final Map<String, dynamic>? summaryConfig;
  final LinkedQueryLinkConfig? linkConfig;

  const LinkedQueryDefinition({
    required this.alias,
    required this.query,
    this.summaryConfig,
    this.linkConfig,
  });

  Map<String, dynamic> toJson() => {
    'alias': alias.value,
    'query': query,
    if (summaryConfig != null) 'summary_config': summaryConfig,
    if (linkConfig != null) 'link_config': linkConfig!.toJson(),
  };
}

class LinkedQueryBuilders {
  static LinkedQueryDefinition documentHeader(String query) =>
      LinkedQueryDefinition(
        alias: ResultTableQueryAlias.documentHeader,
        query: query,
        summaryConfig: ResultTableSummaryConfig.documentHeader(),
      );

  static LinkedQueryDefinition linkedChild({
    required String query,
    required ResultTableQueryAlias alias,
    required ResultTableQueryAlias parentAlias,
    required List<String> parentKeys,
    required List<String> childKeys,
  }) => LinkedQueryDefinition(
    alias: alias,
    query: query,
    linkConfig: LinkedQueryLinkConfig(
      parentAlias: parentAlias,
      parentKeys: parentKeys,
      childKeys: childKeys,
    ),
  );
}

class ResultTableSummaryConfig {
  static Map<String, dynamic> documentHeader() => {
    'levels': [
      {
        'group_by_fields': ['docdate'],
        'sum_fields': ['totalqty', 'totalamount', 'calcamount', 'grossprofit'],
        'typejson': 1,
      },
    ],
    'grand_total': true,
    'grand_total_type': 99,
  };
}

class ResultTableApiService {
  /// ดึง Base URL จาก environment configuration (เหมือน BiReport)
  static String get baseUrl => global.goApiBaseUrl;

  static String _trimTrailingSlash(String url) {
    if (url.endsWith('/')) {
      return url.substring(0, url.length - 1);
    }
    return url;
  }

  /// Timeout duration
  static const Duration timeout = Duration(minutes: 5);

  /// Headers สำหรับ API requests (รวม Authorization token)
  static Map<String, String> _getHeaders() {
    final headers = <String, String>{
      'Content-Type': 'application/json; charset=UTF-8',
    };
    // เพิ่ม Authorization token ถ้ามี
    final token = global.appConfig.getString("token") ?? '';
    if (token.isNotEmpty) {
      headers['Authorization'] = 'Bearer $token';
    }
    return headers;
  }

  /// Base URL for the Postgres select endpoint (/get)
  /// ใช้ goApiUrl ถ้ามี (local dev mode) เพราะ /get endpoint อยู่ใน goapi
  static String get _coreApiBaseUrl => _resolveCoreApiBaseUrl();

  static String _resolveCoreApiBaseUrl() {
    String rawHost;
    String portOverride;

    // ใช้ goApiUrl ถ้ามี (local dev mode) เพราะ /get endpoint อยู่ใน goapi
    if (global.myAppConfig.goApiUrl.isNotEmpty) {
      rawHost = global.myAppConfig.goApiUrl.trim();
      portOverride = global.myAppConfig.goApiPort.trim();
    } else {
      rawHost = global.myAppConfig.serviceApi.trim();
      portOverride = global.myAppConfig.servicePort.trim();
    }

    if (rawHost.isEmpty) {
      throw StateError('API URL is not configured');
    }

    final hasScheme =
        rawHost.startsWith('http://') || rawHost.startsWith('https://');
    final normalized = hasScheme ? rawHost : 'https://$rawHost';
    final parsed = Uri.tryParse(normalized);
    if (parsed == null || parsed.host.isEmpty) {
      throw StateError('Invalid API configuration: $rawHost');
    }

    if (portOverride.isNotEmpty && !parsed.hasPort) {
      final parsedPort = int.tryParse(portOverride);
      if (parsedPort != null) {
        return _trimTrailingSlash(parsed.replace(port: parsedPort).toString());
      }
    }

    return _trimTrailingSlash(parsed.toString());
  }

  /// 📌 Handler 1: Execute query และเก็บผลลัพธ์ใน result table
  ///
  /// รับ shopid, query, guid (optional)
  /// - Execute SELECT query และเก็บผลลัพธ์ใน result table
  /// - สร้าง GUID ใหม่ถ้าไม่ได้ส่งมา
  /// - จำกัดไม่เกิน 10,000 rows
  /// - ใช้ timeout 5 นาที
  static Future<ResultFromQueryResponse> executeQueryAndSave({
    required String shopid,
    required List<LinkedQueryDefinition> queryDefinitions,
    String? guid,
  }) async {
    try {
      AppLogger.info('ResultTableApi: executeQueryAndSave - shopid=$shopid');

      if (queryDefinitions.isEmpty) {
        throw ArgumentError('queryDefinitions must not be empty');
      }

      final queryItems = queryDefinitions.map((def) => def.toJson()).toList();

      final requestBody = {
        'shopid': shopid,
        if (guid != null) 'guid': guid,
        'query_items': queryItems,
      };

      final url = '$baseUrl/resultfromquery';
      AppLogger.debug('🌐 Calling $url');
      AppLogger.debug('📦 Request: $requestBody');

      final response = await http
          .post(
            Uri.parse(url),
            headers: _getHeaders(),
            body: jsonEncode(requestBody),
          )
          .timeout(timeout);

      if (response.statusCode == 200) {
        final result = ResultFromQueryResponse.fromJson(
          jsonDecode(response.body) as Map<String, dynamic>,
        );
        AppLogger.info(
          'ResultTableApi: Query executed successfully - guid=${result.guid}, count=${result.count}',
        );
        return result;
      } else {
        throw Exception('API Error: ${response.statusCode} - ${response.body}');
      }
    } catch (e) {
      AppLogger.error('ResultTableApi: executeQueryAndSave failed - $e');
      rethrow;
    }
  }

  /// 📌 Handler 2: ดึงข้อมูลจาก result table พร้อม pagination
  ///
  /// รับ shopid, guid, limit, offset
  /// - Default: limit=100, offset=0
  /// - Max limit: 1,000 rows ต่อครั้ง
  /// - Return ทั้ง data และ pagination info
  static Future<ResultGetPageResponse> getResults({
    required String shopid,
    required String guid,
    int limit = 100,
    int offset = 0,
  }) async {
    try {
      final stopwatch = Stopwatch()..start();
      AppLogger.info(
        'ResultTableApi: getResults request | shopid=$shopid, guid=$guid, limit=$limit, offset=$offset, baseUrl=$baseUrl',
      );

      final requestBody = {
        'shopid': shopid,
        'guid': guid,
        'limit': limit,
        'offset': offset,
      };

      final url = '$baseUrl/resultget';
      AppLogger.debug('🌐 Calling $url');
      AppLogger.debug('📦 Request: $requestBody');

      final response = await http
          .post(
            Uri.parse(url),
            headers: _getHeaders(),
            body: jsonEncode(requestBody),
          )
          .timeout(timeout);

      if (response.statusCode == 200) {
        final result = ResultGetPageResponse.fromJson(
          jsonDecode(response.body) as Map<String, dynamic>,
        );
        stopwatch.stop();
        AppLogger.info(
          'ResultTableApi: getResults success | rows=${result.data.length}, total=${result.pagination.total}, limit=${result.pagination.limit}, offset=${result.pagination.offset}, duration=${stopwatch.elapsedMilliseconds}ms',
        );
        return result;
      } else {
        throw Exception('API Error: ${response.statusCode} - ${response.body}');
      }
    } catch (e) {
      AppLogger.error(
        'ResultTableApi: getResults failed | shopid=$shopid, guid=$guid, limit=$limit, offset=$offset, error=$e',
      );
      rethrow;
    }
  }

  /// 📌 Handler 3: สร้าง PDF จากข้อมูลใน result table
  ///
  /// รับ shopid, guid, pdf_config, column_order, column_names
  /// - รองรับ Thai font (THSarabunNew.ttf)
  /// - กำหนดลำดับและชื่อคอลัมน์ได้
  /// - Return PDF bytes
  static Future<List<int>> generatePdf({
    required String shopid,
    required String guid,
    required PdfConfig pdfConfig,
    required List<String> columnOrder,
    required Map<String, String> columnNames,
    Map<String, dynamic>? layoutConfig,
  }) async {
    try {
      AppLogger.info('ResultTableApi: generatePdf - guid=$guid');

      final request = ResultToPdfRequest(
        shopid: shopid,
        guid: guid,
        pdfConfig: pdfConfig,
        columnOrder: columnOrder,
        columnNames: columnNames,
        layoutConfig: layoutConfig,
      );

      final url = '$baseUrl/resulttopdf';
      AppLogger.debug('🌐 Calling $url');
      AppLogger.debug('📦 Request: ${request.toJson()}');

      final response = await http
          .post(
            Uri.parse(url),
            headers: _getHeaders(),
            body: jsonEncode(request.toJson()),
          )
          .timeout(timeout);

      if (response.statusCode == 200) {
        AppLogger.info(
          'ResultTableApi: PDF generated successfully (${response.bodyBytes.length} bytes)',
        );
        return response.bodyBytes;
      } else {
        throw Exception('API Error: ${response.statusCode} - ${response.body}');
      }
    } catch (e) {
      AppLogger.error('ResultTableApi: generatePdf failed - $e');
      rethrow;
    }
  }

  /// 📌 Handler 4: Execute ad-hoc SELECT queries via /get
  static Future<Map<String, dynamic>> executeSelectQueries({
    required String database,
    required List<String> queries,
  }) async {
    final trimmedDb = database.trim();
    if (trimmedDb.isEmpty) {
      throw ArgumentError('database must not be empty');
    }
    if (queries.isEmpty) {
      throw ArgumentError('queries must not be empty');
    }

    final url = '$_coreApiBaseUrl/get';
    final payload = {'database': trimmedDb, 'queries': queries};

    try {
      AppLogger.info(
        'ResultTableApi: executeSelectQueries | database=$trimmedDb, queries=${queries.length}',
      );
      AppLogger.debug('🌐 Calling $url');
      AppLogger.debug('📦 Request: $payload');

      final response = await http
          .post(
            Uri.parse(url),
            headers: _getHeaders(),
            body: jsonEncode(payload),
          )
          .timeout(timeout);

      if (response.statusCode == 200) {
        final decoded = jsonDecode(response.body) as Map<String, dynamic>;
        AppLogger.info(
          'ResultTableApi: executeSelectQueries success | status=${decoded['status']}, total=${decoded['total']}',
        );
        return decoded;
      } else {
        throw Exception('API Error: ${response.statusCode} - ${response.body}');
      }
    } catch (e) {
      AppLogger.error('ResultTableApi: executeSelectQueries failed - $e');
      rethrow;
    }
  }

  /// Helper: ดึงข้อมูลทั้งหมด (รวม pagination หลายรอบ)
  static Future<List<Map<String, dynamic>>> getAllResults({
    required String shopid,
    required String guid,
    int pageSize = 100,
  }) async {
    final allData = <Map<String, dynamic>>[];
    int offset = 0;
    bool hasMore = true;

    while (hasMore) {
      final response = await getResults(
        shopid: shopid,
        guid: guid,
        limit: pageSize,
        offset: offset,
      );

      allData.addAll(response.data);
      offset += response.data.length;

      // ถ้าได้ข้อมูลน้อยกว่า page size หรือครบทั้งหมดแล้ว ให้หยุด
      hasMore =
          response.data.length >= pageSize &&
          offset < response.pagination.total;
    }

    AppLogger.info('ResultTableApi: Retrieved all ${allData.length} rows');
    return allData;
  }
}
