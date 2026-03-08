import 'dart:async';
import 'package:dio/dio.dart';
import 'package:bclms/app/core/constants/api_constants.dart';
import 'package:bclms/app/core/utils/app_logger.dart';
import 'package:bclms/app/services/storage_service.dart';

class ApiService {
  late final Dio dio;
  final StorageService _storageService;
  late String _baseUrl;

  // ใช้ Completer เพื่อ queue request ที่รอ refresh พร้อมกัน
  Completer<String>? _refreshCompleter;

  ApiService(this._storageService);

  String get baseUrl => _baseUrl;

  Future<ApiService> init() async {
    // โหลด URL จาก storage หรือใช้ default
    final savedUrl = await _storageService.getBackendUrl();
    _baseUrl = savedUrl ?? ApiConstants.defaultBaseUrl;

    dio = Dio(BaseOptions(
      baseUrl: _baseUrl,
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 15),
      headers: {'Content-Type': 'application/json'},
    ));

    dio.interceptors.add(InterceptorsWrapper(
      onRequest: _onRequest,
      onError: _onError,
    ));

    AppLogger.info('ApiService เริ่มต้นสำเร็จ (baseUrl: $_baseUrl)');
    return this;
  }

  /// เปลี่ยน base URL ขณะ runtime
  void updateBaseUrl(String url) {
    _baseUrl = url;
    dio.options.baseUrl = url;
    AppLogger.info('ApiService เปลี่ยน baseUrl → $url');
  }

  Dio _createPlainDio() {
    return Dio(BaseOptions(
      baseUrl: _baseUrl,
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 15),
      headers: {'Content-Type': 'application/json'},
    ));
  }

  Future<void> _onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final token = await _storageService.getToken();
    if (token != null) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    AppLogger.info('API Request: ${options.method} ${options.path}');
    handler.next(options);
  }

  Future<void> _onError(
    DioException error,
    ErrorInterceptorHandler handler,
  ) async {
    if (error.response?.statusCode != 401) {
      return handler.next(error);
    }

    // ไม่ refresh สำหรับ login และ refresh endpoint
    final path = error.requestOptions.path;
    if (path == ApiConstants.login || path == ApiConstants.refresh) {
      return handler.next(error);
    }

    // ถ้ามี request อื่นกำลัง refresh อยู่ → รอแล้ว retry ด้วย token ใหม่
    if (_refreshCompleter != null) {
      AppLogger.info('รอ refresh token ที่กำลังดำเนินการ...');
      try {
        final newToken = await _refreshCompleter!.future;
        return handler.resolve(await _retryRequest(error.requestOptions, newToken));
      } catch (_) {
        return handler.next(error);
      }
    }

    // เริ่ม refresh token
    AppLogger.warning('ได้รับ 401 — กำลัง refresh token...');
    _refreshCompleter = Completer<String>();

    try {
      final refreshToken = await _storageService.getRefreshToken();
      if (refreshToken == null) {
        throw Exception('ไม่มี refresh token');
      }

      final plainDio = _createPlainDio();
      final response = await plainDio.post(
        ApiConstants.refresh,
        data: {'token': refreshToken},
      );

      if (response.data['success'] != true) {
        throw Exception('Refresh token ล้มเหลว');
      }

      final newToken = response.data['token'] as String;
      final newRefresh = response.data['refresh'] as String;
      await _storageService.saveTokens(token: newToken, refresh: newRefresh);
      AppLogger.info('Refresh token สำเร็จ');

      // แจ้ง request อื่นที่รออยู่ว่า refresh เสร็จแล้ว
      _refreshCompleter!.complete(newToken);
      _refreshCompleter = null;

      // Retry request เดิมด้วย token ใหม่
      try {
        final retryResponse = await _retryRequest(error.requestOptions, newToken);
        return handler.resolve(retryResponse);
      } catch (retryError) {
        // Retry ล้มเหลว แต่ refresh สำเร็จ → ไม่ต้อง logout
        AppLogger.warning('Retry request ล้มเหลว (refresh สำเร็จแล้ว): ${error.requestOptions.path}');
        return handler.next(error);
      }
    } catch (e) {
      // Refresh ล้มเหลว → แจ้ง request อื่นที่รออยู่
      _refreshCompleter?.completeError(e);
      _refreshCompleter = null;

      AppLogger.error('Refresh token ล้มเหลว — กลับหน้า login', error: e);
      await _storageService.clearAll();
      return handler.next(error);
    }
  }

  /// Retry request ด้วย Dio แยก (ไม่มี interceptor ป้องกัน loop)
  Future<Response> _retryRequest(RequestOptions options, String token) {
    final plainDio = _createPlainDio();
    options.headers['Authorization'] = 'Bearer $token';
    AppLogger.info('Retry: ${options.method} ${options.path}');
    return plainDio.fetch(options);
  }
}
