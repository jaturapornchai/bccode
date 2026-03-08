import 'dart:convert';
import 'package:smlaicloud/environment.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/transaction_model.dart';
import 'client.dart';
import 'package:dio/dio.dart';

class TransactionPaidPayRepository {
  Future<ApiResponse> getPaid({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/transaction/paid/list?offset=$offset&limit=$limit&q=$search";
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

  Future<ApiResponse> getPay({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/transaction/pay/list?offset=$offset&limit=$limit&q=$search";
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

  Future<ApiResponse> savePaid(TransactionPaidPayModel postData) async {
    Dio client = Client().init();
    final data = postData.toJson();
    try {
      final response = await client.post('/transaction/paid', data: data);
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

  Future<ApiResponse> savePay(TransactionPaidPayModel postData) async {
    Dio client = Client().init();
    final data = postData.toJson();
    try {
      final response = await client.post('/transaction/pay', data: data);
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

  Future<ApiResponse> updateTransPaid(
      String guid, TransactionPaidPayModel postData) async {
    Dio client = Client().init();
    final data = postData.toJson();
    try {
      final response = await client.put('/transaction/paid/$guid', data: data);
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

  Future<ApiResponse> updateTransPay(
      String guid, TransactionPaidPayModel postData) async {
    Dio client = Client().init();
    final data = postData.toJson();
    try {
      final response = await client.put('/transaction/pay/$guid', data: data);
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

  Future<ApiResponse> deletePaid(String guid) async {
    Dio client = Client().init();

    try {
      final response = await client.delete('/transaction/paid/$guid');
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

  Future<ApiResponse> deletePay(String guid) async {
    Dio client = Client().init();

    try {
      final response = await client.delete('/transaction/pay/$guid');
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

  /// Get Debtor Transaction
  Future<ApiResponse> getCustcodeTransaction(
      global.TransactionTypeEnum type, String creditorcode) async {
    Dio dio = Dio();
    final token = global.appConfig.getString("token");
    late Response<dynamic> response;
    try {
      if (type == global.TransactionTypeEnum.paid) {
        response = await dio.get(
            '${Environment().config.reportApi}/debtortransaction?token=$token&creditorcode=$creditorcode');
      } else {
        response = await dio.get(
            '${Environment().config.reportApi}/creditortransaction?token=$token&creditorcode=$creditorcode');
      }
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
