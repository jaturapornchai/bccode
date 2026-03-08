import 'dart:convert';
import 'package:smlaicloud/model/table_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class TableRepository {
  Future<ApiResponse> getTableList({
    int limit = 0,
    int offset = 0,
    String search = "",
    int groupNumber = 0,
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/restaurant/table?limit=$limit&offset=$offset&search=$search&group-number=$groupNumber";
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

  Future<ApiResponse> deleteTable(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/restaurant/table/$guid');
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
  Future<ApiResponse> deleteTableMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guids);
      final response =
          await client.delete('/restaurant/table', data: guidStrings);
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

  Future<ApiResponse> getTable(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/restaurant/table/$guid');
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

  Future<ApiResponse> saveTable(TableModel tableModel) async {
    Dio client = Client().init();
    final data = tableModel.toJson();
    try {
      final response = await client.post('/restaurant/table', data: data);
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

  Future<ApiResponse> updateTable(String guid, TableModel tableModel) async {
    Dio client = Client().init();
    final data = tableModel.toJson();
    try {
      final response = await client.put('/restaurant/table/$guid', data: data);
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

  Future<ApiResponse> updateTableXorder(
      List<TableXorderModel> tableModel) async {
    Dio client = Client().init();
    try {
      final data = jsonEncode(tableModel);
      final response = await client.put('/restaurant/table/xorder', data: data);
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
