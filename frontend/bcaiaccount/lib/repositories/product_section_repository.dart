import 'package:smlaicloud/model/product_branch_model.dart';
import 'package:smlaicloud/model/product_department_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class ProductSectionRepository {
  /// ดึงสินค้าในสาขา
  Future<ApiResponse> getProductBranch(String branchcode) async {
    Dio client = Client().init();
    try {
      final response =
          await client.get('/product-section/branch/code/$branchcode');
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

  /// บันทึกสินค้าตามสาขา
  Future<ApiResponse> updateBarcodeInBranch(
      ProductBranchModel productBranchModel) async {
    Dio client = Client().init();
    final data = productBranchModel.toJson();
    try {
      final response = await client.put('/product-section/branch', data: data);
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

  /// ดึงสินค้าใน แผนก
  Future<ApiResponse> getProductDepartment(
      String branchcode, String departmentcode) async {
    Dio client = Client().init();
    try {
      final response = await client.get(
          '/product-section/department/$departmentcode/branch/$branchcode');
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

  /// บันทึกสินค้าตาม แผนก
  Future<ApiResponse> updateBarcodeInDepartment(
      ProductDepartmentModel productDepartmentModel) async {
    Dio client = Client().init();
    final data = productDepartmentModel.toJson();
    try {
      final response =
          await client.put('/product-section/department', data: data);
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
