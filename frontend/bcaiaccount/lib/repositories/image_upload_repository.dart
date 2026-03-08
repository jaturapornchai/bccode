import 'dart:io';
import 'package:smlaicloud/model/global_model.dart';
import 'package:flutter/foundation.dart';
import 'client.dart';
import 'package:dio/dio.dart';

class ImageUploadRepository {
  Future<ApiResponse> uploadImage(ImageUpload imageupload) async {
    Dio client = Client().init();
    // final data = imageupload.toJson();
    // print(data);

    FormData formData = FormData.fromMap({
      "file": await MultipartFile.fromFile(imageupload.uri),
    });

    try {
      final response = await client.post('/upload/images', data: formData);
      try {
        // print(response.data);

        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        // print(ex);
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      // print(ex);
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> uploadImageFile(File file, Uint8List image) async {
    Dio client = Client().init();
    String fileName = file.path.split('/').last;
    FormData formData = FormData.fromMap({
      "file": MultipartFile.fromBytes(image, filename: '$fileName.png'),
    });
    try {
      final response = await client.post('/upload/images', data: formData);
      try {
        // print(response.data);
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        // print(ex);
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      // print(ex);
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> importImageProduct(
    File file,
    Uint8List image,
    String filename,
  ) async {
    Dio client = Client().init();
    FormData formData = FormData.fromMap({
      "file": MultipartFile.fromBytes(image, filename: filename),
    });
    try {
      final response = await client.post(
        '/upload/productimage',
        data: formData,
      );
      try {
        // print(response.data);
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        // print(ex);
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      // print(ex);
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> imageUploadResposneUri(
    File file,
    Uint8List image,
    String filename,
  ) async {
    Dio client = Client().init();
    FormData formData = FormData.fromMap({
      "file": MultipartFile.fromBytes(image, filename: filename),
    });
    try {
      final response = await client.post('/upload/images', data: formData);
      try {
        // print(response.data);
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        // print(ex);
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      // print(ex);
      throw Exception(errorMessage);
    }
  }

  /// videoUploadResposneUri
  Future<ApiResponse> videoUploadResponseUri({
    File? file,
    Uint8List? videoWeb,
    required String filename,
  }) async {
    Dio dio = Dio();
    FormData formData = FormData();

    // Check if we're on the web. If so, use the Uint8List, otherwise use the File.
    if (kIsWeb) {
      if (videoWeb != null) {
        formData = FormData.fromMap({
          "file": MultipartFile.fromBytes(videoWeb, filename: filename),
        });
      } else {
        throw Exception("File data must be provided on the web");
      }
    } else {
      if (file != null) {
        formData = FormData.fromMap({
          "file": await MultipartFile.fromFile(file.path, filename: filename),
        });
      } else {
        throw Exception("File must be provided on mobile");
      }
    }

    try {
      final response = await dio.post(
        'https://dedecafe.com/upload',
        data: formData,
      );

      // Assuming ApiResponse.fromMap is a method to parse your specific API response
      return ApiResponse.fromMap(response.data);
    } on DioException catch (ex) {
      // Handle DioException, including timeout, no connection, etc.
      String errorMessage =
          "Error during file upload: ${ex.response?.data ?? 'Unknown error'}";
      throw Exception(errorMessage);
    }
  }
}
