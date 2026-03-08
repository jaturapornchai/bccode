import 'dart:convert';

import 'package:smlaicloud/model/color_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class ColorRepository {
  Future<ApiResponse> getColorList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query = "/color/list?offset=$offset&limit=$limit&q=$search";
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
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> deleteColor(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/color/$guid');
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

  /// ลบที่ละหลาย GUID
  Future<ApiResponse> deleteColorMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/color', data: guids);
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

  Future<ApiResponse> getColor(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/color/$guid');
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

  Future<ApiResponse> saveColor(ColorModel colorModel) async {
    Dio client = Client().init();
    final data = colorModel.toJson();
    try {
      final response = await client.post('/color', data: data);
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

  Future<ApiResponse> updateColor(String guid, ColorModel colorModel) async {
    Dio client = Client().init();
    final data = colorModel.toJson();
    try {
      final response = await client.put('/color/$guid', data: data);
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
}
