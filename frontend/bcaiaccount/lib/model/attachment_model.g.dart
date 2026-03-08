// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'attachment_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

AttachmentModel _$AttachmentModelFromJson(Map<String, dynamic> json) =>
    AttachmentModel(
      id: json['id'] as String?,
      shopid: json['shopid'] as String,
      screenType: json['screen_type'] as String,
      docno: json['docno'] as String,
      guidfixed: json['guidfixed'] as String,
      filename: json['filename'] as String,
      originalName: json['original_name'] as String,
      contentType: json['content_type'] as String,
      fileType: json['file_type'] as String,
      size: (json['size'] as num).toInt(),
      description: json['description'] as String?,
      uploadedBy: json['uploaded_by'] as String,
      uploadedName: json['uploaded_name'] as String?,
      createdAt: json['created_at'] as String,
      updatedAt: json['updated_at'] as String,
      url: json['url'] as String?,
    );

Map<String, dynamic> _$AttachmentModelToJson(AttachmentModel instance) =>
    <String, dynamic>{
      'id': instance.id,
      'shopid': instance.shopid,
      'screen_type': instance.screenType,
      'docno': instance.docno,
      'guidfixed': instance.guidfixed,
      'filename': instance.filename,
      'original_name': instance.originalName,
      'content_type': instance.contentType,
      'file_type': instance.fileType,
      'size': instance.size,
      'description': instance.description,
      'uploaded_by': instance.uploadedBy,
      'uploaded_name': instance.uploadedName,
      'created_at': instance.createdAt,
      'updated_at': instance.updatedAt,
      'url': instance.url,
    };

AttachmentListResponse _$AttachmentListResponseFromJson(
  Map<String, dynamic> json,
) => AttachmentListResponse(
  status: json['status'] as String,
  code: (json['code'] as num).toInt(),
  count: (json['count'] as num).toInt(),
  total: (json['total'] as num).toInt(),
  data:
      (json['data'] as List<dynamic>?)
          ?.map((e) => AttachmentModel.fromJson(e as Map<String, dynamic>))
          .toList() ??
      [],
);

Map<String, dynamic> _$AttachmentListResponseToJson(
  AttachmentListResponse instance,
) => <String, dynamic>{
  'status': instance.status,
  'code': instance.code,
  'count': instance.count,
  'total': instance.total,
  'data': instance.data,
};

AttachmentUploadResponse _$AttachmentUploadResponseFromJson(
  Map<String, dynamic> json,
) => AttachmentUploadResponse(
  status: json['status'] as String,
  code: (json['code'] as num).toInt(),
  message: json['message'] as String?,
  data: json['data'] == null
      ? null
      : AttachmentModel.fromJson(json['data'] as Map<String, dynamic>),
);

Map<String, dynamic> _$AttachmentUploadResponseToJson(
  AttachmentUploadResponse instance,
) => <String, dynamic>{
  'status': instance.status,
  'code': instance.code,
  'message': instance.message,
  'data': instance.data,
};
