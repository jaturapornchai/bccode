import 'dart:convert';

import 'package:smlaicloud/model/journal_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class GlProcessRepository {
  Future<ApiResponse> getTransactionPurchaseList({
    required String fromDate,
    required String toDate,
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/transaction/purchase/list?fromdate=$fromDate&todate=$toDate";
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

  Future<ApiResponse> getTransactionSaleList({
    required String fromDate,
    required String toDate,
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/transaction/sale-invoice/list?fromdate=$fromDate&todate=$toDate";
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

  Future<ApiResponse> saveJournalBulk(List<JournalModel> journalData) async {
    Dio client = Client().init();

    final jsonList = journalData.map((barcode) => barcode.toJson()).toList();
    final jsonString = jsonEncode(jsonList);
    try {
      final response = await client.post('/gl/journal/bulk', data: jsonString);
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

  Future<ApiResponse> getTransactionPurchaseReturnList({
    required String fromDate,
    required String toDate,
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/transaction/purchase-return/list?fromdate=$fromDate&todate=$toDate";
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

  Future<ApiResponse> getTransactionSaleReturnList({
    required String fromDate,
    required String toDate,
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/transaction/sale-invoice-return/list?fromdate=$fromDate&todate=$toDate";
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

  Future<ApiResponse> getTransactionStockReturnProductList({
    required String fromDate,
    required String toDate,
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/transaction/stock-return-product/list?fromdate=$fromDate&todate=$toDate";
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

  Future<ApiResponse> getTransactionStockAdjustList({
    required String fromDate,
    required String toDate,
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/transaction/stock-adjustment/list?fromdate=$fromDate&todate=$toDate";
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

  Future<ApiResponse> getTransactionStockPickupList({
    required String fromDate,
    required String toDate,
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/transaction/stock-prickup-product/list?fromdate=$fromDate&todate=$toDate";
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

  Future<ApiResponse> getTransactionStockReceiveList({
    required String fromDate,
    required String toDate,
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/transaction/stock-receive-product/list?fromdate=$fromDate&todate=$toDate";
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

  Future<ApiResponse> getTransactionPaidList({
    required String fromDate,
    required String toDate,
  }) async {
    Dio client = Client().init();

    try {
      String query = "/transaction/paid/list?fromdate=$fromDate&todate=$toDate";
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

  Future<ApiResponse> getTransactionPayList({
    required String fromDate,
    required String toDate,
  }) async {
    Dio client = Client().init();

    try {
      String query = "/transaction/pay/list?fromdate=$fromDate&todate=$toDate";
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
}
