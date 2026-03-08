import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class JsonRepository {
  Future<ApiResponse> getSetting(String code, String search) async {
    Dio client = Client().init();
    try {
      final response = await client.get(
        '/restaurant/settings/code/$code?q=$search',
      );
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> saveSetting(postData) async {
    Dio client = Client().init();
    try {
      final response = await client.post(
        '/restaurant/settings',
        data: postData,
      );
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> saveSettingBulk(postData) async {
    Dio client = Client().init();
    try {
      final jsonString = jsonEncode(postData);
      final response = await client.post(
        '/restaurant/settings/bulk',
        data: jsonString,
      );
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw errorMessage;
    }
  }

  Future<ApiResponse> updateSetting(String guid, postData) async {
    Dio client = Client().init();

    try {
      final response = await client.put(
        '/restaurant/settings/$guid',
        data: postData,
      );
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> deleteSetting(String guid) async {
    Dio client = Client().init();

    try {
      final response = await client.delete('/restaurant/settings/$guid');
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> deleteManySetting(List<String> guids) async {
    Dio client = Client().init();

    try {
      final guidStrings = jsonEncode(guids);
      final response = await client.delete(
        '/restaurant/settings',
        data: guidStrings,
      );
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> uploadImage(File file, Uint8List image) async {
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
}
