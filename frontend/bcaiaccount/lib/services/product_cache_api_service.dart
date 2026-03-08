import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:smlaicloud/global.dart' as global;

/// Product Cache API Service
/// ใช้ Backend centralized cache สำหรับข้อมูลสินค้า
class ProductCacheApiService {
  static String get _baseUrl => global.goApiBaseUrl;

  /// สร้าง headers สำหรับ API (รวม Authorization token)
  static Map<String, String> get _headers {
    final headers = <String, String>{'Content-Type': 'application/json'};
    final token = global.appConfig.getString("token") ?? '';
    if (token.isNotEmpty) {
      headers['Authorization'] = 'Bearer $token';
    }
    return headers;
  }

  /// ค้นหาสินค้าพร้อม cache
  ///
  /// [search] - คำค้นหา (itemcode, barcode, หรือชื่อสินค้า)
  /// [branchCode] - รหัสสาขา (optional)
  /// [businessTypeCode] - รหัสประเภทธุรกิจ (optional)
  /// [limit] - จำนวน record สูงสุด
  /// [offset] - เริ่มต้นจาก record ที่
  /// [useCache] - ใช้ cache หรือไม่ (default true)
  static Future<ProductSearchResult> searchProducts({
    String search = "",
    String branchCode = "",
    String businessTypeCode = "",
    int limit = 50,
    int offset = 0,
    bool useCache = true,
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/product/search");

      final body = {
        "shop_id": global.getShopId(),
        "search": search,
        "branch_code": branchCode,
        "business_type_code": businessTypeCode,
        "limit": limit,
        "offset": offset,
        "use_cache": useCache,
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          return ProductSearchResult.fromJson(json);
        }
      }

      return ProductSearchResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      return ProductSearchResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ค้นหาสินค้าด้วย barcode (optimized single lookup)
  ///
  /// [barcode] - barcode ของสินค้า
  static Future<ProductBarcodeResult> searchByBarcode({
    required String barcode,
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/product/barcode");

      final body = {
        "shop_id": global.getShopId(),
        "barcode": barcode,
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          return ProductBarcodeResult.fromJson(json);
        }
      }

      if (response.statusCode == 404) {
        return ProductBarcodeResult.notFound();
      }

      return ProductBarcodeResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      return ProductBarcodeResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ดึงสถิติ cache
  static Future<CacheStatsResult> getCacheStats() async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/product/cache/stats");

      final response = await httpClient.get(
        url,
        headers: _headers,
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['status'] == 'success') {
          return CacheStatsResult.fromJson(json['data']);
        }
      }

      return CacheStatsResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      return CacheStatsResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ล้าง cache
  ///
  /// [prefix] - prefix ของ key ที่ต้องการล้าง (optional, ถ้าไม่ระบุจะล้างทั้งหมด)
  static Future<bool> clearCache({String? prefix}) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/product/cache/clear");

      final Map<String, dynamic> body = {
        "shop_id": global.getShopId(),
      };

      if (prefix != null) {
        body["prefix"] = prefix;
      }

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      return response.statusCode == 200;
    } catch (e) {
      return false;
    } finally {
      httpClient.close();
    }
  }
}

/// ผลลัพธ์การค้นหาสินค้า
class ProductSearchResult {
  final bool isSuccess;
  final String? errorMessage;
  final List<ProductData> products;
  final int count;
  final bool cached;
  final int limit;
  final int offset;

  ProductSearchResult({
    required this.isSuccess,
    this.errorMessage,
    this.products = const [],
    this.count = 0,
    this.cached = false,
    this.limit = 0,
    this.offset = 0,
  });

  factory ProductSearchResult.fromJson(Map<String, dynamic> json) {
    final dataList = (json['data'] as List?)
            ?.map((item) => ProductData.fromJson(item))
            .toList() ??
        [];

    return ProductSearchResult(
      isSuccess: true,
      products: dataList,
      count: json['count'] ?? 0,
      cached: json['cached'] ?? false,
      limit: json['limit'] ?? 0,
      offset: json['offset'] ?? 0,
    );
  }

  factory ProductSearchResult.error(String message) => ProductSearchResult(
        isSuccess: false,
        errorMessage: message,
      );
}

/// ผลลัพธ์การค้นหาด้วย barcode
class ProductBarcodeResult {
  final bool isSuccess;
  final bool isFound;
  final String? errorMessage;
  final ProductData? product;
  final bool cached;

  ProductBarcodeResult({
    required this.isSuccess,
    this.isFound = false,
    this.errorMessage,
    this.product,
    this.cached = false,
  });

  factory ProductBarcodeResult.fromJson(Map<String, dynamic> json) {
    return ProductBarcodeResult(
      isSuccess: true,
      isFound: true,
      product: ProductData.fromJson(json['data']),
      cached: json['cached'] ?? false,
    );
  }

  factory ProductBarcodeResult.notFound() => ProductBarcodeResult(
        isSuccess: true,
        isFound: false,
      );

  factory ProductBarcodeResult.error(String message) => ProductBarcodeResult(
        isSuccess: false,
        errorMessage: message,
      );
}

/// ข้อมูลสินค้า
class ProductData {
  final String itemCode;
  final String barcode;
  final String itemName;
  final String unitCode;
  final String unitName;
  final double price;
  final double unitStand;
  final double unitDivide;
  final String? categoryCode;
  final int? vatType;
  final double? costPrice;

  ProductData({
    required this.itemCode,
    required this.barcode,
    required this.itemName,
    required this.unitCode,
    required this.unitName,
    required this.price,
    required this.unitStand,
    required this.unitDivide,
    this.categoryCode,
    this.vatType,
    this.costPrice,
  });

  factory ProductData.fromJson(Map<String, dynamic> json) => ProductData(
        itemCode: json['itemcode'] ?? '',
        barcode: json['barcode'] ?? '',
        itemName: json['itemname'] ?? '',
        unitCode: json['unitcode'] ?? '',
        unitName: json['unitname'] ?? '',
        price: (json['price'] ?? 0).toDouble(),
        unitStand: (json['unitstand'] ?? 1).toDouble(),
        unitDivide: (json['unitdivide'] ?? 1).toDouble(),
        categoryCode: json['categorycode'],
        vatType: json['vattype'],
        costPrice: json['costprice']?.toDouble(),
      );

  Map<String, dynamic> toJson() => {
        'itemcode': itemCode,
        'barcode': barcode,
        'itemname': itemName,
        'unitcode': unitCode,
        'unitname': unitName,
        'price': price,
        'unitstand': unitStand,
        'unitdivide': unitDivide,
        'categorycode': categoryCode,
        'vattype': vatType,
        'costprice': costPrice,
      };
}

/// ผลลัพธ์สถิติ cache
class CacheStatsResult {
  final bool isSuccess;
  final String? errorMessage;
  final int size;
  final int maxSize;
  final int ttlSeconds;
  final int totalHits;

  CacheStatsResult({
    required this.isSuccess,
    this.errorMessage,
    this.size = 0,
    this.maxSize = 0,
    this.ttlSeconds = 0,
    this.totalHits = 0,
  });

  factory CacheStatsResult.fromJson(Map<String, dynamic> json) =>
      CacheStatsResult(
        isSuccess: true,
        size: json['size'] ?? 0,
        maxSize: json['max_size'] ?? 0,
        ttlSeconds: json['ttl_seconds'] ?? 0,
        totalHits: json['total_hits'] ?? 0,
      );

  factory CacheStatsResult.error(String message) => CacheStatsResult(
        isSuccess: false,
        errorMessage: message,
      );
}
