import 'dart:convert';

import 'package:smlaicloud/model/warehouse_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class WarehouseRepository {
  Future<ApiResponse> getWarehouseManyByCode(List<String> codeList) async {
    Dio client = Client().init();
    try {
      final codeJson = jsonEncode(codeList);
      final response = await client.get('/warehouse/by-code?codes=$codeJson');
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

  Future<ApiResponse> getWarehouseList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/warehouse/list?offset=$offset&limit=$limit&q=$search&sort=warehousecode:1";
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

  Future<ApiResponse> deleteWarehouse(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/warehouse/$guid');
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
  Future<ApiResponse> deleteWarehouseMany(List<String> guidList) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guidList);
      final response = await client.delete('/warehouse', data: guidStrings);
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

  Future<ApiResponse> getWarehouse(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/warehouse/$guid');
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

  Future<ApiResponse> getWarehouseLocation(String warehousecode) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/warehouse/code/$warehousecode');
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

  Future<ApiResponse> getWarehouseByCode(String code) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/warehouse/code/$code');
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

  Future<ApiResponse> saveWarehouse(WarehouseModel warehouseModel) async {
    Dio client = Client().init();
    final data = warehouseModel.toJson();
    try {
      final response = await client.post('/warehouse', data: data);
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

  Future<ApiResponse> updateWarehouse(
      String guid, WarehouseModel warehouseModel) async {
    Dio client = Client().init();
    final data = warehouseModel.toJson();
    try {
      final response = await client.put('/warehouse/$guid', data: data);
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
