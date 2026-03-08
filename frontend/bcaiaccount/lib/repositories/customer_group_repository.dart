import 'dart:convert';

import 'package:smlaicloud/model/customer_group_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class CustomerGroupRepository {
  Future<ApiResponse> getCustomerGroupList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/debtaccount/customer-group/list?offset=$offset&limit=$limit&q=$search";
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

  Future<ApiResponse> deleteCustomerGroup(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/debtaccount/customer-group/$guid');
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
  Future<ApiResponse> deleteCustomerGroupMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final response =
          await client.delete('/debtaccount/customer-group', data: guids);
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

  Future<ApiResponse> getCustomerGroup(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/debtaccount/customer-group/$guid');
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

  Future<ApiResponse> saveCustomerGroup(
      CustomerGroupModel customerGroupModel) async {
    Dio client = Client().init();
    final data = customerGroupModel.toJson();
    try {
      final response =
          await client.post('/debtaccount/customer-group', data: data);
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

  Future<ApiResponse> updateCustomerGroup(
      String guid, CustomerGroupModel customerGroupModel) async {
    Dio client = Client().init();
    final data = customerGroupModel.toJson();
    try {
      final response =
          await client.put('/debtaccount/customer-group/$guid', data: data);
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
