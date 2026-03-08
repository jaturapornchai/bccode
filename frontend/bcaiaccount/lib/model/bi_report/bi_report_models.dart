import 'package:json_annotation/json_annotation.dart';
import '../../global.dart' as global;

part 'bi_report_models.g.dart';

// Report Types Enum
enum BiReportType {
  /// รายงานยอดขาย
  sale,

  /// รายงานยอดขายตามวันที่
  saleDaily,

  /// รายงานการเคลื่อนไหวของสต็อก
  stockMovement,

  /// รายงานรับเงิน ตามวันที่
  paymentDaily,

  /// รายงานลดหนี้/รับคืน
  saleReturn,

  /// รายงานสินค้าคงเหลือ
  stockBalance,

  /// รายงานรับสินค้าแบบทยอยรับ
  purchasepartial,

  /// รายงานกำไรขั้นต้นตามเอกสาร
  grossProfitByDocument,

  /// รายงานกำไรขั้นต้นตามสินค้า
  grossProfitByProduct,

  /// รายงานภาษีขาย
  vatSale,

  /// รายงานภาษีซื้อ
  vatBuy,
}

// Report Type Extension
extension BiReportTypeExtension on BiReportType {
  String get endpoint {
    switch (this) {
      case BiReportType.sale:
        return '/salereport';
      case BiReportType.saleDaily:
        return '/saledaily';
      case BiReportType.stockMovement:
        return '/stockmovement';
      case BiReportType.paymentDaily:
        return '/paymentdaily';
      case BiReportType.saleReturn:
        return '/sale_return';
      case BiReportType.stockBalance:
        return '/stockbalance';
      case BiReportType.purchasepartial:
        return '/purchasepartial';
      case BiReportType.grossProfitByDocument:
        return '/grossprofitbydocument';
      case BiReportType.grossProfitByProduct:
        return '/grossprofitbyproduct';
      case BiReportType.vatSale:
        return '/vatsale';
      case BiReportType.vatBuy:
        return '/vatbuy';
    }
  }

  String get displayName {
    switch (this) {
      case BiReportType.sale:
        return global.language('report_dedebi_sales_title');
      case BiReportType.saleDaily:
        return 'รายงานยอดขายตามวันที่';
      case BiReportType.stockMovement:
        return 'รายงานการเคลื่อนไหวของสต็อก';
      case BiReportType.paymentDaily:
        return global.language('payment_received_report_by_date');
      case BiReportType.saleReturn:
        return global.language('debt_reduction_return_report');
      case BiReportType.stockBalance:
        return global.language('inventory_report');
      case BiReportType.purchasepartial:
        return global.language('gradual_receipt_report');
      case BiReportType.grossProfitByDocument:
        return global.language('gross_profit_by_document_report');
      case BiReportType.grossProfitByProduct:
        return global.language('gross_profit_by_product_report');
      case BiReportType.vatSale:
        return global.language('report_vat_sale');
      case BiReportType.vatBuy:
        return global.language('report_vat_buy');
    }
  }
}

@JsonSerializable()
class ReportConditionsModel {
  final String fromdate;
  final String todate;
  final String branchcode;
  final bool showdetail;
  final String iscancel;
  final String
  inquirytype; // เปลี่ยนเป็น String: "" = ทั้งหมด, "1" = ขาย, "2" = คืน
  final bool ispos;
  final String creditorcode;
  final String salecode;
  final String debtorcode;
  final String barcode;

  const ReportConditionsModel({
    this.fromdate = '',
    this.todate = '',
    this.branchcode = '',
    this.showdetail = true,
    this.iscancel = '',
    this.inquirytype = '', // เปลี่ยน default เป็น '' (ทั้งหมด)
    this.ispos = false,
    this.creditorcode = '',
    this.salecode = '',
    this.debtorcode = '',
    this.barcode = '',
  });

  factory ReportConditionsModel.fromJson(Map<String, dynamic> json) =>
      _$ReportConditionsModelFromJson(json);

  Map<String, dynamic> toJson() => _$ReportConditionsModelToJson(this);
}

// Base classes for BI report API responses

// Job submission response (Step 1)
@JsonSerializable()
class BiReportJobResponse {
  final String status;
  @JsonKey(name: 'job_id')
  final String jobId;
  final String message;

  const BiReportJobResponse({
    required this.status,
    required this.jobId,
    required this.message,
  });

  factory BiReportJobResponse.fromJson(Map<String, dynamic> json) =>
      _$BiReportJobResponseFromJson(json);

  Map<String, dynamic> toJson() => _$BiReportJobResponseToJson(this);
}

// Job status response (Step 2)
@JsonSerializable()
class BiReportStatusResponse {
  final String status;
  final BiReportStatusData data;

  const BiReportStatusResponse({required this.status, required this.data});

  factory BiReportStatusResponse.fromJson(Map<String, dynamic> json) =>
      _$BiReportStatusResponseFromJson(json);

  Map<String, dynamic> toJson() => _$BiReportStatusResponseToJson(this);
}

@JsonSerializable()
class BiReportStatusData {
  final bool success;
  @JsonKey(name: 'job_id')
  final String jobId;
  final String state;
  final int progress;
  final String createdAt;
  final String? processedOn;
  final String? finishedOn;
  final String? failedReason; // เพิ่ม field สำหรับ error message

  const BiReportStatusData({
    required this.success,
    required this.jobId,
    required this.state,
    required this.progress,
    required this.createdAt,
    this.processedOn,
    this.finishedOn,
    this.failedReason, // เพิ่มใน constructor
  });

  factory BiReportStatusData.fromJson(Map<String, dynamic> json) =>
      _$BiReportStatusDataFromJson(json);

  Map<String, dynamic> toJson() => _$BiReportStatusDataToJson(this);
}

// Generic report detail response (Step 3)
@JsonSerializable(genericArgumentFactories: true)
class BiReportDetailResponse<T> {
  final String status;
  final List<T> data;
  final BiReportMeta meta; // เปลี่ยนกลับเป็น required

  const BiReportDetailResponse({
    required this.status,
    required this.data,
    required this.meta, // เปลี่ยนกลับเป็น required
  });

  factory BiReportDetailResponse.fromJson(
    Map<String, dynamic> json,
    T Function(Object? json) fromJsonT,
  ) => _$BiReportDetailResponseFromJson(json, fromJsonT);

  Map<String, dynamic> toJson(Object Function(T value) toJsonT) =>
      _$BiReportDetailResponseToJson(this, toJsonT);
}

@JsonSerializable()
class BiReportMeta {
  final int page;
  final int size;
  final int total;
  @JsonKey(name: 'total_page')
  final int totalPage;

  const BiReportMeta({
    required this.page,
    required this.size,
    required this.total,
    required this.totalPage,
  });

  factory BiReportMeta.fromJson(Map<String, dynamic> json) =>
      _$BiReportMetaFromJson(json);

  Map<String, dynamic> toJson() => _$BiReportMetaToJson(this);
}

// Error Response Model
@JsonSerializable()
class BiReportErrorResponse {
  final int code;
  final String message;

  const BiReportErrorResponse({required this.code, required this.message});

  factory BiReportErrorResponse.fromJson(Map<String, dynamic> json) =>
      _$BiReportErrorResponseFromJson(json);

  Map<String, dynamic> toJson() => _$BiReportErrorResponseToJson(this);
}
