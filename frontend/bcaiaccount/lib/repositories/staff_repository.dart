import 'dart:convert';
import 'package:smlaicloud/model/staff_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class StaffRepository {
  String url = "/restaurant/staff";

  Future<ApiResponse> getStaffList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query = "$url/list?offset=$offset&limit=$limit&q=$search";
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

  Future<ApiResponse> deleteStaff(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('$url/$guid');
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
  Future<ApiResponse> deleteStaffMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final response = await client.delete(url, data: guids);
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

  Future<ApiResponse> getStaff(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('$url/$guid');
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

  Future<ApiResponse> saveStaff(StaffModel printerModel) async {
    Dio client = Client().init();
    final data = printerModel.toJson();
    try {
      final response = await client.post(url, data: data);
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

  Future<ApiResponse> updateStaff(String guid, StaffModel printerModel) async {
    Dio client = Client().init();
    final data = printerModel.toJson();
    try {
      final response = await client.put('$url/$guid', data: data);
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
