import 'package:smlaicloud/environment.dart';
import 'package:smlaicloud/imports_bloc.dart';
import 'package:smlaicloud/utils/backend_url_manager.dart';
import 'package:smlaicloud/service_locator.dart';
import 'package:dio/dio.dart';
import 'package:smlaicloud/global.dart' as global;

class Client {
  Dio init() {
    Dio dio = Dio();
    dio.interceptors.add(ApiInterceptors());

    // mainapi URL มาจาก BackendUrlManager (ดึง origin จาก goapi URL)
    final mainApiUrl = BackendUrlManager.getMainApiUrl();
    if (mainApiUrl == null) {
      throw Exception('Backend URL ยังไม่ได้ตั้งค่า กรุณาตั้งค่าในหน้า Login');
    }
    String endPointService = mainApiUrl;

    endPointService += endPointService[endPointService.length - 1] == "/"
        ? ""
        : "/";

    dio.options.baseUrl = endPointService;

    return dio;
  }

  // Specialized client for BI Report API
  Dio initBiReport() {
    Dio dio = Dio();
    dio.interceptors.add(ApiInterceptors());

    String endPointService = Environment().config.reportBiApi;

    endPointService += endPointService[endPointService.length - 1] == "/"
        ? ""
        : "/";

    dio.options.baseUrl = endPointService;
    // _dio.options.connectTimeout = 20000; //20s
    //_dio.options.receiveTimeout = 5000 as Duration?; //5s

    return dio;
  }
}

class ApiResponse<T> {
  late final bool success;
  late final bool error;
  // ignore: unnecessary_question_mark
  late final dynamic? data;
  late final String message;
  late final code;
  final Page? page;
  late final String id;
  final String? token;
  final String? uri;

  ApiResponse({
    required this.success,
    required this.data,
    this.id = "",
    this.error = true,
    required this.message,
    this.code = 00,
    this.page,
    this.token = "",
    this.uri,
  });

  factory ApiResponse.fromMap(Map<String, dynamic> map) {
    return ApiResponse(
      success: map['success'] ?? false,
      message: map['message'] ?? "",
      error: map['error'] ?? true,
      data: map['data'],
      id: map['id'] ?? "",
      page: map['pagination'] == null
          ? Page.empty
          : Page.fromMap(map['pagination']),
      token: map['token'],
      uri: map['uri'],
    );
  }
}

class GoApiQueryManyResultResponse<T> {
  late final int status;
  late final dynamic results;
  late final int total;

  GoApiQueryManyResultResponse({
    required this.status,
    required this.results,
    required this.total,
  });

  // Helper function สำหรับแปลง dynamic เป็น int อย่างปลอดภัย (รองรับทั้ง int และ String)
  static int _safeParseInt(dynamic value, int defaultValue) {
    if (value == null) return defaultValue;
    if (value is int) return value;
    if (value is num) return value.toInt();
    if (value is String) return int.tryParse(value) ?? defaultValue;
    return defaultValue;
  }

  factory GoApiQueryManyResultResponse.fromMap(Map<String, dynamic> map) {
    return GoApiQueryManyResultResponse(
      status: _safeParseInt(map['status'], 0),
      results: map['results'] ?? [],
      total: _safeParseInt(map['total'], 0),
    );
  }
}

class GoApiQueryOneResultResponse<T> {
  late final int status;
  late final dynamic result;

  GoApiQueryOneResultResponse({required this.status, required this.result});

  // Helper function สำหรับแปลง dynamic เป็น int อย่างปลอดภัย (รองรับทั้ง int และ String)
  static int _safeParseInt(dynamic value, int defaultValue) {
    if (value == null) return defaultValue;
    if (value is int) return value;
    if (value is num) return value.toInt();
    if (value is String) return int.tryParse(value) ?? defaultValue;
    return defaultValue;
  }

  factory GoApiQueryOneResultResponse.fromMap(Map<String, dynamic> map) {
    return GoApiQueryOneResultResponse(
      status: _safeParseInt(map['status'], 0),
      result: map['result'] ?? {},
    );
  }
}

class Page {
  final int perPage;
  final int page;
  final int total;
  final int totalPage;

  const Page({
    required this.perPage,
    required this.page,
    required this.total,
    required this.totalPage,
  });

  static const empty = Page(perPage: 0, page: 0, total: 0, totalPage: 0);

  bool get isEmpty => this == Page.empty;

  bool get isNotEmpty => this == Page.empty;

  factory Page.fromMap(Map<String, dynamic> map) {
    return Page(
      perPage: map['perPage'],
      page: map['page'],
      total: map['total'],
      totalPage: map['totalPage'],
    );
  }
}

class ApiInterceptors extends Interceptor {
  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    String authorization = global.appConfig.getString("token") ?? '';
    if (authorization.isNotEmpty) {
      options.headers['Authorization'] = "Bearer $authorization";
    }

    super.onRequest(options, handler);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    if (err.response?.statusCode == 401) {
      serviceLocator<LoginBloc>().add(const TokenInvalid());
    }

    super.onError(err, handler);
  }
}
