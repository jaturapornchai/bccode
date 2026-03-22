import 'package:dio/dio.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'client.dart';

/// ข้อมูลประวัติการซื้อสินค้า 1 รายการ
class PurchaseHistoryItem {
  final String barcode;
  final String docNo;
  final String docDate;
  final String custCode;
  final double totalQty;
  final String unitCode;
  final double price;
  final double sumAmount;
  final int transFlag;

  PurchaseHistoryItem({
    required this.barcode,
    required this.docNo,
    required this.docDate,
    required this.custCode,
    required this.totalQty,
    required this.unitCode,
    required this.price,
    required this.sumAmount,
    required this.transFlag,
  });

  factory PurchaseHistoryItem.fromJson(Map<String, dynamic> json) => PurchaseHistoryItem(
        barcode: json['barcode'] ?? '',
        docNo: json['doc_no'] ?? '',
        docDate: json['doc_date'] ?? '',
        custCode: json['cust_code'] ?? '',
        totalQty: (json['total_qty'] ?? 0).toDouble(),
        unitCode: json['unit_code'] ?? '',
        price: (json['price'] ?? 0).toDouble(),
        sumAmount: (json['sum_amount'] ?? 0).toDouble(),
        transFlag: json['trans_flag'] ?? 0,
      );
}

/// สถิติราคาต่อบาร์โค้ด
class PurchaseHistoryStats {
  final int count;
  final double avgPrice;
  final double minPrice;
  final double maxPrice;
  final double lastPrice;
  final String lastDate;
  final String lastVendor;

  PurchaseHistoryStats({
    required this.count,
    required this.avgPrice,
    required this.minPrice,
    required this.maxPrice,
    required this.lastPrice,
    required this.lastDate,
    required this.lastVendor,
  });

  factory PurchaseHistoryStats.fromJson(Map<String, dynamic> json) => PurchaseHistoryStats(
        count: json['count'] ?? 0,
        avgPrice: (json['avg_price'] ?? 0).toDouble(),
        minPrice: (json['min_price'] ?? 0).toDouble(),
        maxPrice: (json['max_price'] ?? 0).toDouble(),
        lastPrice: (json['last_price'] ?? 0).toDouble(),
        lastDate: json['last_date'] ?? '',
        lastVendor: json['last_vendor'] ?? '',
      );
}

/// Repository สำหรับดึงประวัติการซื้อสินค้า
class PurchaseHistoryRepository {
  /// ดึงประวัติซื้อสินค้าย้อนหลัง N เดือน
  Future<({List<PurchaseHistoryItem> items, PurchaseHistoryStats? stats})> getHistory(
    String barcode, {
    int months = 6,
  }) async {
    Dio client = Client().init();
    try {
      AppLogger.info('[PurchaseHistoryRepo] barcode=$barcode months=$months');
      final response = await client.post('/purchase-history', data: {
        'shop_id': global.getShopId(),
        'barcodes': [barcode],
        'months': months,
      });
      final raw = response.data;
      if (raw == null || raw['success'] != true) {
        return (items: <PurchaseHistoryItem>[], stats: null);
      }
      final dataList = raw['data'] as List? ?? [];
      final items = dataList.map((e) => PurchaseHistoryItem.fromJson(e as Map<String, dynamic>)).toList();
      PurchaseHistoryStats? stats;
      final statsMap = raw['stats'] as Map<String, dynamic>?;
      if (statsMap != null && statsMap.containsKey(barcode)) {
        stats = PurchaseHistoryStats.fromJson(statsMap[barcode] as Map<String, dynamic>);
      }
      return (items: items, stats: stats);
    } on DioException catch (ex) {
      AppLogger.error('[PurchaseHistoryRepo] error: ${ex.response?.data}');
      return (items: <PurchaseHistoryItem>[], stats: null);
    }
  }
}
