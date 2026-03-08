import 'package:dio/dio.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/backend_url_manager.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class Client {
  Dio init() {
    Dio dio = Dio();
    dio.interceptors.add(ApiInterceptors());

    // mainapi URL มาจาก BackendUrlManager เสมอ (ดึง origin จาก goapi URL)
    // ตัวอย่าง: https://dev-api.bcaicloud.com/goapi → https://dev-api.bcaicloud.com
    // ตัวอย่าง: http://localhost:9090/goapi → http://localhost:9090
    final mainApiUrl = BackendUrlManager.getMainApiUrl();
    if (mainApiUrl == null) {
      throw Exception('Backend URL ยังไม่ได้ตั้งค่า กรุณาตั้งค่าในหน้า Login');
    }
    String endPointService = mainApiUrl;

    endPointService += endPointService[endPointService.length - 1] == "/"
        ? ""
        : "/";

    AppLogger.debug('[api/Client] baseUrl: $endPointService');

    dio.options.baseUrl = endPointService;
    dio.options.connectTimeout = const Duration(seconds: 120);
    dio.options.receiveTimeout = const Duration(seconds: 30);

    return dio;
  }
}

class ApiResponse<T> {
  late final bool success;
  late final bool error;
  // ignore: unnecessary_question_mark
  late final dynamic? data;
  late final message;
  late final code;
  final Pages? page;

  ApiResponse({
    required this.success,
    required this.data,
    this.error = true,
    this.message = "",
    this.code = 00,
    this.page,
  });

  factory ApiResponse.fromMap(Map<String, dynamic> map) {
    return ApiResponse(
      success: map['success'] ?? false,
      error: map['error'] ?? true,
      data: map['data'],
      page: map['pagination'] == null
          ? Pages.empty
          : Pages.fromMap(map['pagination']),
    );
  }
}

class Pages {
  final int perPage;
  final int page;
  final int total;
  final int totalPage;

  const Pages({
    required this.perPage,
    required this.page,
    required this.total,
    required this.totalPage,
  });

  static const empty = Pages(perPage: 0, page: 0, total: 0, totalPage: 0);

  bool get isEmpty => this == Pages.empty;

  bool get isNotEmpty => this == Pages.empty;

  factory Pages.fromMap(Map<String, dynamic> map) {
    return Pages(
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

    // ส่งภาษาปัจจุบันของ user ไปกับทุก request — backend ใช้เลือกภาษา response message
    options.headers['Accept-Language'] = global.systemLanguage;

    super.onRequest(options, handler);
  }
}
