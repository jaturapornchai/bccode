import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'client.dart';
import 'package:dio/dio.dart';

class PointTransactionRepository {
  Future<ApiResponse> getPointTransactionsByDebtorCode(
    String debtorCode,
  ) async {
    Dio client = Client().init();

    try {
      String query = "/debtaccount/debtor/code/$debtorCode/pointtransactions";
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

  Future<ApiResponse> deleteManualPointTransaction(
    String pointscode,
    String docno,
  ) async {
    Dio client = Client().init();

    if (kDebugMode) {
      AppLogger.debug(
        '🌐 API DELETE /debtaccount/debtor/pointscode/$pointscode/manual-transaction/$docno',
      );
    }

    try {
      final response = await client.delete(
        '/debtaccount/debtor/pointscode/$pointscode/manual-transaction/$docno',
      );
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      if (kDebugMode) {
        AppLogger.error(
          '❌ Error deleting manual point transaction: $errorMessage',
        );
      }
      throw errorMessage;
    }
  }
}
