import 'package:flutter/foundation.dart';
import 'dart:convert';

import 'package:smlaicloud/global.dart';
import 'package:smlaicloud/model/product_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class ProductMasterRepository {
  Future<ApiResponse> getProductList({int limit = 0, int page = 0, String search = ""}) async {
    Dio client = Client().init();

    try {
      final response = await client.get('/product', queryParameters: {'page': getPageNumber(page, limit), 'limit': limit, 'q': search});
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

  Future<ApiResponse> deleteProduct(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/product/$guid');
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
  Future<ApiResponse> deleteProductMany(List<String> guidList) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guidList);
      final response = await client.delete('/product', data: guidStrings);
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

  Future<ApiResponse> getProduct(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/product/$guid');
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

  Future<ApiResponse> saveProduct(ProductMasterModel productModel) async {
    Dio client = Client().init();
    final data = productModel.toJson();
    try {
      final response = await client.post('/product', data: data);
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

  Future<ApiResponse> saveProductBulk(List<ProductMasterModel> products) async {
    Dio client = Client().init();
    try {
      final jsonList = products.map((products) => products.toJson()).toList();
      final jsonString = jsonEncode(jsonList);
      // print(jsonString);

      final response = await client.post('/product/bulk', data: jsonString);
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

  Future<ApiResponse> updateProduct(String guid, ProductMasterModel productModel) async {
    Dio client = Client().init();
    final data = productModel.toJson();
    if (kDebugMode) {
      AppLogger.debug(data);
    }
    try {
      final response = await client.put('/product/$guid', data: data);
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
