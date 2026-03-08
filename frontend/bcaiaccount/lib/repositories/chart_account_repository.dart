import 'dart:convert';

import 'client.dart';
import 'package:dio/dio.dart';

class ChartAccountRepository {
  Future<ApiResponse> getChartAccount({
    int limit = 1000,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query = "/gl/chartofaccount?limit=$limit&q=$search";
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

  /// ชุดบัญชี
  Future<ApiResponse> getAccountGroup({
    int limit = 1000,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query = "/gl/accountgroup?limit=$limit&q=$search";
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

  /// สมุดบัญชี
  Future<ApiResponse> getAccountBook({
    int limit = 1000,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query = "/gl/journalbook?limit=$limit&q=$search";
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
}
