import 'dart:convert';

import 'package:smlaicloud/model/business_type_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class BusinessTypeRepository {
  Future<ApiResponse> getBusinessTypeList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/organization/business-type/list?offset=$offset&limit=$limit&q=$search";
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

  Future<ApiResponse> deleteBusinessType(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/organization/business-type/$guid');
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
  Future<ApiResponse> deleteBusinessTypeMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final response =
          await client.delete('/organization/business-type', data: guids);
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

  Future<ApiResponse> getBusinessType(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/organization/business-type/$guid');
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

  Future<ApiResponse> saveBusinessType(
      BusinessTypeModel businessTypeModel) async {
    Dio client = Client().init();
    final data = businessTypeModel.toJson();
    try {
      final response =
          await client.post('/organization/business-type', data: data);
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

  Future<ApiResponse> updateBusinessType(
      String guid, BusinessTypeModel businessTypeModel) async {
    Dio client = Client().init();
    final data = businessTypeModel.toJson();
    try {
      final response =
          await client.put('/organization/business-type/$guid', data: data);
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
