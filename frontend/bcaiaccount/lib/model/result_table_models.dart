/// Models สำหรับ Result Table API
/// รองรับ 3 endpoints: resultfromquery, resultget, resulttopdf
library;

import 'package:json_annotation/json_annotation.dart';
import '../global.dart' as global;

part 'result_table_models.g.dart';

// ===== Request Models =====

/// Request model สำหรับ /resultfromquery
@JsonSerializable()
class ResultFromQueryRequest {
  final String shopid;
  final String query;
  final String? guid;

  ResultFromQueryRequest({
    required this.shopid,
    required this.query,
    this.guid,
  });

  factory ResultFromQueryRequest.fromJson(Map<String, dynamic> json) =>
      _$ResultFromQueryRequestFromJson(json);

  Map<String, dynamic> toJson() => _$ResultFromQueryRequestToJson(this);
}

/// Request model สำหรับ /resultget
@JsonSerializable()
class ResultGetRequest {
  final String shopid;
  final String guid;
  final int limit;
  final int offset;

  ResultGetRequest({
    required this.shopid,
    required this.guid,
    this.limit = 100,
    this.offset = 0,
  });

  factory ResultGetRequest.fromJson(Map<String, dynamic> json) =>
      _$ResultGetRequestFromJson(json);

  Map<String, dynamic> toJson() => _$ResultGetRequestToJson(this);
}

/// PDF Configuration
@JsonSerializable()
class PdfConfig {
  final String title;
  final String orientation;
  @JsonKey(name: 'page_size')
  final String pageSize;

  PdfConfig({
    String? title,
    this.orientation = 'P',
    this.pageSize = 'A4',
  }) : title = title ?? global.language('report');

  factory PdfConfig.fromJson(Map<String, dynamic> json) =>
      _$PdfConfigFromJson(json);

  Map<String, dynamic> toJson() => _$PdfConfigToJson(this);
}

/// Request model สำหรับ /resulttopdf
@JsonSerializable()
class ResultToPdfRequest {
  final String shopid;
  final String guid;
  @JsonKey(name: 'pdf_config')
  final PdfConfig pdfConfig;
  @JsonKey(name: 'column_order')
  final List<String> columnOrder;
  @JsonKey(name: 'column_names')
  final Map<String, String> columnNames;
  @JsonKey(name: 'layout_config')
  final Map<String, dynamic>? layoutConfig;

  ResultToPdfRequest({
    required this.shopid,
    required this.guid,
    required this.pdfConfig,
    required this.columnOrder,
    required this.columnNames,
    this.layoutConfig,
  });

  factory ResultToPdfRequest.fromJson(Map<String, dynamic> json) =>
      _$ResultToPdfRequestFromJson(json);

  Map<String, dynamic> toJson() => _$ResultToPdfRequestToJson(this);
}

// ===== Response Models =====

/// Response model สำหรับ /resultfromquery
@JsonSerializable()
class ResultFromQueryResponse {
  final String status;
  final String guid;
  final int count;
  final String message;

  ResultFromQueryResponse({
    required this.status,
    required this.guid,
    required this.count,
    required this.message,
  });

  factory ResultFromQueryResponse.fromJson(Map<String, dynamic> json) =>
      _$ResultFromQueryResponseFromJson(json);

  Map<String, dynamic> toJson() => _$ResultFromQueryResponseToJson(this);
}

/// Pagination information for `/resultget`
@JsonSerializable()
class PaginationInfo {
  final int limit;
  final int offset;
  final int count;

  const PaginationInfo({
    required this.limit,
    required this.offset,
    required this.count,
  });

  /// Maintain backward-compatible access to the total value.
  int get total => count;

  factory PaginationInfo.fromJson(Map<String, dynamic> json) =>
      _$PaginationInfoFromJson(json);

  Map<String, dynamic> toJson() => _$PaginationInfoToJson(this);
}

/// Minimal response model for `/resultget`
@JsonSerializable()
class ResultGetPageResponse {
  final String status;
  final List<Map<String, dynamic>> data;
  final PaginationInfo pagination;
  final List<Map<String, dynamic>> sections;

  const ResultGetPageResponse({
    required this.status,
    required this.data,
    required this.pagination,
    this.sections = const [],
  });

  factory ResultGetPageResponse.fromJson(Map<String, dynamic> json) {
    final rawData = json['data'] ?? json['rows'];
    final decodedData = _decodeMapList(rawData);
    final paginationJson = json['pagination'];
    final decodedSections = _decodeMapList(json['sections']);

    final pagination = paginationJson is Map<String, dynamic>
        ? PaginationInfo.fromJson(paginationJson)
        : PaginationInfo(
            limit: decodedData.length,
            offset: 0,
            count: decodedData.length,
          );

    return ResultGetPageResponse(
      status: json['status']?.toString() ?? 'unknown',
      data: decodedData,
      pagination: pagination,
      sections: decodedSections,
    );
  }

  Map<String, dynamic> toJson() => {
    'status': status,
    'data': data,
    'pagination': pagination.toJson(),
    if (sections.isNotEmpty) 'sections': sections,
  };

  static List<Map<String, dynamic>> _decodeMapList(dynamic raw) {
    if (raw is! List) {
      return <Map<String, dynamic>>[];
    }

    final decoded = <Map<String, dynamic>>[];
    for (final item in raw) {
      if (item is Map<String, dynamic>) {
        decoded.add(item);
      } else if (item is Map) {
        decoded.add(Map<String, dynamic>.from(item));
      }
    }
    return decoded;
  }
}
