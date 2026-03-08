import 'dart:typed_data';

import 'package:smlaicloud/model/stock_balance_import_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class StockBalanceImportRepository {
  Future<ApiResponse> uploadFileExcel(Uint8List file, String filename) async {
    Dio client = Client().init();

    FormData formData = FormData.fromMap({
      "file": MultipartFile.fromBytes(file, filename: filename),
    });

    try {
      final response =
          await client.post('/stockbalanceimport/upload', data: formData);
      try {
        // print(response.data);

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

  /// get stock balance import
  Future<ApiResponse> getStockBalanceImport(
      String taskid, String q, int limit, int page) async {
    Dio client = Client().init();

    try {
      final response = await client
          .get('/stockbalanceimport/$taskid?&limit=$limit&page=$page&q=$q');
      try {
        var data = response.data['data'];
        var pagination = response.data['pagination'];

        return ApiResponse.fromMap({
          'success': true,
          'data': data,
          'pagination': pagination,
        });
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  /// delete detail by guid
  Future<ApiResponse> deleteDetailByGuid(String guid) async {
    Dio client = Client().init();

    try {
      final response = await client.delete('/stockbalanceimport/item/$guid');
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

  /// delete taskid
  Future<ApiResponse> deleteTaskid(String taskid) async {
    Dio client = Client().init();

    try {
      final response = await client.delete('/stockbalanceimport/$taskid');
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

  /// update detail
  Future<ApiResponse> updateDetail(
      String guid, StockBalanceImportModel detail) async {
    Dio client = Client().init();
    final data = detail.toJson();
    try {
      final response =
          await client.put('/stockbalanceimport/item/$guid', data: data);
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

  /// add detail
  Future<ApiResponse> addDetail(StockBalanceImportModel detail) async {
    Dio client = Client().init();
    final data = detail.toJson();
    try {
      final response = await client.post('/stockbalanceimport', data: data);
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

  /// get total
  Future<ApiResponse> getTotal(String taskid) async {
    Dio client = Client().init();

    try {
      final response = await client.get('/stockbalanceimport/$taskid/meta');
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

  /// save
  Future<ApiResponse> saveTransStockBalance(
      String taskid, TransactionModel transactionModel) async {
    Dio client = Client().init();
    final data = transactionModel.toJson();
    try {
      final response =
          await client.post('/stockbalanceimport/$taskid', data: data);
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

  /// load trans stock balance by docno
  Future<ApiResponse> loadTransStockBalanceDetailByDocno(
      String docno, String q, int limit, int page) async {
    Dio client = Client().init();

    try {
      final response = await client.get(
          '/transaction/stock-balance-detail/doc/$docno?&limit=$limit&page=$page&q=$q');
      try {
        var data = response.data['data'];
        var pagination = response.data['pagination'];

        return ApiResponse.fromMap({
          'success': true,
          'data': data,
          'pagination': pagination,
        });
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }
}
