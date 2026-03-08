import 'dart:convert';

import 'package:smlaicloud/model/creditor_group_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class CreditorGroupRepository {
  Future<ApiResponse> getCreditorGroupList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/debtaccount/creditor-group/list?offset=$offset&limit=$limit&q=$search";
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

  Future<ApiResponse> deleteCreditorGroup(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/debtaccount/creditor-group/$guid');
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
  Future<ApiResponse> deleteCreditorGroupMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guids);
      final response =
          await client.delete('/debtaccount/creditor-group', data: guidStrings);
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

  Future<ApiResponse> getCreditorGroup(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/debtaccount/creditor-group/$guid');
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

  Future<ApiResponse> saveCreditorGroup(
      CreditorGroupModel creditorGroupModel) async {
    Dio client = Client().init();
    final data = creditorGroupModel.toJson();
    try {
      final response =
          await client.post('/debtaccount/creditor-group', data: data);
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

  Future<ApiResponse> updateCreditorGroup(
      String guid, CreditorGroupModel creditorGroupModel) async {
    Dio client = Client().init();
    final data = creditorGroupModel.toJson();
    try {
      final response =
          await client.put('/debtaccount/creditor-group/$guid', data: data);
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
