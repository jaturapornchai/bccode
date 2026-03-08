import 'package:dio/dio.dart';
import 'package:smlaicloud/api/client.dart';
import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class CurrencyApiService {
  final Dio _dio = Client().init();

  // ============ Currency CRUD ============

  /// Get all currencies
  /// [shopId] - Shop ID (required)
  /// [disabled] - Filter by disabled status (optional)
  /// [q] - Search query (optional)
  /// [page] - Page number (optional, default 1)
  /// [limit] - Items per page (optional, default 50)
  Future<ApiResponse> getCurrencies({
    required String shopId,
    bool? disabled,
    String? q,
    int? page,
    int? limit,
  }) async {
    try {
      final queryParams = <String, dynamic>{};
      if (disabled != null) queryParams['disabled'] = disabled;
      if (q != null && q.isNotEmpty) queryParams['q'] = q;
      if (page != null) queryParams['page'] = page;
      if (limit != null) queryParams['limit'] = limit;

      final response = await _dio.get(
        'currency',
        queryParameters: queryParams,
      );

      AppLogger.info('Get currencies response: ${response.data}');
      return ApiResponse.fromMap(response.data);
    } catch (e) {
      AppLogger.error('Failed to get currencies: $e');
      rethrow;
    }
  }

  /// Get currency list (for dropdown/selector)
  /// [shopId] - Shop ID (required)
  /// [disabled] - Filter by disabled status (optional)
  /// [q] - Search query (optional)
  /// [offset] - Offset (optional)
  /// [limit] - Items limit (optional)
  Future<ApiResponse> getCurrencyList({
    required String shopId,
    bool? disabled,
    String? q,
    int? offset,
    int? limit,
  }) async {
    try {
      final queryParams = <String, dynamic>{};
      if (disabled != null) queryParams['disabled'] = disabled;
      if (q != null && q.isNotEmpty) queryParams['q'] = q;
      if (offset != null) queryParams['offset'] = offset;
      if (limit != null) queryParams['limit'] = limit;

      final response = await _dio.get(
        'currency/list',
        queryParameters: queryParams,
      );

      AppLogger.info('Get currency list response: total=${response.data['total']}');
      return ApiResponse.fromMap(response.data);
    } catch (e) {
      AppLogger.error('Failed to get currency list: $e');
      rethrow;
    }
  }

  /// Get currency by guid
  /// [guidfixed] - Currency GUID (required)
  Future<CurrencyModel?> getCurrency(String guidfixed) async {
    try {
      final response = await _dio.get('currency/$guidfixed');

      AppLogger.info('Get currency response: ${response.data}');
      final apiResp = ApiResponse.fromMap(response.data);
      if (apiResp.success && apiResp.data != null) {
        return CurrencyModel.fromJson(apiResp.data);
      }
      return null;
    } catch (e) {
      AppLogger.error('Failed to get currency: $e');
      rethrow;
    }
  }

  /// Create currency
  /// [currency] - Currency model (required)
  Future<ApiResponse> createCurrency(CurrencyModel currency) async {
    try {
      final response = await _dio.post(
        'currency',
        data: currency.toJson(),
      );

      AppLogger.info('Create currency response: ${response.data}');
      return ApiResponse.fromMap(response.data);
    } catch (e) {
      AppLogger.error('Failed to create currency: $e');
      rethrow;
    }
  }

  /// Update currency
  /// [guidfixed] - Currency GUID (required)
  /// [currency] - Currency model (required)
  Future<ApiResponse> updateCurrency(String guidfixed, CurrencyModel currency) async {
    try {
      final response = await _dio.put(
        'currency/$guidfixed',
        data: currency.toJson(),
      );

      AppLogger.info('Update currency response: ${response.data}');
      return ApiResponse.fromMap(response.data);
    } catch (e) {
      AppLogger.error('Failed to update currency: $e');
      rethrow;
    }
  }

  /// Delete currency
  /// [guidfixed] - Currency GUID (required)
  Future<ApiResponse> deleteCurrency(String guidfixed) async {
    try {
      final response = await _dio.delete('currency/$guidfixed');

      AppLogger.info('Delete currency response: ${response.data}');
      return ApiResponse.fromMap(response.data);
    } catch (e) {
      AppLogger.error('Failed to delete currency: $e');
      rethrow;
    }
  }

  /// Delete multiple currencies
  /// [guidfixeds] - List of currency GUIDs (required)
  Future<ApiResponse> deleteCurrencies(List<String> guidfixeds) async {
    try {
      final response = await _dio.delete(
        'currency',
        data: guidfixeds,
      );

      AppLogger.info('Delete currencies response: ${response.data}');
      return ApiResponse.fromMap(response.data);
    } catch (e) {
      AppLogger.error('Failed to delete currencies: $e');
      rethrow;
    }
  }

  // ============ Exchange Rate History CRUD ============

  /// Get exchange rate history
  /// [shopId] - Shop ID (required)
  /// [currency] - Filter by currency code (optional)
  /// [q] - Search query (optional)
  /// [page] - Page number (optional, default 1)
  /// [limit] - Items per page (optional, default 50)
  Future<ApiResponse> getExchangeRateHistory({
    required String shopId,
    String? currency,
    String? q,
    int? page,
    int? limit,
  }) async {
    try {
      final queryParams = <String, dynamic>{};
      if (currency != null && currency.isNotEmpty) queryParams['currency'] = currency;
      if (q != null && q.isNotEmpty) queryParams['q'] = q;
      if (page != null) queryParams['page'] = page;
      if (limit != null) queryParams['limit'] = limit;

      final response = await _dio.get(
        'exchange-rate-history',
        queryParameters: queryParams,
      );

      AppLogger.info('Get exchange rate history response: ${response.data}');
      return ApiResponse.fromMap(response.data);
    } catch (e) {
      AppLogger.error('Failed to get exchange rate history: $e');
      rethrow;
    }
  }

  /// Get exchange rate history list (for dropdown/selector)
  /// [shopId] - Shop ID (required)
  /// [currency] - Filter by currency code (optional)
  /// [q] - Search query (optional)
  /// [offset] - Offset (optional)
  /// [limit] - Items limit (optional)
  Future<ApiResponse> getExchangeRateHistoryList({
    required String shopId,
    String? currency,
    String? q,
    int? offset,
    int? limit,
  }) async {
    try {
      final queryParams = <String, dynamic>{};
      if (currency != null && currency.isNotEmpty) queryParams['currency'] = currency;
      if (q != null && q.isNotEmpty) queryParams['q'] = q;
      if (offset != null) queryParams['offset'] = offset;
      if (limit != null) queryParams['limit'] = limit;

      final response = await _dio.get(
        'exchange-rate-history/list',
        queryParameters: queryParams,
      );

      AppLogger.info('Get exchange rate history list response: total=${response.data['total']}');
      return ApiResponse.fromMap(response.data);
    } catch (e) {
      AppLogger.error('Failed to get exchange rate history list: $e');
      rethrow;
    }
  }

  /// Get exchange rate by guid
  /// [guidfixed] - Exchange rate GUID (required)
  Future<ExchangeRateHistoryModel?> getExchangeRate(String guidfixed) async {
    try {
      final response = await _dio.get('exchange-rate-history/$guidfixed');

      AppLogger.info('Get exchange rate response: ${response.data}');
      final apiResp = ApiResponse.fromMap(response.data);
      if (apiResp.success && apiResp.data != null) {
        return ExchangeRateHistoryModel.fromJson(apiResp.data);
      }
      return null;
    } catch (e) {
      AppLogger.error('Failed to get exchange rate: $e');
      rethrow;
    }
  }

  /// Get latest exchange rate for a currency on or before a specific date
  /// [currency] - Currency code (e.g. USD, EUR, JPY) (required)
  /// [date] - Date in YYYY-MM-DD format (required)
  Future<ExchangeRateHistoryModel?> getLatestExchangeRate({
    required String currency,
    required String date,
  }) async {
    try {
      final response = await _dio.get(
        'exchange-rate-history/latest',
        queryParameters: {
          'currency': currency,
          'date': date,
        },
      );

      AppLogger.info('Get latest exchange rate response: ${response.data}');
      final apiResp = ApiResponse.fromMap(response.data);
      if (apiResp.success && apiResp.data != null) {
        return ExchangeRateHistoryModel.fromJson(apiResp.data);
      }
      return null;
    } catch (e) {
      AppLogger.error('Failed to get latest exchange rate: $e');
      return null; // Return null instead of rethrow for graceful handling
    }
  }

  /// Create exchange rate history
  /// [exchangeRate] - Exchange rate model (required)
  Future<ApiResponse> createExchangeRate(ExchangeRateHistoryModel exchangeRate) async {
    try {
      final requestBody = exchangeRate.toJson();
      AppLogger.info('=== CREATE EXCHANGE RATE ===');
      AppLogger.info('Currency: ${exchangeRate.currency}');
      AppLogger.info('Date: ${exchangeRate.date}');
      AppLogger.info('Rate: ${exchangeRate.rate}');
      AppLogger.info('Request JSON: $requestBody');

      final response = await _dio.post(
        'exchange-rate-history',
        data: requestBody,
      );

      AppLogger.info('Create exchange rate response: ${response.data}');
      return ApiResponse.fromMap(response.data);
    } on DioException catch (e) {
      // Log detailed error information
      AppLogger.error('=== CREATE EXCHANGE RATE FAILED ===');
      AppLogger.error('Error: ${e.message}');
      if (e.response != null) {
        AppLogger.error('Status: ${e.response?.statusCode}');
        AppLogger.error('Response body: ${e.response?.data}');
      }
      rethrow;
    } catch (e) {
      AppLogger.error('Failed to create exchange rate: $e');
      rethrow;
    }
  }

  /// Update exchange rate history
  /// [guidfixed] - Exchange rate GUID (required)
  /// [exchangeRate] - Exchange rate model (required)
  Future<ApiResponse> updateExchangeRate(String guidfixed, ExchangeRateHistoryModel exchangeRate) async {
    try {
      final response = await _dio.put(
        'exchange-rate-history/$guidfixed',
        data: exchangeRate.toJson(),
      );

      AppLogger.info('Update exchange rate response: ${response.data}');
      return ApiResponse.fromMap(response.data);
    } catch (e) {
      AppLogger.error('Failed to update exchange rate: $e');
      rethrow;
    }
  }

  /// Delete exchange rate history
  /// [guidfixed] - Exchange rate GUID (required)
  Future<ApiResponse> deleteExchangeRate(String guidfixed) async {
    try {
      final response = await _dio.delete('exchange-rate-history/$guidfixed');

      AppLogger.info('Delete exchange rate response: ${response.data}');
      return ApiResponse.fromMap(response.data);
    } catch (e) {
      AppLogger.error('Failed to delete exchange rate: $e');
      rethrow;
    }
  }

  /// Delete multiple exchange rate history entries
  /// [guidfixeds] - List of exchange rate GUIDs (required)
  Future<ApiResponse> deleteExchangeRates(List<String> guidfixeds) async {
    try {
      final response = await _dio.delete(
        'exchange-rate-history',
        data: guidfixeds,
      );

      AppLogger.info('Delete exchange rates response: ${response.data}');
      return ApiResponse.fromMap(response.data);
    } catch (e) {
      AppLogger.error('Failed to delete exchange rates: $e');
      rethrow;
    }
  }
}
