import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';
import 'package:smlaicloud/model/creditor_model.dart';
import 'package:flutter/foundation.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class CreditorRepository {
  Future<ApiResponse> getCreditorList({
    int limit = 0,
    int offset = 0,
    String search = "",
    List<String>? groups,
  }) async {
    Dio client = Client().init();
    String filterGroup = "";

    try {
      // ตรวจสอบ null ก่อนเข้าถึง groups
      if (groups != null && groups.isNotEmpty) {
        filterGroup = "&groups=${groups.join(',')}";
      }

      String query =
          "/debtaccount/creditor/list?offset=$offset&limit=$limit&q=$search$filterGroup";
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
      throw errorMessage;
    }
  }

  Future<ApiResponse> deleteCreditor(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/debtaccount/creditor/$guid');
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

  /// ลบที่ละหลาย GUID
  Future<ApiResponse> deleteCreditorMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guids);
      final response = await client.delete(
        '/debtaccount/creditor',
        data: guidStrings,
      );
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

  Future<ApiResponse> getCreditor(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/debtaccount/creditor/$guid');
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

  Future<ApiResponse> getCreditorBycode(String custcode) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/debtaccount/creditor/code/$custcode');
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

  Future<ApiResponse> saveCreditor(
    CreditorRequestModel creditorRequestModel,
  ) async {
    Dio client = Client().init();
    final data = creditorRequestModel.toJson();

    // print(data);
    try {
      final response = await client.post('/debtaccount/creditor', data: data);
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

  Future<ApiResponse> saveBulkCreditors(
    List<CreditorRequestModel> creditors,
  ) async {
    Dio client = Client().init();
    final data = creditors.map((creditor) => creditor.toJson()).toList();

    try {
      final response = await client.post(
        '/debtaccount/creditor/bulk',
        data: data,
      );
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

  Future<ApiResponse> updateCreditor(
    String guid,
    CreditorRequestModel creditorModel,
  ) async {
    Dio client = Client().init();
    final data = creditorModel.toJson();
    try {
      // print(data);
      final response = await client.put(
        '/debtaccount/creditor/$guid',
        data: data,
      );
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

  Future<ApiResponse> uploadImage(File file, Uint8List image) async {
    Dio client = Client().init();
    String fileName = file.path.split('/').last;
    FormData formData = FormData.fromMap({
      "file": MultipartFile.fromBytes(image, filename: '$fileName.png'),
    });
    try {
      final response = await client.post('/upload/images', data: formData);
      try {
        // print(response.data);
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        // print(ex);
        rethrow;
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      // print(ex);
      throw errorMessage;
    }
  }
}
