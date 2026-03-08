import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';
import 'package:smlaicloud/model/customer_request_model.dart';
import 'package:flutter/foundation.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class CustomerRepository {
  Future<ApiResponse> getCustomerList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/debtaccount/customer/list?offset=$offset&limit=$limit&q=$search";
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

  Future<ApiResponse> getSupplierList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/debtaccount/customer/list?offset=$offset&limit=$limit&q=$search&iscreditor=true";
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

  Future<ApiResponse> deleteCustomer(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/debtaccount/customer/$guid');
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
  Future<ApiResponse> deleteCustomerMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guids);
      final response = await client.delete(
        '/debtaccount/customer',
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

  Future<ApiResponse> getCustomer(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/debtaccount/customer/$guid');
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

  Future<ApiResponse> getCustomerByCode(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/debtaccount/customer/code/$guid');
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

  Future<ApiResponse> getSupplierByCode(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get(
        '/debtaccount/customer/code/$guid?iscreditor=true',
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

  Future<ApiResponse> saveCustomer(
    CustomerRequestModel customerRequestModel,
  ) async {
    Dio client = Client().init();
    final data = customerRequestModel.toJson();

    // print(data);
    try {
      final response = await client.post('/debtaccount/customer', data: data);
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

  Future<ApiResponse> updateCustomer(
    String guid,
    CustomerRequestModel customerRequestModel,
  ) async {
    Dio client = Client().init();
    final data = customerRequestModel.toJson();
    try {
      // print(data);
      final response = await client.put(
        '/debtaccount/customer/$guid',
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
