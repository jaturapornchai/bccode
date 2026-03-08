import 'dart:convert';

import 'package:smlaicloud/model/company_branch_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class CompanyBranchRepository {
  Future<ApiResponse> getBranchList({
    int limit = 0,
    int offset = 0,
    String search = "",
    String? businesstypecode = "",
  }) async {
    Dio client = Client().init();

    if (businesstypecode!.isNotEmpty) {
      businesstypecode = "&businesstypecode=$businesstypecode";
    }

    try {
      String query =
          "/organization/branch/list?offset=$offset&limit=$limit&q=$search$businesstypecode";
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

  Future<ApiResponse> deleteBranch(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/organization/branch/$guid');
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
  Future<ApiResponse> deleteBranchMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guids);
      final response =
          await client.delete('/organization/branch', data: guidStrings);
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

  Future<ApiResponse> getBranch(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/organization/branch/$guid');
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

  Future<ApiResponse> getBranchBycode(String code) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/organization/branch/code/$code');
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

  Future<ApiResponse> saveBranch(CompanyBranchModel branchModel) async {
    Dio client = Client().init();
    final data = branchModel.toJson();
    try {
      final response = await client.post('/organization/branch', data: data);
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

  Future<ApiResponse> updateBranch(
      String guid, CompanyBranchModel branchModel) async {
    Dio client = Client().init();
    final data = branchModel.toJson();
    try {
      final response =
          await client.put('/organization/branch/$guid', data: data);
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
}
