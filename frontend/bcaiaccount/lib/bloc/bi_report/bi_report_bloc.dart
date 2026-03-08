import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'dart:async';
import 'package:smlaicloud/model/bi_report/bi_report_models.dart';
import 'package:smlaicloud/model/bi_report/bi_sale_report_data.dart';
import 'package:smlaicloud/model/bi_report/payment_daily_model.dart';
import 'package:smlaicloud/model/bi_report/sale_daily_report_model.dart';
import 'package:smlaicloud/model/bi_report/sale_daily_report_summary.dart';
import 'package:smlaicloud/model/bi_report/sale_report_summary.dart';
import 'package:smlaicloud/model/bi_report/sale_return_model.dart';
import 'package:smlaicloud/model/bi_report/stock_balance_model.dart';
import 'package:smlaicloud/model/bi_report/stock_movment_model.dart';
import 'package:smlaicloud/model/bi_report/stock_movment_summary_model.dart';
import 'package:smlaicloud/model/bi_report/purchase_partial_model.dart';
import 'package:smlaicloud/model/bi_report/purchase_partial_summary_model.dart';
import 'package:smlaicloud/model/bi_report/gross_profit_by_document_model.dart';
import 'package:smlaicloud/model/bi_report/gross_profit_by_document_summary_model.dart';
import 'package:smlaicloud/model/bi_report/gross_profit_by_product_model.dart';
import 'package:smlaicloud/model/bi_report/gross_profit_by_product_summary_model.dart';
import 'package:smlaicloud/model/bi_report/vat_sale_model.dart';
import 'package:smlaicloud/model/bi_report/vat_sale_summary_model.dart';
import 'package:smlaicloud/model/bi_report/vat_buy_model.dart';
import 'package:smlaicloud/model/bi_report/vat_buy_summary_model.dart';
import 'package:smlaicloud/model/bi_report/export_models.dart';
import 'package:smlaicloud/repositories/bi_report_repository.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import '../../global.dart' as global;

part 'bi_report_event.dart';
part 'bi_report_state.dart';

class BiReportBloc extends Bloc<BiReportEvent, BiReportState> {
  final BiReportRepository _repository;
  bool _isCancelled = false; // Add cancellation flag
  Completer<void>? _currentPollingCompleter; // Add polling completer
  String _currentJobId = ''; // Track current job ID
  String _currentJobState =
      ''; // Track current job state (waiting, progress, completed, failed)

  // Report type to data model mapping for type safety
  static const _reportDataTypeHandlers = {
    BiReportType.sale: SaleReportData,
    BiReportType.saleDaily: SaleDailyReportData,
    BiReportType.stockMovement: StockMovementModel,
    BiReportType.paymentDaily: PaymentDailyModel,
    BiReportType.saleReturn: SaleReturnModel,
    BiReportType.stockBalance: StockBalanceModel,
    BiReportType.purchasepartial: PurchasePartialModel,
    BiReportType.grossProfitByDocument: GrossProfitByDocumentModel,
    BiReportType.grossProfitByProduct: GrossProfitByProductModel,
    BiReportType.vatSale: VatSaleModel,
    BiReportType.vatBuy: VatBuyModel,
  };

  BiReportBloc({required BiReportRepository biReportRepository})
    : _repository = biReportRepository,
      super(BiReportInitial()) {
    on<GenerateBiReportRequested>(_onGenerateBiReportRequested);
    on<GetBiReportDetailRequested>(_onGetBiReportDetailRequested);
    on<GetBiReportDetailByDocNoRequested>(_onGetBiReportDetailByDocNoRequested);
    on<GetBiReportDetailByDocDateRequested>(
      _onGetBiReportDetailByDocDateRequested,
    );
    on<GetBiReportSummaryRequested>(_onGetBiReportSummaryRequested);
    on<CancelBiReportRequested>(_onCancelBiReportRequested);
    on<ResetBiReportState>(_onResetBiReportState);

    // Export event handlers
    on<SubmitBiReportExportRequested>(_onSubmitBiReportExportRequested);
    on<CheckBiReportExportStatusRequested>(
      _onCheckBiReportExportStatusRequested,
    );
    on<DownloadBiReportExportRequested>(_onDownloadBiReportExportRequested);
    on<DownloadBiReportExportFile>(_onDownloadBiReportExportFile);
    on<CancelBiReportExportRequested>(_onCancelBiReportExportRequested);
    on<ResetBiReportExportState>(_onResetBiReportExportState);
  }

  // Complete report generation process (Submit → Poll → Get Details)
  Future<void> _onGenerateBiReportRequested(
    GenerateBiReportRequested event,
    Emitter<BiReportState> emit,
  ) async {
    try {
      _isCancelled = false; // Reset cancellation flag
      _currentPollingCompleter?.complete(); // Cancel any existing polling
      _currentPollingCompleter = Completer<void>();
      _currentJobId = ''; // Reset job ID
      _currentJobState = ''; // Reset job state

      emit(BiReportGenerating(reportType: event.reportType));

      // Validate conditions type
      if (event.conditions is! ReportConditionsModel) {
        emit(
          BiReportGenerateFailure(message: 'รูปแบบเงื่อนไขรายงานไม่ถูกต้อง'),
        );
        return;
      }

      // Check if report type is supported
      if (!_reportDataTypeHandlers.containsKey(event.reportType)) {
        emit(
          BiReportGenerateFailure(
            message:
                'รายงานประเภท ${event.reportType.displayName} ยังไม่รองรับ',
          ),
        );
        return;
      }

      // Use generic report generation
      await _generateGenericReport(
        event.reportType,
        event.conditions as ReportConditionsModel,
        event.token,
        emit,
        timeout: event.timeout,
        pollInterval: event.pollInterval,
        page: event.page,
        size: event.size,
      );
    } on BiReportException catch (e) {
      emit(BiReportGenerateFailure(message: e.message, errorCode: e.code));
    } catch (e) {
      emit(
        BiReportGenerateFailure(
          message: 'เกิดข้อผิดพลาดไม่คาดคิด: ${e.toString()}',
        ),
      );
    } finally {
      _currentPollingCompleter = null;
    }
  }

  // Generic report generation method
  Future<void> _generateGenericReport(
    BiReportType reportType,
    ReportConditionsModel conditions,
    String token,
    Emitter<BiReportState> emit, {
    Duration? timeout,
    Duration? pollInterval,
    int? page,
    int? size,
  }) async {
    // Step 1: Submit report job
    final jobResponse = await _repository.submitReport(
      reportType: reportType,
      conditions: conditions.toJson(),
      token: token,
    );

    final jobId = jobResponse.jobId;
    _currentJobId = jobId; // Store current job ID

    emit(
      BiReportGenerateProgress(
        reportType: reportType,
        jobId: jobId,
        progress: 0,
        statusMessage: 'เริ่มสร้าง${reportType.displayName}...',
      ),
    );

    // Step 2: Poll for completion with progress updates
    final startTime = DateTime.now();
    final timeoutDuration = timeout ?? const Duration(minutes: 10);
    final pollIntervalDuration = pollInterval ?? const Duration(seconds: 2);

    if (kDebugMode) {
      AppLogger.debug('⏰ Starting polling for job: $jobId');
      AppLogger.debug(
        '   - Poll interval: ${pollIntervalDuration.inSeconds} seconds',
      );
      AppLogger.debug('   - Timeout: ${timeoutDuration.inMinutes} minutes');
      AppLogger.debug(
        '   - Start time: ${startTime.toString().substring(11, 19)}',
      );
    }

    try {
      while (DateTime.now().difference(startTime) < timeoutDuration &&
          !_isCancelled &&
          !(_currentPollingCompleter?.isCompleted ?? false)) {
        // Use Future.any to listen for both delay and cancellation
        await Future.any([
          Future.delayed(pollIntervalDuration),
          _currentPollingCompleter?.future ?? Future<void>.value(),
        ]);

        // Check if cancelled after delay
        if (_isCancelled || (_currentPollingCompleter?.isCompleted ?? false)) {
          if (kDebugMode) {
            AppLogger.debug('🛑 Polling cancelled for job: $jobId');
          }
          emit(
            BiReportGenerateFailure(
              message: 'การสร้าง${reportType.displayName}ถูกยกเลิก',
            ),
          );
          return;
        }

        if (kDebugMode) {
          AppLogger.debug(
            '🔍 Checking status for job: $jobId (${DateTime.now().toString().substring(11, 19)})',
          );
        }

        final statusResponse = await _repository.getReportStatus(
          reportType: reportType,
          jobId: jobId,
          token: token,
        );

        if (kDebugMode) {
          AppLogger.debug(
            '📊 Status response: ${statusResponse.data.state} (${statusResponse.data.progress}%)',
          );
        }

        // Update current job state
        _currentJobState = statusResponse.data.state;

        // Emit progress update
        emit(
          BiReportGenerateProgress(
            reportType: reportType,
            jobId: jobId,
            progress: statusResponse.data.progress,
            statusMessage: _getStatusMessage(
              statusResponse.data.state,
              statusResponse.data.progress,
            ),
          ),
        );

        if (statusResponse.data.state == 'completed') {
          // Step 3: Get report details using generic method
          await _getReportDetailByType(
            reportType: reportType,
            jobId: jobId,
            token: token,
            emit: emit,
            page: page ?? 1,
            size: size ?? 20,
          );
          return;
        } else if (statusResponse.data.state == 'failed') {
          emit(
            BiReportGenerateFailure(
              message: _formatErrorMessage(statusResponse.data.failedReason),
            ),
          );
          return;
        }
      }

      // Check if cancelled after timeout
      if (_isCancelled || (_currentPollingCompleter?.isCompleted ?? false)) {
        emit(
          BiReportGenerateFailure(
            message: 'การสร้าง${reportType.displayName}ถูกยกเลิก',
          ),
        );
      } else {
        // Timeout reached
        emit(
          BiReportGenerateFailure(
            message: 'การสร้าง${reportType.displayName}ใช้เวลานานเกินไป',
          ),
        );
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error in polling: $e');
      }
      emit(
        BiReportGenerateFailure(
          message: 'เกิดข้อผิดพลาดในการตรวจสอบสถานะรายงาน',
        ),
      );
    }
  }

  // Generic method to get report details based on type
  Future<void> _getReportDetailByType({
    required BiReportType reportType,
    required String jobId,
    required String token,
    required Emitter<BiReportState> emit,
    required int page,
    required int size,
  }) async {
    switch (reportType) {
      case BiReportType.sale:
        final result = await _repository.getReportDetail<SaleReportData>(
          reportType: reportType,
          jobId: jobId,
          token: token,
          fromJsonT: (json) =>
              SaleReportData.fromJson(json as Map<String, dynamic>),
          page: page,
          size: size,
        );

        emit(
          BiReportGenerateSuccess(
            reportType: reportType,
            jobId: jobId,
            data: result.data,
            meta: result.meta,
          ),
        );
        break;

      case BiReportType.saleDaily:
        final result = await _repository.getReportDetail<SaleDailyReportData>(
          reportType: reportType,
          jobId: jobId,
          token: token,
          fromJsonT: (json) =>
              SaleDailyReportData.fromJson(json as Map<String, dynamic>),
          page: page,
          size: size,
        );

        emit(
          BiReportGenerateSuccess(
            reportType: reportType,
            jobId: jobId,
            data: result.data,
            meta: result.meta,
          ),
        );
        break;

      case BiReportType.stockMovement:
        final result = await _repository.getReportDetail<StockMovementModel>(
          reportType: reportType,
          jobId: jobId,
          token: token,
          fromJsonT: (json) =>
              StockMovementModel.fromJson(json as Map<String, dynamic>),
          page: page,
          size: size,
        );

        emit(
          BiReportGenerateSuccess(
            reportType: reportType,
            jobId: jobId,
            data: result.data,
            meta: result.meta,
          ),
        );
        break;

      case BiReportType.paymentDaily:
        final result = await _repository.getReportDetail<PaymentDailyModel>(
          reportType: reportType,
          jobId: jobId,
          token: token,
          fromJsonT: (json) =>
              PaymentDailyModel.fromJson(json as Map<String, dynamic>),
          page: page,
          size: size,
        );

        emit(
          BiReportGenerateSuccess(
            reportType: reportType,
            jobId: jobId,
            data: result.data,
            meta: result.meta,
          ),
        );
        break;

      case BiReportType.saleReturn:
        final result = await _repository.getReportDetail<SaleReturnModel>(
          reportType: reportType,
          jobId: jobId,
          token: token,
          fromJsonT: (json) =>
              SaleReturnModel.fromJson(json as Map<String, dynamic>),
          page: page,
          size: size,
        );

        emit(
          BiReportGenerateSuccess(
            reportType: reportType,
            jobId: jobId,
            data: result.data,
            meta: result.meta,
          ),
        );
        break;

      // stockbalance
      case BiReportType.stockBalance:
        final result = await _repository.getReportDetail<StockBalanceModel>(
          reportType: reportType,
          jobId: jobId,
          token: token,
          fromJsonT: (json) =>
              StockBalanceModel.fromJson(json as Map<String, dynamic>),
          page: page,
          size: size,
        );

        emit(
          BiReportGenerateSuccess(
            reportType: reportType,
            jobId: jobId,
            data: result.data,
            meta: result.meta,
          ),
        );
        break;

      // purchasepartial
      case BiReportType.purchasepartial:
        final result = await _repository.getReportDetail<PurchasePartialModel>(
          reportType: reportType,
          jobId: jobId,
          token: token,
          fromJsonT: (json) =>
              PurchasePartialModel.fromJson(json as Map<String, dynamic>),
          page: page,
          size: size,
        );

        emit(
          BiReportGenerateSuccess(
            reportType: reportType,
            jobId: jobId,
            data: result.data,
            meta: result.meta,
          ),
        );
        break;

      // grossProfitByDocument
      case BiReportType.grossProfitByDocument:
        final result = await _repository
            .getReportDetail<GrossProfitByDocumentModel>(
              reportType: reportType,
              jobId: jobId,
              token: token,
              fromJsonT: (json) => GrossProfitByDocumentModel.fromJson(
                json as Map<String, dynamic>,
              ),
              page: page,
              size: size,
            );

        emit(
          BiReportGenerateSuccess(
            reportType: reportType,
            jobId: jobId,
            data: result.data,
            meta: result.meta,
          ),
        );
        break;

      // grossProfitByProduct
      case BiReportType.grossProfitByProduct:
        final result = await _repository
            .getReportDetail<GrossProfitByProductModel>(
              reportType: reportType,
              jobId: jobId,
              token: token,
              fromJsonT: (json) => GrossProfitByProductModel.fromJson(
                json as Map<String, dynamic>,
              ),
              page: page,
              size: size,
            );

        emit(
          BiReportGenerateSuccess(
            reportType: reportType,
            jobId: jobId,
            data: result.data,
            meta: result.meta,
          ),
        );
        break;

      // vatSale
      case BiReportType.vatSale:
        final result = await _repository.getReportDetail<VatSaleModel>(
          reportType: reportType,
          jobId: jobId,
          token: token,
          fromJsonT: (json) =>
              VatSaleModel.fromJson(json as Map<String, dynamic>),
          page: page,
          size: size,
        );

        emit(
          BiReportGenerateSuccess(
            reportType: reportType,
            jobId: jobId,
            data: result.data,
            meta: result.meta,
          ),
        );
        break;

      // vatBuy
      case BiReportType.vatBuy:
        final result = await _repository.getReportDetail<VatBuyModel>(
          reportType: reportType,
          jobId: jobId,
          token: token,
          fromJsonT: (json) =>
              VatBuyModel.fromJson(json as Map<String, dynamic>),
          page: page,
          size: size,
        );

        emit(
          BiReportGenerateSuccess(
            reportType: reportType,
            jobId: jobId,
            data: result.data,
            meta: result.meta,
          ),
        );
        break;
    }
  }

  String _getStatusMessage(String state, int progress) {
    switch (state) {
      case 'queued':
        return 'อยู่ในคิวรอ...';
      case 'processing':
        return 'กำลังประมวลผล... ($progress%)';
      case 'completed':
        return global.language('done');
      case 'failed':
        return global.language('failed');
      default:
        return 'สถานะ: $state ($progress%)';
    }
  }

  String _formatErrorMessage(String? failedReason) {
    if (failedReason == null) return 'การสร้างรายงานล้มเหลวโดยไม่ทราบสาเหตุ';

    // แปลง error messages ที่พบบ่อยเป็นภาษาไทย
    if (failedReason.contains('invalid input syntax for type bigint')) {
      return 'รูปแบบข้อมูลไม่ถูกต้อง กรุณาตรวจสอบเงื่อนไขการค้นหา';
    } else if (failedReason.contains('timeout')) {
      return 'การประมวลผลใช้เวลานานเกินไป กรุณาลองใหม่';
    } else if (failedReason.contains('permission denied')) {
      return 'ไม่มีสิทธิ์เข้าถึงข้อมูล';
    } else if (failedReason.contains('connection')) {
      return 'เกิดปัญหาการเชื่อมต่อ กรุณาลองใหม่';
    }

    return 'การสร้างรายงานล้มเหลว: $failedReason';
  }

  // Step 3: Get report details
  Future<void> _onGetBiReportDetailRequested(
    GetBiReportDetailRequested event,
    Emitter<BiReportState> emit,
  ) async {
    try {
      emit(
        BiReportDetailLoading(reportType: event.reportType, jobId: event.jobId),
      );

      // Use generic method to get report details
      await _getReportDetailByType(
        reportType: event.reportType,
        jobId: event.jobId,
        token: event.token,
        emit: emit,
        page: event.page ?? 1,
        size: event.size ?? 20,
      );
    } on BiReportException catch (e) {
      emit(
        BiReportDetailFailure(
          jobId: event.jobId,
          message: e.message,
          errorCode: e.code,
        ),
      );
    } catch (e) {
      emit(
        BiReportDetailFailure(
          jobId: event.jobId,
          message: 'Unexpected error: ${e.toString()}',
        ),
      );
    }
  }

  // Get report detail by docno
  Future<void> _onGetBiReportDetailByDocNoRequested(
    GetBiReportDetailByDocNoRequested event,
    Emitter<BiReportState> emit,
  ) async {
    try {
      emit(
        BiReportDetailByDocNoLoading(
          reportType: event.reportType,
          jobId: event.jobId,
          docNo: event.docNo,
        ),
      );

      final response = await _repository.getDetaildocByDocNo(
        reportType: event.reportType,
        jobId: event.jobId,
        token: event.token,
        docNo: event.docNo,
      );

      // Parse response based on report type
      List<dynamic> parsedData = [];

      if (response is Map<String, dynamic> && response.containsKey('data')) {
        final responseData = response['data'];

        if (responseData is List) {
          // Parse each item in the list based on report type
          for (final item in responseData) {
            if (item is Map<String, dynamic>) {
              if (event.reportType case BiReportType.sale) {
                parsedData.add(SaleReportData.fromJson(item));
              } else if (event.reportType case BiReportType.saleReturn) {
                parsedData.add(SaleReturnModel.fromJson(item));
              }
            }
          }
        } else if (responseData is Map<String, dynamic>) {
          // Single item response
          if (event.reportType case BiReportType.sale) {
            parsedData.add(SaleReportData.fromJson(responseData));
          } else if (event.reportType case BiReportType.saleReturn) {
            parsedData.add(SaleReturnModel.fromJson(responseData));
          }
        }
      }

      emit(
        BiReportDetailByDocNoSuccess(
          reportType: event.reportType,
          jobId: event.jobId,
          docNo: event.docNo,
          data: parsedData,
        ),
      );
    } on BiReportException catch (e) {
      emit(
        BiReportDetailByDocNoFailure(
          jobId: event.jobId,
          docNo: event.docNo,
          message: e.message,
          errorCode: e.code,
        ),
      );
    } catch (e) {
      emit(
        BiReportDetailByDocNoFailure(
          jobId: event.jobId,
          docNo: event.docNo,
          message: 'เกิดข้อผิดพลาดในการดึงข้อมูลรายละเอียด: ${e.toString()}',
        ),
      );
    }
  }

  // Get report detail by docDate
  Future<void> _onGetBiReportDetailByDocDateRequested(
    GetBiReportDetailByDocDateRequested event,
    Emitter<BiReportState> emit,
  ) async {
    try {
      emit(
        BiReportDetailByDocDateLoading(
          reportType: event.reportType,
          jobId: event.jobId,
          docDate: event.docDate,
        ),
      );

      final response = await _repository.getDetaildocByDocDate(
        reportType: event.reportType,
        jobId: event.jobId,
        token: event.token,
        docDate: event.docDate,
      );

      // Parse response based on report type
      List<dynamic> parsedData = [];

      if (response is Map<String, dynamic> && response.containsKey('data')) {
        final responseData = response['data'];

        if (responseData is List) {
          // Parse each item in the list based on report type
          for (final item in responseData) {
            if (item is Map<String, dynamic>) {
              if (event.reportType case BiReportType.saleDaily) {
                parsedData.add(SaleDailyReportData.fromJson(item));
              } else if (event.reportType case BiReportType.paymentDaily) {
                parsedData.add(PaymentDailyModel.fromJson(item));
              }
            }
          }
        } else if (responseData is Map<String, dynamic>) {
          // Single item response
          if (event.reportType case BiReportType.saleDaily) {
            parsedData.add(SaleDailyReportData.fromJson(responseData));
          } else if (event.reportType case BiReportType.paymentDaily) {
            parsedData.add(PaymentDailyModel.fromJson(responseData));
          }
        }
      }

      emit(
        BiReportDetailByDocDateSuccess(
          reportType: event.reportType,
          jobId: event.jobId,
          docDate: event.docDate,
          data: parsedData,
        ),
      );
    } on BiReportException catch (e) {
      emit(
        BiReportDetailByDocDateFailure(
          jobId: event.jobId,
          docDate: event.docDate,
          message: e.message,
          errorCode: e.code,
        ),
      );
    } catch (e) {
      emit(
        BiReportDetailByDocDateFailure(
          jobId: event.jobId,
          docDate: event.docDate,
          message: 'เกิดข้อผิดพลาดในการดึงข้อมูลรายละเอียด: ${e.toString()}',
        ),
      );
    }
  }

  // Get report summary - แก้ไขให้เป็น generic
  Future<void> _onGetBiReportSummaryRequested(
    GetBiReportSummaryRequested event,
    Emitter<BiReportState> emit,
  ) async {
    try {
      emit(
        BiReportSummaryLoading(
          reportType: event.reportType,
          jobId: event.jobId,
        ),
      );

      final summaryData = await _repository.getReportSummary(
        reportType: event.reportType,
        jobId: event.jobId,
        token: event.token,
      );

      // แยกการจัดการตาม reportType ด้วย generic method
      await _handleReportSummaryByType(
        emit: emit,
        event: event,
        summaryData: summaryData,
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error fetching report summary: $e');
        AppLogger.debug('Summary data type: ${e.runtimeType}');
      }

      String errorMessage = global.language('error_fetching_report_summary');
      int? errorCode;

      if (e is BiReportException) {
        errorMessage = e.message;
        errorCode = e.code;
      } else if (e.toString().contains('type') ||
          e.toString().contains('format')) {
        errorMessage = 'รูปแบบข้อมูลสรุปรายงานไม่ถูกต้อง';
      }

      emit(
        BiReportSummaryFailure(
          jobId: event.jobId,
          message: errorMessage,
          errorCode: errorCode,
        ),
      );
    }
  }

  // Generic method to handle report summary by type
  Future<void> _handleReportSummaryByType({
    required Emitter<BiReportState> emit,
    required GetBiReportSummaryRequested event,
    required dynamic summaryData,
  }) async {
    try {
      if (summaryData is! Map<String, dynamic>) {
        throw Exception(
          'Invalid summary data format: expected Map<String, dynamic>, got ${summaryData.runtimeType}',
        );
      }

      switch (event.reportType) {
        case BiReportType.sale:
          final dataSum = SaleReportSummary.fromJson(summaryData["data"]);
          emit(
            BiReportSummarySuccess(
              reportType: event.reportType,
              jobId: event.jobId,
              summaryData: dataSum,
            ),
          );
          break;

        case BiReportType.saleDaily:
          final dataSum = SaleDailyReportSummary.fromJson(summaryData["data"]);
          emit(
            BiReportSummarySuccess(
              reportType: event.reportType,
              jobId: event.jobId,
              dailySummaryData: dataSum,
            ),
          );
          break;

        case BiReportType.stockMovement:
          final dataSum = StockMovmentSummaryModel.fromJson(
            summaryData["data"],
          );
          emit(
            BiReportSummarySuccess(
              reportType: event.reportType,
              jobId: event.jobId,
              stockMovementSummaryData: dataSum,
            ),
          );
          break;

        case BiReportType.paymentDaily:
          final dataSum = PaymentDailySummaryModel.fromJson(
            summaryData["data"],
          );
          emit(
            BiReportSummarySuccess(
              reportType: event.reportType,
              jobId: event.jobId,
              paymentDailySummaryData: dataSum,
            ),
          );
          break;

        case BiReportType.saleReturn:
          final dataSum = SaleReturnSummaryModel.fromJson(summaryData["data"]);
          emit(
            BiReportSummarySuccess(
              reportType: event.reportType,
              jobId: event.jobId,
              saleReturnSummaryData: dataSum,
            ),
          );
          break;

        case BiReportType.stockBalance:
          final dataSum = StockBalanceSummaryModel.fromJson(
            summaryData["data"],
          );
          emit(
            BiReportSummarySuccess(
              reportType: event.reportType,
              jobId: event.jobId,
              stockBalanceSummaryData: dataSum,
            ),
          );
          break;

        case BiReportType.purchasepartial:
          final dataSum = PurchasePartialSummaryModel.fromJson(
            summaryData["data"],
          );
          emit(
            BiReportSummarySuccess(
              reportType: event.reportType,
              jobId: event.jobId,
              purchasePartialSummaryData: dataSum,
            ),
          );
          break;

        case BiReportType.grossProfitByDocument:
          final dataSum = GrossProfitByDocumentSummary.fromJson(
            summaryData["data"],
          );
          emit(
            BiReportSummarySuccess(
              reportType: event.reportType,
              jobId: event.jobId,
              grossProfitSummaryData: dataSum,
            ),
          );
          break;

        case BiReportType.grossProfitByProduct:
          final dataSum = GrossProfitByProductSummary.fromJson(
            summaryData["data"],
          );
          emit(
            BiReportSummarySuccess(
              reportType: event.reportType,
              jobId: event.jobId,
              grossProfitByProductSummaryData: dataSum,
            ),
          );
          break;

        case BiReportType.vatSale:
          final dataSum = VatSaleSummary.fromJson(summaryData["data"]);
          emit(
            BiReportSummarySuccess(
              reportType: event.reportType,
              jobId: event.jobId,
              vatSaleSummaryData: dataSum,
            ),
          );
          break;

        case BiReportType.vatBuy:
          final dataSum = VatBuySummary.fromJson(summaryData["data"]);
          emit(
            BiReportSummarySuccess(
              reportType: event.reportType,
              jobId: event.jobId,
              vatBuySummaryData: dataSum,
            ),
          );
          break;
      }
    } catch (e) {
      throw BiReportException(
        message:
            'เกิดข้อผิดพลาดในการประมวลผลข้อมูลสรุป${event.reportType.displayName}: ${e.toString()}',
        code: 1000 + event.reportType.index,
      );
    }
  }

  // Cancel ongoing report generation
  Future<void> _onCancelBiReportRequested(
    CancelBiReportRequested event,
    Emitter<BiReportState> emit,
  ) async {
    if (kDebugMode) {
      AppLogger.debug(
        '🛑 Cancelling report generation for job: ${event.jobId}',
      );
      AppLogger.debug('🛑 Current job ID: $_currentJobId');
      AppLogger.debug('🛑 Current job state: $_currentJobState');
    }

    // Only cancel if the job ID matches current job and job is not already completed
    if ((event.jobId == _currentJobId || event.jobId.isEmpty) &&
        _currentJobState != 'completed') {
      if (_currentJobState == 'completed') {
        if (kDebugMode) {
          AppLogger.info('🛑 Report already completed, skipping cancellation');
        }
        return;
      }

      _isCancelled = true;

      // Complete the polling completer to stop polling loop
      if (_currentPollingCompleter != null &&
          !_currentPollingCompleter!.isCompleted) {
        _currentPollingCompleter!.complete();
        if (kDebugMode) {
          AppLogger.info('🛑 Polling completer completed');
        }
      }

      // Emit cancelled state
      emit(BiReportGenerateFailure(message: 'การสร้างรายงานถูกยกเลิก'));
    } else {
      if (kDebugMode) {
        if (event.jobId != _currentJobId && event.jobId.isNotEmpty) {
          AppLogger.debug('🛑 Job ID mismatch, not cancelling');
        } else if (_currentJobState == 'completed') {
          AppLogger.info('🛑 Report already completed, not cancelling');
        }
      }
    }
  }

  // Reset state
  Future<void> _onResetBiReportState(
    ResetBiReportState event,
    Emitter<BiReportState> emit,
  ) async {
    _isCancelled = false; // Reset cancellation flag
    _currentJobId = ''; // Reset job ID

    // Complete any ongoing polling
    if (_currentPollingCompleter != null &&
        !_currentPollingCompleter!.isCompleted) {
      _currentPollingCompleter!.complete();
    }
    _currentPollingCompleter = null;

    emit(BiReportInitial());
  }

  @override
  Future<void> close() {
    // Clean up any ongoing polling
    if (_currentPollingCompleter != null &&
        !_currentPollingCompleter!.isCompleted) {
      _currentPollingCompleter!.complete();
    }

    _repository.dispose();
    return super.close();
  }

  // ==================== EXPORT HANDLERS ====================

  /// Submit export job
  Future<void> _onSubmitBiReportExportRequested(
    SubmitBiReportExportRequested event,
    Emitter<BiReportState> emit,
  ) async {
    try {
      if (kDebugMode) {
        AppLogger.debug('🚀 Submitting export job');
        AppLogger.debug('   - Report Type: ${event.reportType}');
        AppLogger.debug('   - Job ID: ${event.jobId}');
        AppLogger.debug('   - Format: ${event.format.value}');
      }

      emit(
        BiReportExportSubmitting(
          reportType: event.reportType,
          jobId: event.jobId,
          format: event.format,
        ),
      );

      final response = await _repository.submitExportJob(
        reportType: event.reportType,
        jobId: event.jobId,
        format: event.format,
        token: event.token,
      );

      if (kDebugMode) {
        AppLogger.info('✅ Export job submitted successfully');
        AppLogger.debug('   - Export Job ID: ${response.jobId}');
      }

      emit(
        BiReportExportSubmitSuccess(
          reportType: event.reportType,
          originalJobId: event.jobId,
          exportJobId: response.jobId,
          format: event.format,
          message: response.message,
        ),
      );

      // Auto-start polling export status
      add(
        CheckBiReportExportStatusRequested(
          reportType: event.reportType,
          exportJobId: response.jobId,
          token: event.token,
        ),
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error submitting export job: $e');
      }

      final errorMessage = e is BiReportException ? e.message : e.toString();
      final errorCode = e is BiReportException ? e.code : 500;

      emit(
        BiReportExportSubmitFailure(
          reportType: event.reportType,
          jobId: event.jobId,
          format: event.format,
          message: errorMessage,
          errorCode: errorCode,
        ),
      );
    }
  }

  /// Check export status with polling
  Future<void> _onCheckBiReportExportStatusRequested(
    CheckBiReportExportStatusRequested event,
    Emitter<BiReportState> emit,
  ) async {
    try {
      if (kDebugMode) {
        AppLogger.debug('🔍 Starting export status polling');
        AppLogger.debug('   - Report Type: ${event.reportType}');
        AppLogger.debug('   - Export Job ID: ${event.exportJobId}');
      }

      emit(
        BiReportExportStatusChecking(
          reportType: event.reportType,
          exportJobId: event.exportJobId,
        ),
      );

      // Poll export status every 2 seconds with 10 minute timeout
      const pollInterval = Duration(seconds: 2);
      const timeout = Duration(minutes: 10);
      final startTime = DateTime.now();

      while (DateTime.now().difference(startTime) < timeout) {
        final statusResponse = await _repository.getExportStatus(
          reportType: event.reportType,
          exportJobId: event.exportJobId,
          token: event.token,
        );

        final state = exportStateFromString(statusResponse.state);

        if (kDebugMode) {
          AppLogger.debug(
            '📊 Export status: ${statusResponse.state} (${statusResponse.progress}%)',
          );
        }

        // Emit progress update
        emit(
          BiReportExportStatusProgress(
            reportType: event.reportType,
            exportJobId: event.exportJobId,
            state: state,
            progress: statusResponse.progress,
            statusMessage: _getExportStatusMessage(
              state,
              statusResponse.progress,
            ),
          ),
        );

        if (state == ExportState.completed) {
          if (kDebugMode) {
            AppLogger.info('✅ Export completed successfully');
            AppLogger.debug('   - File Name: ${statusResponse.fileName}');
          }

          final downloadUrl = _repository.getExportDownloadUrl(
            reportType: event.reportType,
            exportJobId: event.exportJobId,
            token: event.token,
          );

          emit(
            BiReportExportReady(
              reportType: event.reportType,
              exportJobId: event.exportJobId,
              fileName: statusResponse.fileName ?? 'export_file',
              downloadUrl: downloadUrl,
              format: _getFormatFromFileName(statusResponse.fileName ?? ''),
            ),
          );
          return;
        } else if (state == ExportState.failed) {
          if (kDebugMode) {
            AppLogger.error('❌ Export failed: ${statusResponse.failedReason}');
          }

          emit(
            BiReportExportStatusFailure(
              reportType: event.reportType,
              exportJobId: event.exportJobId,
              message: statusResponse.failedReason ?? 'Export failed',
            ),
          );
          return;
        }

        // Wait before next poll
        await Future.delayed(pollInterval);
      }

      // Timeout
      emit(
        BiReportExportStatusFailure(
          reportType: event.reportType,
          exportJobId: event.exportJobId,
          message: 'Export timeout - การ export ใช้เวลานานเกินกำหนด',
        ),
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error checking export status: $e');
      }

      final errorMessage = e is BiReportException ? e.message : e.toString();
      final errorCode = e is BiReportException ? e.code : 500;

      emit(
        BiReportExportStatusFailure(
          reportType: event.reportType,
          exportJobId: event.exportJobId,
          message: errorMessage,
          errorCode: errorCode,
        ),
      );
    }
  }

  /// Download export file directly (for web)
  Future<void> _onDownloadBiReportExportRequested(
    DownloadBiReportExportRequested event,
    Emitter<BiReportState> emit,
  ) async {
    emit(
      BiReportExportDownloading(
        reportType: event.reportType,
        exportJobId: event.exportJobId,
        fileName: event.fileName,
      ),
    );

    try {
      if (kDebugMode) {
        AppLogger.debug('🌐 Starting direct download for web platform');
      }

      await _repository.downloadExportFileForWeb(
        reportType: event.reportType,
        exportJobId: event.exportJobId,
        token: event.token,
        fileName: event.fileName,
      );

      emit(
        BiReportExportDownloadSuccess(
          reportType: event.reportType,
          exportJobId: event.exportJobId,
          fileName: event.fileName,
          filePath: '', // For web, no local path
        ),
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Direct download failed: $e');
      }

      String errorMessage = global.language('error_downloading_file');
      int? errorCode;

      if (e is BiReportException) {
        errorMessage = e.message;
        errorCode = e.code;
      }

      emit(
        BiReportExportDownloadFailure(
          reportType: event.reportType,
          exportJobId: event.exportJobId,
          fileName: event.fileName,
          message: errorMessage,
          errorCode: errorCode,
        ),
      );
    }
  }

  /// Cancel export process
  Future<void> _onCancelBiReportExportRequested(
    CancelBiReportExportRequested event,
    Emitter<BiReportState> emit,
  ) async {
    if (kDebugMode) {
      AppLogger.debug('🛑 Cancelling export process');
    }

    emit(const BiReportExportCancelled());
  }

  /// Reset export state
  Future<void> _onResetBiReportExportState(
    ResetBiReportExportState event,
    Emitter<BiReportState> emit,
  ) async {
    if (kDebugMode) {
      AppLogger.debug('🔄 Resetting export state');
    }

    emit(BiReportExportInitial());
  }

  // Helper methods for export
  String _getExportStatusMessage(ExportState state, int progress) {
    switch (state) {
      case ExportState.waiting:
        return 'รอการดำเนินการ...';
      case ExportState.progress:
        return 'กำลังสร้างไฟล์... ($progress%)';
      case ExportState.completed:
        return global.language('done');
      case ExportState.failed:
        return global.language('failed');
    }
  }

  ExportFormat _getFormatFromFileName(String fileName) {
    if (fileName.toLowerCase().endsWith('.pdf')) {
      return ExportFormat.pdf;
    } else if (fileName.toLowerCase().endsWith('.xlsx') ||
        fileName.toLowerCase().endsWith('.xls')) {
      return ExportFormat.excel;
    }
    return ExportFormat.pdf; // Default
  }

  /// Download export file directly (for web)
  Future<void> _onDownloadBiReportExportFile(
    DownloadBiReportExportFile event,
    Emitter<BiReportState> emit,
  ) async {
    emit(
      BiReportExportDownloading(
        reportType: event.reportType,
        exportJobId: event.exportJobId,
        fileName: event.fileName,
      ),
    );

    try {
      if (kDebugMode) {
        AppLogger.debug('🌐 Starting direct download for web');
      }

      await _repository.downloadExportFileForWeb(
        reportType: event.reportType,
        exportJobId: event.exportJobId,
        token: event.token,
        fileName: event.fileName,
      );

      emit(
        BiReportExportDownloadSuccess(
          reportType: event.reportType,
          exportJobId: event.exportJobId,
          fileName: event.fileName,
          filePath: '', // For web, no local path
        ),
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Direct download failed: $e');
      }

      String errorMessage = global.language('error_downloading_file');
      int? errorCode;

      if (e is BiReportException) {
        errorMessage = e.message;
        errorCode = e.code;
      }

      emit(
        BiReportExportDownloadFailure(
          reportType: event.reportType,
          exportJobId: event.exportJobId,
          fileName: event.fileName,
          message: errorMessage,
          errorCode: errorCode,
        ),
      );
    }
  }
}

// Exception class for BI Report errors
class BiReportException implements Exception {
  final String message;
  final int code;

  const BiReportException({required this.message, required this.code});

  @override
  String toString() => 'BiReportException($code): $message';
}
