import 'dart:io';
import 'package:dio/dio.dart';
import 'package:smlaicloud/model/attachment_model.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/global.dart' as global;

class AttachmentApiService {
  final Dio _dio = Dio();

  /// Base URL สำหรับ goapi (attachment endpoints)
  static String get _baseUrl => global.goApiBaseUrl;

  /// Upload attachment
  /// [file] - File to upload
  /// [shopId] - Shop ID
  /// [screenType] - Screen type (purchaseorder, sale, purchase, etc.)
  /// [docNo] - Document number
  /// [guidFixed] - Document GUID
  /// [uploadedBy] - User code
  /// [uploadedName] - User name (optional)
  /// [description] - File description (optional)
  Future<AttachmentUploadResponse> uploadAttachment({
    required File file,
    required String shopId,
    required String screenType,
    required String docNo,
    required String guidFixed,
    required String uploadedBy,
    String? uploadedName,
    String? description,
  }) async {
    try {
      final formData = FormData.fromMap({
        'file': await MultipartFile.fromFile(file.path, filename: file.path.split('/').last),
        'shopid': shopId,
        'screen_type': screenType,
        'docno': docNo,
        'guidfixed': guidFixed,
        'uploaded_by': uploadedBy,
        if (uploadedName != null) 'uploaded_name': uploadedName,
        if (description != null && description.isNotEmpty) 'description': description,
      });

      final response = await _dio.post(
        '$_baseUrl/api/attachment/upload',
        data: formData,
        options: Options(
          headers: {'Content-Type': 'multipart/form-data'},
        ),
      );

      AppLogger.info('Upload attachment response: ${response.data}');
      return AttachmentUploadResponse.fromJson(response.data);
    } catch (e) {
      AppLogger.error('Failed to upload attachment: $e');
      rethrow;
    }
  }

  /// List attachments
  /// [shopId] - Shop ID (required)
  /// [screenType] - Filter by screen type (optional)
  /// [docNo] - Filter by document number (optional)
  /// [guidFixed] - Filter by GUID (optional)
  Future<AttachmentListResponse> listAttachments({
    required String shopId,
    String? screenType,
    String? docNo,
    String? guidFixed,
  }) async {
    try {
      final requestBody = {
        'shopid': shopId,
        if (screenType != null && screenType.isNotEmpty) 'screen_type': screenType,
        if (docNo != null && docNo.isNotEmpty) 'docno': docNo,
        if (guidFixed != null && guidFixed.isNotEmpty) 'guidfixed': guidFixed,
      };

      final response = await _dio.post(
        '$_baseUrl/api/attachment/list',
        data: requestBody,
        options: Options(
          headers: {'Content-Type': 'application/json'},
        ),
      );

      AppLogger.info('List attachments response: count=${response.data['count']}');
      return AttachmentListResponse.fromJson(response.data);
    } catch (e) {
      AppLogger.error('Failed to list attachments: $e');
      rethrow;
    }
  }

  /// Delete attachment
  /// [shopId] - Shop ID (required)
  /// [attachmentId] - Attachment ID (required)
  Future<bool> deleteAttachment({
    required String shopId,
    required String attachmentId,
  }) async {
    try {
      final response = await _dio.post(
        '$_baseUrl/api/attachment/delete',
        data: {
          'shopid': shopId,
          'attachment_id': attachmentId,
        },
        options: Options(
          headers: {'Content-Type': 'application/json'},
        ),
      );

      AppLogger.info('Delete attachment response: ${response.data}');
      return response.data['status'] == 'success';
    } catch (e) {
      AppLogger.error('Failed to delete attachment: $e');
      rethrow;
    }
  }

  /// Download attachment (รองรับทั้ง full URL และ relative path /s3/file/...)
  String? getDownloadUrl(AttachmentModel attachment) {
    if (attachment.url == null) return null;
    return global.resolveFileUrl(attachment.url!);
  }
}
