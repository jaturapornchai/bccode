import 'dart:async';
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:crypto/crypto.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ChunkedUploadService {
  static const int defaultChunkSize = 5 * 1024 * 1024; // 5 MB
  static const int maxRetries = 3;

  String get baseUrl => global.goApiBaseUrl;

  // Upload from bytes (for web)
  Future<({String fileName, String fileUrl})> uploadFromBytes({
    required Uint8List bytes,
    required String fileName,
    required String shopId,
    int chunkSize = defaultChunkSize,
    Function(double progress, int uploadedChunks, int totalChunks)? onProgress,
  }) async {
    // Initialize upload
    final uploadId = await _initializeUpload(
      fileName: fileName,
      fileSize: bytes.length,
      shopId: shopId,
    );

    try {
      // Calculate total chunks
      final totalChunks = (bytes.length / chunkSize).ceil();

      // Upload chunks
      for (int i = 0; i < totalChunks; i++) {
        final start = i * chunkSize;
        final end = (start + chunkSize > bytes.length)
            ? bytes.length
            : start + chunkSize;

        final chunk = bytes.sublist(start, end);

        await _uploadChunkWithRetry(
          uploadId: uploadId,
          chunkData: chunk,
          chunkIndex: i,
          totalChunks: totalChunks,
        );

        // Update progress
        onProgress?.call((i + 1) / totalChunks, i + 1, totalChunks);
      }

      // Merge chunks
      return await _mergeChunks(uploadId, shopId);
    } catch (e) {
      // Cancel upload on error
      await cancelUpload(uploadId);
      rethrow;
    }
  }

  // Upload from stream (for mobile/desktop)
  Future<({String fileName, String fileUrl})> uploadFromStream({
    required Stream<List<int>> stream,
    required int fileSize,
    required String fileName,
    required String shopId,
    int chunkSize = defaultChunkSize,
    Function(double progress, int uploadedChunks, int totalChunks)? onProgress,
  }) async {
    // Initialize upload
    final uploadId = await _initializeUpload(
      fileName: fileName,
      fileSize: fileSize,
      shopId: shopId,
    );

    try {
      final totalChunks = (fileSize / chunkSize).ceil();
      int chunkIndex = 0;
      List<int> buffer = [];

      await for (final data in stream) {
        buffer.addAll(data);

        // Upload when buffer reaches chunk size or stream ends
        while (buffer.length >= chunkSize) {
          final chunk = buffer.sublist(0, chunkSize);
          buffer = buffer.sublist(chunkSize);

          await _uploadChunkWithRetry(
            uploadId: uploadId,
            chunkData: Uint8List.fromList(chunk),
            chunkIndex: chunkIndex,
            totalChunks: totalChunks,
          );

          chunkIndex++;
          onProgress?.call(chunkIndex / totalChunks, chunkIndex, totalChunks);
        }
      }

      // Upload remaining data
      if (buffer.isNotEmpty) {
        await _uploadChunkWithRetry(
          uploadId: uploadId,
          chunkData: Uint8List.fromList(buffer),
          chunkIndex: chunkIndex,
          totalChunks: totalChunks,
        );

        chunkIndex++;
        onProgress?.call(1.0, chunkIndex, totalChunks);
      }

      // Merge chunks
      return await _mergeChunks(uploadId, shopId);
    } catch (e) {
      // Cancel upload on error
      await cancelUpload(uploadId);
      rethrow;
    }
  }

  // Initialize upload session
  Future<String> _initializeUpload({
    required String fileName,
    required int fileSize,
    required String shopId,
  }) async {
    try {
      final uri = Uri.parse('$baseUrl/upload/init');

      if (fileSize <= 0) {
        throw Exception('File size must be greater than 0');
      }

      final response = await http.post(
        uri,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ${global.userLoginData.token}',
        },
        body: jsonEncode({
          'fileName': fileName,
          'totalSize': fileSize, // Changed from 'fileSize' to 'totalSize'
          'shopId': shopId,
        }),
      );

      if (response.statusCode != 200) {
        throw Exception('Failed to initialize upload: ${response.body}');
      }

      final data = jsonDecode(response.body);

      if (data['uploadID'] == null) {
        throw Exception('Upload ID not returned from server');
      }

      return data['uploadID'] as String;
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error initializing upload: $e');
      }
      rethrow;
    }
  }

  // Upload single chunk with retry logic
  Future<void> _uploadChunkWithRetry({
    required String uploadId,
    required Uint8List chunkData,
    required int chunkIndex,
    required int totalChunks,
    int retryCount = 0,
  }) async {
    try {
      await _uploadChunk(
        uploadId: uploadId,
        chunkData: chunkData,
        chunkIndex: chunkIndex,
        totalChunks: totalChunks,
      );
    } catch (e) {
      if (retryCount < maxRetries) {
        if (kDebugMode) {
          AppLogger.debug(
            'Retry chunk $chunkIndex (attempt ${retryCount + 1})',
          );
        }

        // Exponential backoff
        await Future.delayed(Duration(seconds: retryCount + 1));

        return _uploadChunkWithRetry(
          uploadId: uploadId,
          chunkData: chunkData,
          chunkIndex: chunkIndex,
          totalChunks: totalChunks,
          retryCount: retryCount + 1,
        );
      }
      rethrow;
    }
  }

  // Upload single chunk
  Future<void> _uploadChunk({
    required String uploadId,
    required Uint8List chunkData,
    required int chunkIndex,
    required int totalChunks,
  }) async {
    try {
      final uri = Uri.parse('$baseUrl/upload/chunk');

      // Calculate MD5 checksum
      final checksum = md5.convert(chunkData).toString();

      final request = http.MultipartRequest('POST', uri);

      request.headers.addAll({
        'Authorization': 'Bearer ${global.userLoginData.token}',
      });

      request.fields['uploadID'] = uploadId;
      request.fields['chunkIndex'] = chunkIndex.toString();
      request.fields['totalChunks'] = totalChunks.toString();
      request.fields['checksum'] = checksum;

      request.files.add(
        http.MultipartFile.fromBytes(
          'chunk',
          chunkData,
          filename: 'chunk_$chunkIndex',
        ),
      );

      final streamedResponse = await request.send();
      final response = await http.Response.fromStream(streamedResponse);

      if (response.statusCode != 200) {
        throw Exception('Failed to upload chunk $chunkIndex: ${response.body}');
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error uploading chunk $chunkIndex: $e');
      }
      rethrow;
    }
  }

  // Merge all chunks
  Future<({String fileName, String fileUrl})> _mergeChunks(String uploadId, String shopId) async {
    try {
      final uri = Uri.parse('$baseUrl/upload/merge');

      final response = await http.post(
        uri,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ${global.userLoginData.token}',
        },
        body: jsonEncode({'uploadID': uploadId, 'shopId': shopId}),
      );

      if (response.statusCode != 200) {
        throw Exception('Failed to merge chunks: ${response.body}');
      }

      final data = jsonDecode(response.body);

      if (data['fileName'] == null) {
        throw Exception('File name not returned from server');
      }

      return (
        fileName: data['fileName'] as String,
        fileUrl: (data['fileUrl'] as String?) ?? '',
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error merging chunks: $e');
      }
      rethrow;
    }
  }

  // Get upload status
  Future<Map<String, dynamic>> getUploadStatus(String uploadId) async {
    try {
      final uri = Uri.parse('$baseUrl/upload/status/$uploadId');

      final response = await http.get(
        uri,
        headers: {'Authorization': 'Bearer ${global.userLoginData.token}'},
      );

      if (response.statusCode != 200) {
        throw Exception('Failed to get upload status: ${response.body}');
      }

      return jsonDecode(response.body) as Map<String, dynamic>;
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error getting upload status: $e');
      }
      rethrow;
    }
  }

  // Cancel upload
  Future<void> cancelUpload(String uploadId) async {
    try {
      final uri = Uri.parse('$baseUrl/upload/cancel/$uploadId');

      final response = await http.delete(
        uri,
        headers: {'Authorization': 'Bearer ${global.userLoginData.token}'},
      );

      if (response.statusCode != 200) {
        if (kDebugMode) {
          AppLogger.error('Failed to cancel upload: ${response.body}');
        }
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error cancelling upload: $e');
      }
    }
  }

  void dispose() {
    // Cleanup if needed
  }
}
