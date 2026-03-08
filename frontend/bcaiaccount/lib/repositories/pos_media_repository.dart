import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';
import 'package:flutter/foundation.dart';
import 'package:smlaicloud/model/pos_media_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class PosMediaRepository {
  Future<ApiResponse> getPosList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query = "/pos/media/list?offset=$offset&limit=$limit&q=$search";
      final response = await client.get(query);
      try {
        final rawData = json.decode(response.toString());
        if (rawData['error'] != null) {
          throw Exception('${rawData['code']}: ${rawData['message']}');
        }
        return ApiResponse.fromMap(rawData);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw errorMessage;
    }
  }

  Future<ApiResponse> deletePos(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/pos/media/$guid');
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

  /// ลบที่ละหลาย GUID
  Future<ApiResponse> deletePosMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guids);
      final response = await client.delete('/pos/media', data: guidStrings);
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

  Future<ApiResponse> getPos(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/pos/media/$guid');
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

  Future<ApiResponse> savePos(PosMediaModel posModel) async {
    Dio client = Client().init();
    final data = posModel.toJson();
    try {
      final response = await client.post('/pos/media', data: data);
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

  Future<ApiResponse> updatePos(String guid, PosMediaModel posModel) async {
    Dio client = Client().init();
    final data = posModel.toJson();
    // print(data);
    try {
      final response = await client.put('/pos/media/$guid', data: data);
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
        rethrow;
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      // print(ex);
      throw errorMessage;
    }
  }
}
