import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/lineoa_model.dart';

/// Line OA API Service
/// บริการเชื่อมต่อ Line OA ผ่าน Backend API
class LineOAApiService {
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

  /// ดึงรายการ Line OA Config ทั้งหมดของ shop
  static Future<LineOAConfigListResult> getConfigs() async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/lineoa/configs");

      final body = {
        "shop_id": global.getShopId(),
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          return LineOAConfigListResult.fromJson(json);
        }
      }

      return LineOAConfigListResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      return LineOAConfigListResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ดึง Line OA Config ตามประเภท
  static Future<LineOAConfigResult> getConfigByType(LineOAType type) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/lineoa/config");

      final body = {
        "shop_id": global.getShopId(),
        "lineoa_type": type.code,
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          return LineOAConfigResult.fromJson(json);
        }
      }

      if (response.statusCode == 404) {
        return LineOAConfigResult.notFound();
      }

      return LineOAConfigResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      return LineOAConfigResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// บันทึก Line OA Config
  static Future<LineOASaveResult> saveConfig(LineOAConfigModel config) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/lineoa/config/save");

      final Map<String, dynamic> body = {
        "shop_id": global.getShopId(),
        ...config.toJson(),
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          return LineOASaveResult.success(json['guid'] ?? '');
        }
      }

      return LineOASaveResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      return LineOASaveResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ทดสอบการเชื่อมต่อ Line OA
  static Future<LineOATestResult> testConnection({
    required String channelId,
    required String channelSecret,
    required String accessToken,
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/lineoa/test");

      final body = {
        "shop_id": global.getShopId(),
        "channel_id": channelId,
        "channel_secret": channelSecret,
        "access_token": accessToken,
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        return LineOATestResult.fromJson(json);
      }

      return LineOATestResult(
        success: false,
        message: "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      return LineOATestResult(
        success: false,
        message: "Connection failed: $e",
      );
    } finally {
      httpClient.close();
    }
  }

  /// ดึงรายการพนักงานที่เชื่อมต่อกับ Line OA
  static Future<LineOAEmployeeListResult> getEmployees(
      String lineOaConfigGuid) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/lineoa/employees");

      final body = {
        "shop_id": global.getShopId(),
        "lineoa_config_guid": lineOaConfigGuid,
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          return LineOAEmployeeListResult.fromJson(json);
        }
      }

      return LineOAEmployeeListResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      return LineOAEmployeeListResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// เพิ่มพนักงานเข้า Line OA
  static Future<LineOASaveResult> addEmployee({
    required String lineOaConfigGuid,
    required String employeeCode,
    required String employeeName,
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/lineoa/employee/add");

      final body = {
        "shop_id": global.getShopId(),
        "lineoa_config_guid": lineOaConfigGuid,
        "employee_code": employeeCode,
        "employee_name": employeeName,
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          return LineOASaveResult.success(json['guid'] ?? '');
        }
      }

      return LineOASaveResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      return LineOASaveResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ลบพนักงานออกจาก Line OA
  static Future<bool> removeEmployee(String employeeGuid) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/lineoa/employee/remove");

      final body = {
        "shop_id": global.getShopId(),
        "guid": employeeGuid,
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      return response.statusCode == 200;
    } catch (e) {
      return false;
    } finally {
      httpClient.close();
    }
  }

  /// สร้าง Link สำหรับ LIFF เพื่อให้พนักงานเชื่อมต่อ
  static Future<LineOALinkResult> generateEmployeeLink({
    required String lineOaConfigGuid,
    required String employeeCode,
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/lineoa/employee/link");

      final body = {
        "shop_id": global.getShopId(),
        "lineoa_config_guid": lineOaConfigGuid,
        "employee_code": employeeCode,
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          return LineOALinkResult.success(json['link'] ?? '');
        }
      }

      return LineOALinkResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      return LineOALinkResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }
}

/// ผลลัพธ์รายการ Config
class LineOAConfigListResult {
  final bool isSuccess;
  final String? errorMessage;
  final List<LineOAConfigModel> configs;

  LineOAConfigListResult({
    required this.isSuccess,
    this.errorMessage,
    this.configs = const [],
  });

  factory LineOAConfigListResult.fromJson(Map<String, dynamic> json) {
    final dataList = (json['data'] as List?)
            ?.map((item) => LineOAConfigModel.fromJson(item))
            .toList() ??
        [];

    return LineOAConfigListResult(
      isSuccess: true,
      configs: dataList,
    );
  }

  factory LineOAConfigListResult.error(String message) =>
      LineOAConfigListResult(
        isSuccess: false,
        errorMessage: message,
      );
}

/// ผลลัพธ์ Config เดียว
class LineOAConfigResult {
  final bool isSuccess;
  final bool isFound;
  final String? errorMessage;
  final LineOAConfigModel? config;

  LineOAConfigResult({
    required this.isSuccess,
    this.isFound = false,
    this.errorMessage,
    this.config,
  });

  factory LineOAConfigResult.fromJson(Map<String, dynamic> json) {
    return LineOAConfigResult(
      isSuccess: true,
      isFound: true,
      config: LineOAConfigModel.fromJson(json['data']),
    );
  }

  factory LineOAConfigResult.notFound() => LineOAConfigResult(
        isSuccess: true,
        isFound: false,
      );

  factory LineOAConfigResult.error(String message) => LineOAConfigResult(
        isSuccess: false,
        errorMessage: message,
      );
}

/// ผลลัพธ์การบันทึก
class LineOASaveResult {
  final bool isSuccess;
  final String? errorMessage;
  final String? guid;

  LineOASaveResult({
    required this.isSuccess,
    this.errorMessage,
    this.guid,
  });

  factory LineOASaveResult.success(String guid) => LineOASaveResult(
        isSuccess: true,
        guid: guid,
      );

  factory LineOASaveResult.error(String message) => LineOASaveResult(
        isSuccess: false,
        errorMessage: message,
      );
}

/// ผลลัพธ์รายการพนักงาน
class LineOAEmployeeListResult {
  final bool isSuccess;
  final String? errorMessage;
  final List<LineOAEmployeeModel> employees;

  LineOAEmployeeListResult({
    required this.isSuccess,
    this.errorMessage,
    this.employees = const [],
  });

  factory LineOAEmployeeListResult.fromJson(Map<String, dynamic> json) {
    final dataList = (json['data'] as List?)
            ?.map((item) => LineOAEmployeeModel.fromJson(item))
            .toList() ??
        [];

    return LineOAEmployeeListResult(
      isSuccess: true,
      employees: dataList,
    );
  }

  factory LineOAEmployeeListResult.error(String message) =>
      LineOAEmployeeListResult(
        isSuccess: false,
        errorMessage: message,
      );
}

/// ผลลัพธ์ Link
class LineOALinkResult {
  final bool isSuccess;
  final String? errorMessage;
  final String? link;

  LineOALinkResult({
    required this.isSuccess,
    this.errorMessage,
    this.link,
  });

  factory LineOALinkResult.success(String link) => LineOALinkResult(
        isSuccess: true,
        link: link,
      );

  factory LineOALinkResult.error(String message) => LineOALinkResult(
        isSuccess: false,
        errorMessage: message,
      );
}
