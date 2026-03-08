import 'dart:convert';
import 'package:smlaicloud/model/product_type_model.dart';
import 'client.dart';
import 'package:dio/dio.dart';

class ProductTypeRepository {
  Future<ApiResponse> getProductTypeByCode(String code) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/product/type/code/$code');
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

  Future<ApiResponse> getProductTypeList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/product/type/list?offset=$offset&limit=$limit&q=$search&sort=code:1";
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

  Future<ApiResponse> deleteProductType(String guid) async {
    Dio client = Client().init();
    // print(guid);
    try {
      final response = await client.delete('/product/type/$guid');
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
  Future<ApiResponse> deleteProductTypeMany(List<String> guidList) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guidList);
      final response = await client.delete('/product/type', data: guidStrings);
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

  Future<ApiResponse> getProductType(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/product/type/$guid');
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

  Future<ApiResponse> saveProductType(ProductTypeModel productTypemodel) async {
    Dio client = Client().init();
    final data = productTypemodel.toJson();
    try {
      final response = await client.post('/product/type', data: data);
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

  Future<ApiResponse> updateProductType(
      String guid, ProductTypeModel productTypemodel) async {
    Dio client = Client().init();
    final data = productTypemodel.toJson();
    try {
      final response = await client.put('/product/type/$guid', data: data);
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
