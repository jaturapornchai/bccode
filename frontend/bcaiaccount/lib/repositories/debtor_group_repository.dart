import 'dart:convert';

import 'package:smlaicloud/model/debtor_group_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class DebtorGroupRepository {
  Future<ApiResponse> getDebtorGroupList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/debtaccount/debtor-group/list?offset=$offset&limit=$limit&q=$search";
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

  Future<ApiResponse> deleteDebtorGroup(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/debtaccount/debtor-group/$guid');
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
  Future<ApiResponse> deleteDebtorGroupMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guids);
      final response =
          await client.delete('/debtaccount/debtor-group', data: guidStrings);
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

  Future<ApiResponse> getDebtorGroup(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/debtaccount/debtor-group/$guid');
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

  Future<ApiResponse> saveDebtorGroup(DebtorGroupModel debtorGroupModel) async {
    Dio client = Client().init();
    final data = debtorGroupModel.toJson();
    try {
      final response =
          await client.post('/debtaccount/debtor-group', data: data);
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

  Future<ApiResponse> updateDebtorGroup(
      String guid, DebtorGroupModel debtorGroupModel) async {
    Dio client = Client().init();
    final data = debtorGroupModel.toJson();
    try {
      final response =
          await client.put('/debtaccount/debtor-group/$guid', data: data);
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
