import 'dart:convert';
import 'package:smlaicloud/model/printer_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class PrinterRepository {
  Future<ApiResponse> getPrinterList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/restaurant/printer/list?offset=$offset&limit=$limit&q=$search";
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

  Future<ApiResponse> deletePrinter(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/restaurant/printer/$guid');
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
  Future<ApiResponse> deletePrinterMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/restaurant/printer', data: guids);
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

  Future<ApiResponse> getPrinter(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/restaurant/printer/$guid');
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

  Future<ApiResponse> savePrinter(PrinterModel printerModel) async {
    Dio client = Client().init();
    final data = printerModel.toJson();
    try {
      final response = await client.post('/restaurant/printer', data: data);
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

  Future<ApiResponse> updatePrinter(
      String guid, PrinterModel printerModel) async {
    Dio client = Client().init();
    final data = printerModel.toJson();
    try {
      final response =
          await client.put('/restaurant/printer/$guid', data: data);
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
