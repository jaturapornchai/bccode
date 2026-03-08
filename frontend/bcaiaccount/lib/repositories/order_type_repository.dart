import 'dart:convert';
import 'package:smlaicloud/model/order_type_model.dart';
import 'client.dart';
import 'package:dio/dio.dart';

class OrderTypeRepository {
  Future<ApiResponse> getOrderTypeByCode(String code) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/product/order-type/code/$code');
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

  Future<ApiResponse> getOrderTypeList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/product/order-type/list?offset=$offset&limit=$limit&q=$search&sort=code:1";
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

  Future<ApiResponse> deleteOrderType(String guid) async {
    Dio client = Client().init();
    // print(guid);
    try {
      final response = await client.delete('/product/order-type/$guid');
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
  Future<ApiResponse> deleteOrderTypeMany(List<String> guidList) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guidList);
      final response =
          await client.delete('/product/order-type', data: guidStrings);
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

  Future<ApiResponse> getOrderType(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/product/order-type/$guid');
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

  Future<ApiResponse> saveOrderType(OrderTypeModel ordertypemodel) async {
    Dio client = Client().init();
    final data = ordertypemodel.toJson();
    try {
      final response = await client.post('/product/order-type', data: data);
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

  Future<ApiResponse> updateOrderType(
      String guid, OrderTypeModel ordertypemodel) async {
    Dio client = Client().init();
    final data = ordertypemodel.toJson();
    try {
      final response =
          await client.put('/product/order-type/$guid', data: data);
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
