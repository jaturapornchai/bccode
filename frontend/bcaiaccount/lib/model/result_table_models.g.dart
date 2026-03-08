// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'result_table_models.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ResultFromQueryRequest _$ResultFromQueryRequestFromJson(
  Map<String, dynamic> json,
) => ResultFromQueryRequest(
  shopid: json['shopid'] as String,
  query: json['query'] as String,
  guid: json['guid'] as String?,
);

Map<String, dynamic> _$ResultFromQueryRequestToJson(
  ResultFromQueryRequest instance,
) => <String, dynamic>{
  'shopid': instance.shopid,
  'query': instance.query,
  'guid': instance.guid,
};

ResultGetRequest _$ResultGetRequestFromJson(Map<String, dynamic> json) =>
    ResultGetRequest(
      shopid: json['shopid'] as String,
      guid: json['guid'] as String,
      limit: (json['limit'] as num?)?.toInt() ?? 100,
      offset: (json['offset'] as num?)?.toInt() ?? 0,
    );

Map<String, dynamic> _$ResultGetRequestToJson(ResultGetRequest instance) =>
    <String, dynamic>{
      'shopid': instance.shopid,
      'guid': instance.guid,
      'limit': instance.limit,
      'offset': instance.offset,
    };

PdfConfig _$PdfConfigFromJson(Map<String, dynamic> json) => PdfConfig(
  title: json['title'] as String?,
  orientation: json['orientation'] as String? ?? 'P',
  pageSize: json['page_size'] as String? ?? 'A4',
);

Map<String, dynamic> _$PdfConfigToJson(PdfConfig instance) => <String, dynamic>{
  'title': instance.title,
  'orientation': instance.orientation,
  'page_size': instance.pageSize,
};

ResultToPdfRequest _$ResultToPdfRequestFromJson(Map<String, dynamic> json) =>
    ResultToPdfRequest(
      shopid: json['shopid'] as String,
      guid: json['guid'] as String,
      pdfConfig: PdfConfig.fromJson(json['pdf_config'] as Map<String, dynamic>),
      columnOrder: (json['column_order'] as List<dynamic>)
          .map((e) => e as String)
          .toList(),
      columnNames: Map<String, String>.from(json['column_names'] as Map),
      layoutConfig: json['layout_config'] as Map<String, dynamic>?,
    );

Map<String, dynamic> _$ResultToPdfRequestToJson(ResultToPdfRequest instance) =>
    <String, dynamic>{
      'shopid': instance.shopid,
      'guid': instance.guid,
      'pdf_config': instance.pdfConfig,
      'column_order': instance.columnOrder,
      'column_names': instance.columnNames,
      'layout_config': instance.layoutConfig,
    };

ResultFromQueryResponse _$ResultFromQueryResponseFromJson(
  Map<String, dynamic> json,
) => ResultFromQueryResponse(
  status: json['status'] as String,
  guid: json['guid'] as String,
  count: (json['count'] as num).toInt(),
  message: json['message'] as String,
);

Map<String, dynamic> _$ResultFromQueryResponseToJson(
  ResultFromQueryResponse instance,
) => <String, dynamic>{
  'status': instance.status,
  'guid': instance.guid,
  'count': instance.count,
  'message': instance.message,
};

PaginationInfo _$PaginationInfoFromJson(Map<String, dynamic> json) =>
    PaginationInfo(
      limit: (json['limit'] as num).toInt(),
      offset: (json['offset'] as num).toInt(),
      count: (json['count'] as num).toInt(),
    );

Map<String, dynamic> _$PaginationInfoToJson(PaginationInfo instance) =>
    <String, dynamic>{
      'limit': instance.limit,
      'offset': instance.offset,
      'count': instance.count,
    };

ResultGetPageResponse _$ResultGetPageResponseFromJson(
  Map<String, dynamic> json,
) => ResultGetPageResponse(
  status: json['status'] as String,
  data: (json['data'] as List<dynamic>)
      .map((e) => e as Map<String, dynamic>)
      .toList(),
  pagination: PaginationInfo.fromJson(
    json['pagination'] as Map<String, dynamic>,
  ),
  sections:
      (json['sections'] as List<dynamic>?)
          ?.map((e) => e as Map<String, dynamic>)
          .toList() ??
      const [],
);

Map<String, dynamic> _$ResultGetPageResponseToJson(
  ResultGetPageResponse instance,
) => <String, dynamic>{
  'status': instance.status,
  'data': instance.data,
  'pagination': instance.pagination,
  'sections': instance.sections,
};
