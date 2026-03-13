import 'dart:convert';

import 'package:smlaicloud/model/inventory_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class InventoryRepository {
  Future<ApiResponse> getInventoryList({
    int page = 0,
    int perPage = 1,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      final response = await client.get(
        '/inventory?page=$page&limit=$perPage&q=$search&sort=code:1',
      );
      try {
        final rawData = json.decode(response.toString());

        // print(rawData);

        if (rawData['error'] != null) {
          throw Exception('${rawData['code']}: ${rawData['message']}');
        }

        return ApiResponse.fromMap(rawData);
      } catch (ex) {
        // print(ex);
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      throw Exception(ex.response.toString());
    }
  }

  Future<ApiResponse> getInventoryId(String id) async {
    Dio client = Client().init();

    try {
      final response = await client.get('/inventory/$id');
      try {
        final rawData = json.decode(response.toString());

        // print(rawData);

        if (rawData['error'] != null) {
          throw Exception('${rawData['code']}: ${rawData['message']}');
        }

        return ApiResponse.fromMap(rawData);
      } catch (ex) {
        // print(ex);
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      throw Exception(ex.response.toString());
    }
  }

  Future<ApiResponse> saveInventory(InventoryModel inventory) async {
    Dio client = Client().init();
    final data = inventory.toJson();
    // print(data);
    try {
      final response = await client.post('/inventory', data: data);
      try {
        // final rawData = json.decode(response.toString());

        // print(response.data);

        // if (rawData['error'] != null) {
        //   String errorMessage = '${rawData['code']}: ${rawData['message']}';
        //   // print(errorMessage);
        //   throw new Exception('${rawData['code']}: ${rawData['message']}');
        // }

        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        // print(ex);
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      // print(ex);
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> updateInventory(InventoryModel inventory) async {
    Dio client = Client().init();
    final data = inventory.toJson();
    // print(data);
    try {
      final response = await client.put(
        '/inventory/${inventory.guidfixed}',
        data: data,
      );
      try {
        // final rawData = json.decode(response.toString());

        // print(response.data);

        // if (rawData['error'] != null) {
        //   String errorMessage = '${rawData['code']}: ${rawData['message']}';
        //   // print(errorMessage);
        //   throw new Exception('${rawData['code']}: ${rawData['message']}');
        // }

        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        // print(ex);
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      // print(ex);
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> deleteInventory(String id) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/inventory/$id');
      try {
        // final rawData = json.decode(response.toString());

        // print(response.data);

        // if (rawData['error'] != null) {
        //   String errorMessage = '${rawData['code']}: ${rawData['message']}';
        //   // print(errorMessage);
        //   throw new Exception('${rawData['code']}: ${rawData['message']}');
        // }

        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        // print(ex);
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      // print(ex);
      throw Exception(errorMessage);
    }
  }
}
