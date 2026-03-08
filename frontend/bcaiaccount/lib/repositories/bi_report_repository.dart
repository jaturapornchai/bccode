import 'package:flutter/foundation.dart';
import 'package:dio/dio.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/model/bi_report/bi_report_models.dart';
import 'package:smlaicloud/model/bi_report/export_models.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/global.dart' as global;
// Conditional imports for web download
import 'bi_report_download_stub.dart'
    if (dart.library.html) 'bi_report_download_web.dart';

class BiReportRepository {
  // Step 1: Submit report generation job
  Future<BiReportJobResponse> submitReport({
    required BiReportType reportType,
    required Map<String, dynamic> conditions,
    required String token,
  }) async {
    Dio client = Client().initBiReport();

    try {
      // Clean conditions by removing UI-only fields for specific report types
      Map<String, dynamic> cleanedConditions = Map.from(conditions);

      // Remove UI-only fields for specific report types
      if (reportType == BiReportType.sale ||
          reportType == BiReportType.saleDaily ||
          reportType == BiReportType.grossProfitByDocument ||
          reportType == BiReportType.purchasepartial) {
        cleanedConditions.remove('showdetail');
        cleanedConditions.remove('iscancel');
        cleanedConditions.remove('ispos');
      }

      // Gross Profit By Document และ Gross Profit By Product ส่ง conditions โดยตรง
      Map<String, dynamic> requestBody;

      if (reportType == BiReportType.grossProfitByDocument) {
        // ส่งเฉพาะ 5 fields: fromdate, todate, branchcode, creditorcode, iscancel
        requestBody = {
          'fromdate': cleanedConditions['fromdate'],
          'todate': cleanedConditions['todate'],
          'branchcode': cleanedConditions['branchcode'],
          'creditorcode': cleanedConditions['creditorcode'],
        };
      } else if (reportType == BiReportType.grossProfitByProduct) {
        // ส่งเฉพาะ 4 fields: fromdate, todate, barcode, branchcode
        requestBody = {
          'fromdate': cleanedConditions['fromdate'],
          'todate': cleanedConditions['todate'],
          'barcode': cleanedConditions['barcode'] ?? '',
          'branchcode': cleanedConditions['branchcode'] ?? '',
        };
      } else if (reportType == BiReportType.vatSale) {
        // ส่ง conditions พร้อม companyName และ printedBy จาก storage
        final shopName = global.getShopName(); // ดึงจาก storage
        final printedBy =
            global.appConfig.getString("user") ?? ''; // ดึงจาก storage

        requestBody = {
          'conditions': {
            'fromdate': cleanedConditions['fromdate'],
            'todate': cleanedConditions['todate'],
            'branchcode': cleanedConditions['branchcode'] ?? '',
            'companyName': shopName,
            'printedBy': printedBy,
          },
        };
      } else if (reportType == BiReportType.vatBuy) {
        // ส่ง conditions พร้อม companyName และ printedBy จาก storage (เหมือน vatSale)
        final shopName = global.getShopName(); // ดึงจาก storage
        final printedBy =
            global.appConfig.getString("user") ?? ''; // ดึงจาก storage

        requestBody = {
          'conditions': {
            'fromdate': cleanedConditions['fromdate'],
            'todate': cleanedConditions['todate'],
            'branchcode': cleanedConditions['branchcode'] ?? '',
            'companyName': shopName,
            'printedBy': printedBy,
          },
        };
      } else {
        // รายงานอื่นๆ ส่งแบบ {'conditions': {...}}
        requestBody = {'conditions': cleanedConditions};
      }

      final fullEndpoint =
          '${client.options.baseUrl.trim()}${reportType.endpoint}';
      AppLogger.debug('🌐 Submitting report to ${reportType.endpoint}');
      AppLogger.debug('🔗 Resolved endpoint: $fullEndpoint');
      AppLogger.debug('📦 Request body: $requestBody');

      final response = await client.post(
        reportType.endpoint,
        data: requestBody,
        options: Options(
          headers: {
            'Authorization': 'Bearer $token',
            'Content-Type': 'application/json',
          },
        ),
      );

      return BiReportJobResponse.fromJson(
        response.data as Map<String, dynamic>,
      );
    } on DioException catch (ex) {
      if (kDebugMode) {
        AppLogger.error(
          '❌ DioException submitting report: ${ex.response?.data}',
        );
      }
      final errorData = ex.response?.data as Map<String, dynamic>? ?? {};
      throw BiReportException(
        code: ex.response?.statusCode ?? 500,
        message: errorData['message'] ?? global.language('error_in_request'),
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error submitting report: $e');
      }
      throw BiReportException(
        code: 500,
        message: 'เกิดข้อผิดพลาดไม่คาดคิด: ${e.toString()}',
      );
    }
  }

  // Step 2: Check report status
  Future<BiReportStatusResponse> getReportStatus({
    required BiReportType reportType,
    required String jobId,
    required String token,
  }) async {
    Dio client = Client().initBiReport();

    try {
      final response = await client.get(
        '${reportType.endpoint}/$jobId/status',
        options: Options(headers: {'Authorization': 'Bearer $token'}),
      );

      return BiReportStatusResponse.fromJson(
        response.data as Map<String, dynamic>,
      );
    } on DioException catch (ex) {
      if (kDebugMode) {
        AppLogger.error(
          '❌ DioException checking report status: ${ex.response?.data}',
        );
      }
      final errorData = ex.response?.data as Map<String, dynamic>? ?? {};
      throw BiReportException(
        code: ex.response?.statusCode ?? 500,
        message: errorData['message'] ?? 'เกิดข้อผิดพลาดในการตรวจสอบสถานะ',
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error checking report status: $e');
      }
      throw BiReportException(
        code: 500,
        message: 'เกิดข้อผิดพลาดไม่คาดคิด: ${e.toString()}',
      );
    }
  }

  // Step 3: Get report detail data
  Future<BiReportDetailResponse<T>> getReportDetail<T>({
    required BiReportType reportType,
    required String jobId,
    required String token,
    required T Function(Object? json) fromJsonT,
    int page = 1,
    int size = 20,
  }) async {
    Dio client = Client().initBiReport();

    try {
      final response = await client.get(
        '${reportType.endpoint}/$jobId/detail',
        queryParameters: {'page': page, 'size': size},
        options: Options(headers: {'Authorization': 'Bearer $token'}),
      );

      try {
        final result = BiReportDetailResponse.fromJson(
          response.data as Map<String, dynamic>,
          fromJsonT,
        );
        return result;
      } catch (parseError) {
        if (kDebugMode) {
          AppLogger.error('❌ JSON parsing error: $parseError');
          AppLogger.debug(
            '🔍 Raw response data type: ${response.data.runtimeType}',
          );
        }
        if (kDebugMode) {
          AppLogger.debug('🔍 Response data: ${response.data}');
        }
        throw BiReportException(
          code: 422,
          message: 'เกิดข้อผิดพลาดในการแปลงข้อมูล: ${parseError.toString()}',
        );
      }
    } on DioException catch (ex) {
      if (kDebugMode) {
        AppLogger.error(
          '❌ DioException getting report detail: ${ex.response?.data}',
        );
      }
      final errorData = ex.response?.data as Map<String, dynamic>? ?? {};
      throw BiReportException(
        code: ex.response?.statusCode ?? 500,
        message: errorData['message'] ?? global.language('error_fetching_report'),
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error getting report detail: $e');
      }
      throw BiReportException(
        code: 500,
        message: 'เกิดข้อผิดพลาดไม่คาดคิด: ${e.toString()}',
      );
    }
  }

  // Get Sale Report Summary
  Future<dynamic> getReportSummary({
    required BiReportType reportType,
    required String jobId,
    required String token,
  }) async {
    Dio client = Client().initBiReport();

    try {
      final response = await client.get(
        '${reportType.endpoint}/$jobId/summary',
        options: Options(headers: {'Authorization': 'Bearer $token'}),
      );

      // Return raw response data
      return response.data;
    } on DioException catch (ex) {
      if (kDebugMode) {
        AppLogger.error(
          '❌ DioException getting report summary: ${ex.response?.data}',
        );
      }
      final errorData = ex.response?.data as Map<String, dynamic>? ?? {};
      throw BiReportException(
        code: ex.response?.statusCode ?? 500,
        message:
            errorData['message'] ?? global.language('error_fetching_report_summary'),
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error getting report summary: $e');
      }
      throw BiReportException(
        code: 500,
        message: 'เกิดข้อผิดพลาดไม่คาดคิด: ${e.toString()}',
      );
    }
  }

  // Get detail by docno
  Future<dynamic> getDetaildocByDocNo({
    required BiReportType reportType,
    required String jobId,
    required String token,
    required String docNo,
  }) async {
    Dio client = Client().initBiReport();

    try {
      final response = await client.get(
        '${reportType.endpoint}/$jobId/detaildoc/$docNo',
        options: Options(headers: {'Authorization': 'Bearer $token'}),
      );

      // Return raw response data
      return response.data;
    } on DioException catch (ex) {
      if (kDebugMode) {
        AppLogger.debug(
          '❌ DioException getting report detail by docNo: ${ex.response?.data}',
        );
      }
      final errorData = ex.response?.data as Map<String, dynamic>? ?? {};
      throw BiReportException(
        code: ex.response?.statusCode ?? 500,
        message: errorData['message'] ?? global.language('error_fetching_report'),
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error getting report detail by docNo: $e');
      }
      throw BiReportException(
        code: 500,
        message: 'เกิดข้อผิดพลาดไม่คาดคิด: ${e.toString()}',
      );
    }
  }

  // Get detail by docDate
  Future<dynamic> getDetaildocByDocDate({
    required BiReportType reportType,
    required String jobId,
    required String token,
    required String docDate,
  }) async {
    Dio client = Client().initBiReport();

    try {
      final response = await client.get(
        '${reportType.endpoint}/$jobId/detaildate/$docDate',
        options: Options(headers: {'Authorization': 'Bearer $token'}),
      );

      // Return raw response data
      return response.data;
    } on DioException catch (ex) {
      if (kDebugMode) {
        AppLogger.debug(
          '❌ DioException getting report detail by docDate: ${ex.response?.data}',
        );
      }
      final errorData = ex.response?.data as Map<String, dynamic>? ?? {};
      throw BiReportException(
        code: ex.response?.statusCode ?? 500,
        message: errorData['message'] ?? global.language('error_fetching_report'),
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error getting report detail by docDate: $e');
      }
      throw BiReportException(
        code: 500,
        message: 'เกิดข้อผิดพลาดไม่คาดคิด: ${e.toString()}',
      );
    }
  }

  void dispose() {
    // Dio clients are automatically managed, no need to close manually
    // Each method creates its own client instance
  }

  // ==================== EXPORT METHODS ====================

  /// Step 1: Submit export job
  Future<BiReportExportResponse> submitExportJob({
    required BiReportType reportType,
    required String jobId,
    required ExportFormat format,
    required String token,
  }) async {
    Dio client = Client().initBiReport();

    try {
      final requestBody = BiReportExportRequest(format: format).toJson();

      // Build endpoint based on report type
      final endpoint = '${reportType.endpoint}/export/$jobId';

      if (kDebugMode) {
        AppLogger.debug('🚀 Submitting export job:');
        AppLogger.debug('   - Report Type: $reportType');
        AppLogger.debug('   - Job ID: $jobId');
        AppLogger.debug('   - Format: ${format.value}');
        AppLogger.debug('   - Endpoint: $endpoint');
      }

      final response = await client.post(
        endpoint,
        data: requestBody,
        options: Options(
          headers: {
            'Authorization': 'Bearer $token',
            'Content-Type': 'application/json',
          },
        ),
      );

      if (kDebugMode) {
        AppLogger.debug('✅ Export job response: ${response.data}');
      }

      return BiReportExportResponse.fromJson(
        response.data as Map<String, dynamic>,
      );
    } on DioException catch (ex) {
      if (kDebugMode) {
        AppLogger.error(
          '❌ DioException submitting export job: ${ex.response?.data}',
        );
      }
      final errorData = ex.response?.data as Map<String, dynamic>? ?? {};
      throw BiReportException(
        code: ex.response?.statusCode ?? 500,
        message: errorData['message'] ?? 'เกิดข้อผิดพลาดในการส่งคำขอ export',
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error submitting export job: $e');
      }
      throw BiReportException(
        code: 500,
        message: 'เกิดข้อผิดพลาดไม่คาดคิดในการ export: ${e.toString()}',
      );
    }
  }

  /// Step 2: Check export job status
  Future<BiReportExportStatusResponse> getExportStatus({
    required BiReportType reportType,
    required String exportJobId,
    required String token,
  }) async {
    Dio client = Client().initBiReport();

    try {
      // Build endpoint based on report type
      final endpoint = '${reportType.endpoint}/export/$exportJobId/status';

      if (kDebugMode) {
        AppLogger.debug('🔍 Checking export status:');
        AppLogger.debug('   - Report Type: $reportType');
        AppLogger.debug('   - Export Job ID: $exportJobId');
        AppLogger.debug('   - Endpoint: $endpoint');
      }

      final response = await client.get(
        endpoint,
        options: Options(headers: {'Authorization': 'Bearer $token'}),
      );

      if (kDebugMode) {
        AppLogger.debug('📊 Export status response: ${response.data}');
      }

      return BiReportExportStatusResponse.fromJson(
        response.data as Map<String, dynamic>,
      );
    } on DioException catch (ex) {
      if (kDebugMode) {
        AppLogger.error(
          '❌ DioException getting export status: ${ex.response?.data}',
        );
      }
      final errorData = ex.response?.data as Map<String, dynamic>? ?? {};
      throw BiReportException(
        code: ex.response?.statusCode ?? 500,
        message:
            errorData['message'] ?? 'เกิดข้อผิดพลาดในการตรวจสอบสถานะ export',
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error getting export status: $e');
      }
      throw BiReportException(
        code: 500,
        message: 'เกิดข้อผิดพลาดไม่คาดคิดในการตรวจสอบสถานะ: ${e.toString()}',
      );
    }
  }

  /// Step 3: Get download URL
  String getExportDownloadUrl({
    required BiReportType reportType,
    required String exportJobId,
    String? token,
  }) {
    // Build download endpoint based on report type
    final baseUrl = Client().initBiReport().options.baseUrl;
    final endpoint = '${reportType.endpoint}/export/$exportJobId/download';

    String fullUrl = '$baseUrl$endpoint';

    if (kDebugMode) {
      AppLogger.debug('📥 Export download URL:');
      AppLogger.debug('   - Report Type: $reportType');
      AppLogger.debug('   - Export Job ID: $exportJobId');
      AppLogger.debug('   - URL: $fullUrl');
    }

    return fullUrl;
  }

  /// Download export file directly (for web platforms)
  Future<void> downloadExportFileForWeb({
    required BiReportType reportType,
    required String exportJobId,
    required String token,
    required String fileName,
  }) async {
    Dio client = Client().initBiReport();

    try {
      // Build endpoint based on report type
      final endpoint = '${reportType.endpoint}/export/$exportJobId/download';

      if (kDebugMode) {
        AppLogger.debug('📥 Downloading export file for web:');
        AppLogger.debug('   - Report Type: $reportType');
        AppLogger.debug('   - Export Job ID: $exportJobId');
        AppLogger.debug('   - Endpoint: $endpoint');
        AppLogger.debug('   - File Name: $fileName');
      }

      final response = await client.get(
        endpoint,
        options: Options(
          headers: {
            'Authorization': 'Bearer $token',
            'Accept':
                'application/pdf, application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
          },
          responseType: ResponseType.bytes,
        ),
      );

      if (kDebugMode) {
        AppLogger.info('✅ Export file downloaded successfully for web');
        AppLogger.debug(
          '   - Content Length: ${response.data?.length ?? 0} bytes',
        );
      }

      // Create blob and trigger download
      if (kIsWeb) {
        // Import dart:html only for web
        final bytes = response.data as List<int>;
        _triggerWebDownload(bytes, fileName);
      }
    } on DioException catch (ex) {
      if (kDebugMode) {
        AppLogger.debug(
          '❌ DioException downloading export file for web: ${ex.response?.data}',
        );
      }
      final errorData = ex.response?.data as Map<String, dynamic>? ?? {};
      throw BiReportException(
        code: ex.response?.statusCode ?? 500,
        message: errorData['message'] ?? global.language('error_downloading_file'),
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error downloading export file for web: $e');
      }
      throw BiReportException(
        code: 500,
        message: 'เกิดข้อผิดพลาดไม่คาดคิดในการดาวน์โหลด: ${e.toString()}',
      );
    }
  }

  /// Trigger download for web platform
  void _triggerWebDownload(List<int> bytes, String fileName) {
    if (kIsWeb) {
      // This will work only on web platform
      try {
        // Use platform-specific implementation
        triggerWebDownload(bytes, fileName);
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('❌ Error triggering web download: $e');
        }
      }
    } else {
      // For non-web platforms, this function shouldn't be called
      if (kDebugMode) {
        AppLogger.debug('⚠️ _triggerWebDownload called on non-web platform');
      }
    }
  }
}

// Custom Exception Class
class BiReportException implements Exception {
  final int code;
  final String message;

  const BiReportException({required this.code, required this.message});

  @override
  String toString() => 'BiReportException($code): $message';
}
