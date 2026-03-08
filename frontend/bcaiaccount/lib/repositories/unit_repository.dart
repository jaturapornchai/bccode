import 'dart:convert';

import 'package:smlaicloud/model/product_model.dart';
import 'package:flutter/foundation.dart';

import 'client.dart';
import 'package:dio/dio.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class UnitRepository {
  Future<ApiResponse> getUnitManyByCode(List<String> codeList) async {
    Dio client = Client().init();
    try {
      final codeJson = jsonEncode(codeList);
      final response = await client.get('/unit/by-code?codes=$codeJson');
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

  Future<ApiResponse> getUnitList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/unit/list?offset=$offset&limit=$limit&q=$search&sort=unitcode:1";
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

  Future<ApiResponse> deleteUnit(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/unit/$guid');
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
  Future<ApiResponse> deleteUnitMany(List<String> guidList) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guidList);
      final response = await client.delete('/unit', data: guidStrings);
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

  Future<ApiResponse> getUnit(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/unit/$guid');
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

  Future<ApiResponse> saveUnit(UnitModel unitModel) async {
    Dio client = Client().init();
    final data = unitModel.toJson();
    try {
      final response = await client.post('/unit', data: data);
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

  Future<ApiResponse> saveUnitBulk(List<UnitModel> units) async {
    Dio client = Client().init();
    try {
      final jsonList = units.map((units) => units.toJson()).toList();
      final jsonString = jsonEncode(jsonList);
      // print(jsonString);

      final response = await client.post('/unit/bulk', data: jsonString);
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

  Future<ApiResponse> updateUnit(String guid, UnitModel unitModel) async {
    Dio client = Client().init();
    final data = unitModel.toJson();
    try {
      final response = await client.put('/unit/$guid', data: data);
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

  Future<ApiResponse> getUnitListFromMainShop(String mainShopId) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/unit?limit=1000&shopsid=$mainShopId');
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

  Future<ApiResponse> uploadExcelFile(
    Uint8List fileBytes,
    String filename,
  ) async {
    Dio client = Client().init();
    try {
      FormData formData = FormData.fromMap({
        "excelfile": MultipartFile.fromBytes(fileBytes, filename: filename),
      });

      final response = await client.post('/unit/uploadfile', data: formData);

      if (kDebugMode) {
        AppLogger.debug("Upload response type: ${response.data.runtimeType}");
        if (kDebugMode) {
          AppLogger.debug("Upload response data: ${response.data}");
        }
      }

      try {
        if (response.data is String) {
          Map<String, dynamic> responseMap = {
            'success': true,
            'data': response.data,
            'message': 'Upload successful',
            'error': false,
          };
          return ApiResponse.fromMap(responseMap);
        } else {
          return ApiResponse.fromMap(response.data);
        }
      } catch (ex) {
        if (kDebugMode) {
          AppLogger.error("Error parsing response: $ex");
        }
        throw Exception("Failed to parse upload response: $ex");
      }
    } on DioException catch (ex) {
      String errorMessage =
          ex.response?.data?.toString() ?? ex.message.toString();
      throw Exception(errorMessage);
    }
  }
}
