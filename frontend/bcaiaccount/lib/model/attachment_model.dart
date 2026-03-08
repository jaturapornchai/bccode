import 'package:json_annotation/json_annotation.dart';

part 'attachment_model.g.dart';

@JsonSerializable()
class AttachmentModel {
  @JsonKey(name: 'id')
  String? id;

  @JsonKey(name: 'shopid')
  String shopid;

  @JsonKey(name: 'screen_type')
  String screenType;

  @JsonKey(name: 'docno')
  String docno;

  @JsonKey(name: 'guidfixed')
  String guidfixed;

  @JsonKey(name: 'filename')
  String filename;

  @JsonKey(name: 'original_name')
  String originalName;

  @JsonKey(name: 'content_type')
  String contentType;

  @JsonKey(name: 'file_type')
  String fileType;

  @JsonKey(name: 'size')
  int size;

  @JsonKey(name: 'description')
  String? description;

  @JsonKey(name: 'uploaded_by')
  String uploadedBy;

  @JsonKey(name: 'uploaded_name')
  String? uploadedName;

  @JsonKey(name: 'created_at')
  String createdAt;

  @JsonKey(name: 'updated_at')
  String updatedAt;

  @JsonKey(name: 'url')
  String? url; // Presigned URL

  AttachmentModel({
    this.id,
    required this.shopid,
    required this.screenType,
    required this.docno,
    required this.guidfixed,
    required this.filename,
    required this.originalName,
    required this.contentType,
    required this.fileType,
    required this.size,
    this.description,
    required this.uploadedBy,
    this.uploadedName,
    required this.createdAt,
    required this.updatedAt,
    this.url,
  });

  factory AttachmentModel.fromJson(Map<String, dynamic> json) => _$AttachmentModelFromJson(json);

  Map<String, dynamic> toJson() => _$AttachmentModelToJson(this);
}

@JsonSerializable()
class AttachmentListResponse {
  @JsonKey(name: 'status')
  String status;

  @JsonKey(name: 'code')
  int code;

  @JsonKey(name: 'count')
  int count;

  @JsonKey(name: 'total')
  int total;

  @JsonKey(name: 'data', defaultValue: [])
  List<AttachmentModel> data;

  AttachmentListResponse({
    required this.status,
    required this.code,
    required this.count,
    required this.total,
    this.data = const [],
  });

  factory AttachmentListResponse.fromJson(Map<String, dynamic> json) => _$AttachmentListResponseFromJson(json);

  Map<String, dynamic> toJson() => _$AttachmentListResponseToJson(this);
}

@JsonSerializable()
class AttachmentUploadResponse {
  @JsonKey(name: 'status')
  String status;

  @JsonKey(name: 'code')
  int code;

  @JsonKey(name: 'message')
  String? message;

  @JsonKey(name: 'data')
  AttachmentModel? data;

  AttachmentUploadResponse({
    required this.status,
    required this.code,
    this.message,
    this.data,
  });

  factory AttachmentUploadResponse.fromJson(Map<String, dynamic> json) => _$AttachmentUploadResponseFromJson(json);

  Map<String, dynamic> toJson() => _$AttachmentUploadResponseToJson(this);
}
