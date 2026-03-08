import 'dart:convert';
import 'dart:io';
import 'package:smlaicloud/model/debtor_model.dart';
import 'package:smlaicloud/model/point_transaction_model.dart';
import 'package:smlaicloud/model/bulk_points_request_model.dart';
import 'package:flutter/foundation.dart';

import 'client.dart';
import 'package:dio/dio.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class DebtorRepository {
  Future<ApiResponse> getDebtorList({
    int limit = 0,
    int offset = 0,
    String search = "",
    List<String>? groups,
  }) async {
    Dio client = Client().init();
    String filterGroup = "";

    try {
      if (groups!.isNotEmpty) {
        filterGroup = "&groups=${groups.join(',')}";
      }
      String query =
          "/debtaccount/debtor/list?offset=$offset&limit=$limit&q=$search$filterGroup";
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

  Future<ApiResponse> deleteDebtor(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/debtaccount/debtor/$guid');
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
  Future<ApiResponse> deleteDebtorMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guids);
      final response = await client.delete(
        '/debtaccount/debtor',
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

  Future<ApiResponse> getDebtor(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/debtaccount/debtor/$guid');
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

  Future<ApiResponse> getDebtorByCode(String custcode) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/debtaccount/debtor/code/$custcode');
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

  Future<ApiResponse> saveDebtor(DebtorRequestModel debtorRequestModel) async {
    Dio client = Client().init();
    final data = debtorRequestModel.toJson();

    // print(data);
    try {
      final response = await client.post('/debtaccount/debtor', data: data);
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

  Future<ApiResponse> saveBulkDebtors(List<DebtorRequestModel> debtors) async {
    Dio client = Client().init();
    final data = debtors.map((debtor) => debtor.toJson()).toList();

    if (kDebugMode) {
      AppLogger.debug(
        '🌐 API POST /debtaccount/debtor/bulk with ${data.length} records:',
      );
      for (int i = 0; i < data.length; i++) {
        AppLogger.debug(
          '   [$i] code=${data[i]['code']}, pointscode=${data[i]['pointscode']}, groups=${data[i]['groups']}',
        );
      }
    }

    try {
      final response = await client.post(
        '/debtaccount/debtor/bulk',
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

  Future<ApiResponse> updateDebtor(
    String guid,
    DebtorRequestModel debtorModel,
  ) async {
    Dio client = Client().init();
    final data = debtorModel.toJson();
    try {
      // print(data);
      final response = await client.put(
        '/debtaccount/debtor/$guid',
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

  Future<ApiResponse> addPointToDebtor(
    String pointscode,
    PointTransactionCreateModel pointTransaction,
  ) async {
    Dio client = Client().init();
    final data = pointTransaction.toJson();

    if (kDebugMode) {
      AppLogger.debug(
        '🌐 API POST /debtaccount/debtor/pointscode/$pointscode/addpoint',
      );
      AppLogger.debug('   Data: $data');
    }

    try {
      final response = await client.post(
        '/debtaccount/debtor/pointscode/$pointscode/addpoint',
        data: data,
      );
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      if (kDebugMode) {
        AppLogger.error('❌ Error adding points: $errorMessage');
      }
      throw errorMessage;
    }
  }

  // ฟังก์ชันสำหรับ Bulk Add Points
  Future<BulkPointsResponseModel> bulkAddPoints(
    List<BulkPointsRequestModel> pointsList,
  ) async {
    Dio client = Client().init();
    final data = pointsList.map((p) => p.toJson()).toList();

    if (kDebugMode) {
      AppLogger.debug('🌐 API POST /debtaccount/debtor/points/bulk-add');
      AppLogger.debug('   Data: ${pointsList.length} items');
      AppLogger.debug('   First item: ${data.isNotEmpty ? data[0] : "empty"}');
    }

    try {
      final response = await client.post(
        '/debtaccount/debtor/points/bulk-add',
        data: data,
      );

      if (kDebugMode) {
        AppLogger.debug('✅ Bulk add points response: ${response.data}');
      }

      try {
        return BulkPointsResponseModel.fromJson(response.data);
      } catch (ex) {
        if (kDebugMode) {
          AppLogger.error('❌ Error parsing response: $ex');
        }
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage =
          ex.response?.toString() ?? ex.message ?? 'Unknown error';
      if (kDebugMode) {
        AppLogger.error('❌ Error bulk adding points: $errorMessage');
      }
      throw errorMessage;
    }
  }
}
