import 'dart:convert';
import 'package:smlaicloud/model/shelf_product_model.dart';
import 'client.dart';
import 'package:dio/dio.dart';

class ShelfProductRepository {
  /// POST : /warehouse/{warehouseCode}/location/{locationCode}/shelf/{shelfCode}/products/bulk
  /// เพิ่มสินค้าเข้าชั้นวาง
  Future<ApiResponse> addProductsToShelf({
    required String warehouseCode,
    required String locationCode,
    required String shelfCode,
    required ShelfProductBulkAddModel products,
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/warehouse/$warehouseCode/location/$locationCode/shelf/$shelfCode/products/bulk";
      final response = await client.post(query, data: products.toJson());

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

  /// DELETE : /warehouse/{warehouseCode}/location/{locationCode}/shelf/{shelfCode}/products/bulk
  /// ลบสินค้าออกจากชั้นวาง
  Future<ApiResponse> removeProductsFromShelf({
    required String warehouseCode,
    required String locationCode,
    required String shelfCode,
    required ShelfProductBulkDeleteModel productGuids,
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/warehouse/$warehouseCode/location/$locationCode/shelf/$shelfCode/products/bulk";
      final response = await client.delete(query, data: productGuids.toJson());

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

  /// GET : /warehouse/{warehouseCode}/location/{locationCode}/shelf/{shelfCode}/products
  /// ดึงรายการสินค้าที่มีอยู่ในชั้นวางแล้ว (สำหรับแสดงผล)
  Future<ApiResponse> getProductsInShelf({
    required String warehouseCode,
    required String locationCode,
    required String shelfCode,
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/warehouse/$warehouseCode/location/$locationCode/shelf/$shelfCode/products?offset=$offset&limit=$limit&q=$search";
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

  /// GET : /warehouse/{warehouseCode}/location/{locationCode}/shelf/{shelfCode}/products (Count only)
  /// ดึงจำนวนสินค้าที่มีอยู่ในชั้นวาง (สำหรับแสดงจำนวน)
  Future<int> getProductCountInShelf({
    required String warehouseCode,
    required String locationCode,
    required String shelfCode,
  }) async {
    Dio client = Client().init();

    try {
      // ใช้ limit=1000 เพื่อให้ได้ข้อมูลครบทั้งหมด (หรือจำนวนที่คาดหวัง)
      String query =
          "/warehouse/$warehouseCode/location/$locationCode/shelf/$shelfCode/products?offset=0&limit=1000&q=";
      final response = await client.get(query);

      try {
        final rawData = json.decode(response.toString());
        if (rawData['error'] != null || rawData['success'] != true) {
          return 0; // Return 0 if there's an error or no products
        }

        final apiResponse = ApiResponse.fromMap(rawData);

        // ถ้ามี pagination ใช้ total จาก pagination
        if (apiResponse.page != null && apiResponse.page!.total > 0) {
          return apiResponse.page!.total;
        }

        // ถ้าไม่มี pagination นับจาก data array
        if (apiResponse.data is List) {
          return (apiResponse.data as List).length;
        }

        return 0;
      } catch (ex) {
        return 0;
      }
    } catch (ex) {
      return 0;
    }
  }
}
