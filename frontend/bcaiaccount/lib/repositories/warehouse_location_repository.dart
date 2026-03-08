import 'dart:convert';

import 'package:smlaicloud/model/warehouse_location_update_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class WarehouseLocationRepository {
  /// GET : /warehouse/location
  Future<ApiResponse> getWarehouseLocationList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/warehouse/location?offset=$offset&limit=$limit&q=$search&sort=warehousecode:1";
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

  /// GET : /warehouse/{warehouseCode}/location/{locationCode}
  Future<ApiResponse> getWarehouseLocationByCode(
      String warehousecode, String locationcode) async {
    Dio client = Client().init();
    try {
      final response =
          await client.get('/warehouse/$warehousecode/location/$locationcode');
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

  /// PUT : /warehouse/{warehouseCode}/location/{locationCode}
  /// body : WarehouseLocationModel

  Future<ApiResponse> updateWarehouseLocation(
      String warehousecode,
      String locationcode,
      WarehouseLocationUpdateModel warehouseLocationUpdateModel) async {
    Dio client = Client().init();
    try {
      final response = await client.put(
          '/warehouse/$warehousecode/location/$locationcode',
          data: warehouseLocationUpdateModel.toJson());
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

  ///  delete location
  Future<ApiResponse> deleteWarehouseLocationMany(
      String warehousecode, List<String> locationcode) async {
    Dio client = Client().init();
    try {
      final locationcodeList = jsonEncode(locationcode);
      final response = await client.delete('/warehouse/$warehousecode/location',
          data: locationcodeList);
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
