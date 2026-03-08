import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';
import 'package:flutter/foundation.dart';
import 'package:smlaicloud/model/product_category_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'client.dart';
import 'package:dio/dio.dart';

class ProductCategoryRepository {
  Future<ApiResponse> getCategoryList({
    int limit = 0,
    int offset = 0,
    String search = "",
    int groupNumber = 0,
  }) async {
    Dio client = Client().init();

    String findGroupNumber = "";
    String findSearch = "";

    // ignore: unnecessary_null_comparison
    if (groupNumber != 0) {
      findGroupNumber = "&group-number=$groupNumber";
    }

    if (search.isNotEmpty) {
      findSearch = "&q=$search";
    }

    try {
      String query =
          "/product/category/list?offset=$offset&limit=$limit$findSearch$findGroupNumber";
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

  Future<ApiResponse> deleteCategory(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/product/category/$guid');
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
  Future<ApiResponse> deleteCategoryMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/product/category', data: guids);
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

  Future<ApiResponse> getCategory(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/product/category/$guid');
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

  Future<ApiResponse> uploadImage(File file, Uint8List image) async {
    Dio client = Client().init();
    String fileName = file.path.split('/').last;
    FormData formData = FormData.fromMap({
      "file": MultipartFile.fromBytes(image, filename: '$fileName.png'),
    });
    try {
      final response = await client.post('/upload/images', data: formData);
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      // print(ex);
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> saveCategory(ProductCategoryModel categoryModel) async {
    Dio client = Client().init();
    final data = categoryModel.toJson();
    try {
      final response = await client.post('/product/category', data: data);
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

  Future<ApiResponse> updateCategory(
      String guid, ProductCategoryModel categoryModel) async {
    Dio client = Client().init();
    final data = categoryModel.toJson();
    try {
      final response = await client.put('/product/category/$guid', data: data);
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

  Future<ApiResponse> updateCategoryXOrder(List<XSortModel> list) async {
    Dio client = Client().init();
    final data = jsonEncode(list);
    try {
      final response = await client.put('/product/category/xsort', data: data);
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
