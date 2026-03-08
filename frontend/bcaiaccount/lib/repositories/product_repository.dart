import 'dart:convert';
import 'dart:io';
import 'package:flutter/services.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'client.dart';
import 'package:dio/dio.dart';

class ProductRepository {
  Future<ApiResponse> getProductList({int limit = 0, int offset = 0, String search = ""}) async {
    Dio client = Client().init();

    try {
      final response = await client.get('/product/list', queryParameters: {'offset': offset, 'limit': limit, 'q': search, 'sort': 'itemcode:1'});
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

  Future<ApiResponse> getProductByBarcode(String barcode) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/product/by-barcode?code=$barcode');
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

  Future<ApiResponse> saveProduct(ProductModel Product) async {
    Dio client = Client().init();
    final data = Product.toJson();
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

  Future<ApiResponse> updateProduct(String guid, ProductModel Product) async {
    Dio client = Client().init();
    final data = Product.toJson();
    try {
      final response = await client.put('/product/$guid', data: data);
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
    FormData formData = FormData.fromMap({"file": MultipartFile.fromBytes(image, filename: '$fileName.png')});
    try {
      final response = await client.post('/upload/images', data: formData);
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
}
