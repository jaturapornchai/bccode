import 'dart:convert';
import 'package:smlaicloud/model/bank_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class BankRepository {
  Future<ApiResponse> getBankList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/payment/bankmaster/list?offset=$offset&limit=$limit&q=$search";
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

  Future<ApiResponse> deleteBank(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/payment/bankmaster/$guid');
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
  Future<ApiResponse> deleteBankMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guids);
      final response =
          await client.delete('/payment/bankmaster', data: guidStrings);
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

  Future<ApiResponse> getBank(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/payment/bankmaster/$guid');
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

  Future<ApiResponse> saveBank(BankModel bankModel) async {
    Dio client = Client().init();
    final data = bankModel.toJson();
    try {
      final response = await client.post('/payment/bankmaster', data: data);
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

  Future<ApiResponse> saveBankBulk(List<BankModel> banks) async {
    Dio client = Client().init();
    try {
      final jsonList = banks.map((banks) => banks.toJson()).toList();
      final jsonString = jsonEncode(jsonList);
      // print(jsonString);

      final response =
          await client.post('/payment/bankmaster/bulk', data: jsonString);
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

  Future<ApiResponse> updateBank(String guid, BankModel bankModel) async {
    Dio client = Client().init();
    final data = bankModel.toJson();
    try {
      final response =
          await client.put('/payment/bankmaster/$guid', data: data);
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
