import 'package:dio/dio.dart';
import 'package:smlaicloud/global.dart' as global;
import 'client.dart';

/// Repository สำหรับ Inventory Costing API
/// Endpoints อยู่ที่ GoAPI: /goapi/api/inventory/...
class InventoryCostingRepository {
  String _goApiUrl(String endpoint) => global.goApiUrlPath('api/$endpoint');

  // === Costing Config ===

  Future<Map<String, dynamic>> getCostingConfig(String itemCode) async {
    Dio client = Client().init();
    try {
      final url = _goApiUrl('products/$itemCode/costing-config');
      final response = await client.get(url,
          queryParameters: {'shopid': global.getShopId()});
      return response.data as Map<String, dynamic>;
    } on DioException catch (ex) {
      throw Exception(ex.response?.toString() ?? ex.message);
    }
  }

  Future<Map<String, dynamic>> updateCostingConfig(
      String itemCode, Map<String, dynamic> config) async {
    Dio client = Client().init();
    try {
      final url = _goApiUrl('products/$itemCode/costing-config');
      final response = await client.put(url,
          data: config,
          queryParameters: {'shopid': global.getShopId()});
      return response.data as Map<String, dynamic>;
    } on DioException catch (ex) {
      throw Exception(ex.response?.toString() ?? ex.message);
    }
  }

  // === Cost Layers ===

  Future<List<dynamic>> getCostLayers(String itemCode,
      {String whCode = ''}) async {
    Dio client = Client().init();
    try {
      final url = _goApiUrl('products/$itemCode/cost-layers');
      final params = <String, dynamic>{'shopid': global.getShopId()};
      if (whCode.isNotEmpty) params['whcode'] = whCode;
      final response = await client.get(url, queryParameters: params);
      return response.data as List<dynamic>;
    } on DioException catch (ex) {
      throw Exception(ex.response?.toString() ?? ex.message);
    }
  }

  // === Transactions ===

  Future<Map<String, dynamic>> processReceipt(
      Map<String, dynamic> params) async {
    Dio client = Client().init();
    try {
      final url = _goApiUrl('inventory/receipt');
      final response = await client.post(url, data: params);
      return response.data as Map<String, dynamic>;
    } on DioException catch (ex) {
      throw Exception(ex.response?.toString() ?? ex.message);
    }
  }

  Future<Map<String, dynamic>> processIssue(
      Map<String, dynamic> params) async {
    Dio client = Client().init();
    try {
      final url = _goApiUrl('inventory/issue');
      final response = await client.post(url, data: params);
      return response.data as Map<String, dynamic>;
    } on DioException catch (ex) {
      throw Exception(ex.response?.toString() ?? ex.message);
    }
  }

  Future<Map<String, dynamic>> processTransfer(
      Map<String, dynamic> params) async {
    Dio client = Client().init();
    try {
      final url = _goApiUrl('inventory/transfer');
      final response = await client.post(url, data: params);
      return response.data as Map<String, dynamic>;
    } on DioException catch (ex) {
      throw Exception(ex.response?.toString() ?? ex.message);
    }
  }

  Future<Map<String, dynamic>> processAdjustment(
      Map<String, dynamic> params) async {
    Dio client = Client().init();
    try {
      final url = _goApiUrl('inventory/adjustment');
      final response = await client.post(url, data: params);
      return response.data as Map<String, dynamic>;
    } on DioException catch (ex) {
      throw Exception(ex.response?.toString() ?? ex.message);
    }
  }

  Future<Map<String, dynamic>> processSalesReturn(
      Map<String, dynamic> params) async {
    Dio client = Client().init();
    try {
      final url = _goApiUrl('inventory/sales-return');
      final response = await client.post(url, data: params);
      return response.data as Map<String, dynamic>;
    } on DioException catch (ex) {
      throw Exception(ex.response?.toString() ?? ex.message);
    }
  }

  Future<Map<String, dynamic>> processPurchaseReturn(
      Map<String, dynamic> params) async {
    Dio client = Client().init();
    try {
      final url = _goApiUrl('inventory/purchase-return');
      final response = await client.post(url, data: params);
      return response.data as Map<String, dynamic>;
    } on DioException catch (ex) {
      throw Exception(ex.response?.toString() ?? ex.message);
    }
  }

  // === Reports ===

  Future<Map<String, dynamic>> getInventoryValuation() async {
    Dio client = Client().init();
    try {
      final url = _goApiUrl('reports/inventory-valuation');
      final response = await client.get(url,
          queryParameters: {'shopid': global.getShopId()});
      return response.data as Map<String, dynamic>;
    } on DioException catch (ex) {
      throw Exception(ex.response?.toString() ?? ex.message);
    }
  }

  Future<Map<String, dynamic>> getStockCard(String itemCode,
      {String? fromDate, String? toDate}) async {
    Dio client = Client().init();
    try {
      final url = _goApiUrl('reports/stock-card/$itemCode');
      final params = <String, dynamic>{'shopid': global.getShopId()};
      if (fromDate != null) params['from'] = fromDate;
      if (toDate != null) params['to'] = toDate;
      final response = await client.get(url, queryParameters: params);
      return response.data as Map<String, dynamic>;
    } on DioException catch (ex) {
      throw Exception(ex.response?.toString() ?? ex.message);
    }
  }

  // === Database ===

  Future<Map<String, dynamic>> createTables() async {
    Dio client = Client().init();
    try {
      final url = _goApiUrl('inventory/create-tables');
      final response = await client.post(url);
      return response.data as Map<String, dynamic>;
    } on DioException catch (ex) {
      throw Exception(ex.response?.toString() ?? ex.message);
    }
  }
}
