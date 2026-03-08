import 'dart:convert';

import 'package:smlaicloud/model/report_list_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class FileStatusRepository {
  Future<ApiResponse> getFileStatue({
    int limit = 0,
    int offset = 0,
    String menu = "",
  }) async {
    Dio client = Client().init();
    try {
      String query = "/file-status/list?offset=$offset&limit=$limit&menu=$menu";
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

  Future<ApiResponse> deleteFileStatueById({required String guid}) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/file-status/$guid');
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

  Future<ApiResponse> deleteFileStatueByMenu({required String menu}) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/file-status/menu/$menu');
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

  Future<ApiResponse> saveFileStatue(
      LogDownloadParthModel logDownloadParthModel) async {
    Dio client = Client().init();
    final data = logDownloadParthModel.toJson();
    try {
      final response = await client.post('/file-status', data: data);
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

  Future<ApiResponse> updateFileStatue(
      LogDownloadParthModel logDownloadParthModel, String guid) async {
    Dio client = Client().init();
    final data = logDownloadParthModel.toJson();
    try {
      final response = await client.put('/file-status/$guid', data: data);
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
}
