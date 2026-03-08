import 'dart:convert';

import 'package:smlaicloud/model/master_pattern_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class MasterPatternRepository {
  Future<ApiResponse> getPatternManyByCode(List<String> codeList) async {
    Dio client = Client().init();
    try {
      final codeJson = jsonEncode(codeList);
      final response =
          await client.get('/aicloud/pattern/by-code?codes=$codeJson');
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

  Future<ApiResponse> getPatternList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/aicloud/pattern/list?offset=$offset&limit=$limit&q=$search&sort=code:1";
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

  Future<ApiResponse> deletePattern(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/aicloud/pattern/$guid');
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

  Future<ApiResponse> deletePatternMany(List<String> guidList) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guidList);
      final response =
          await client.delete('/aicloud/pattern', data: guidStrings);
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

  Future<ApiResponse> getPattern(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/aicloud/pattern/$guid');
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

  Future<ApiResponse> getPatternByCode(String code) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/aicloud/pattern/code/$code');
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

  Future<ApiResponse> savePattern(MasterPatternModel patternModel) async {
    Dio client = Client().init();
    final data = patternModel.toJson();
    try {
      final response = await client.post('/aicloud/pattern', data: data);
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

  Future<ApiResponse> updatePattern(
      String guid, MasterPatternModel patternModel) async {
    Dio client = Client().init();
    final data = patternModel.toJson();
    try {
      final response = await client.put('/aicloud/pattern/$guid', data: data);
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
