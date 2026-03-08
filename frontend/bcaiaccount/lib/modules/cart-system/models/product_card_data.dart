import 'product_search_model.dart';

/// Model สำหรับเก็บข้อมูลที่ใช้แสดงใน Product Card
/// รวมข้อมูล product หลัก, units ทั้งหมด, และยอดคงเหลือไว้ด้วยกัน
class ProductCardData {
  /// สินค้าหลักที่จะแสดง
  final ProductSearchModel product;

  /// รายการ units ทั้งหมดของ itemcode นี้ (barcode ต่างๆ)
  final List<ProductSearchModel> allUnits;

  /// ยอดคงเหลือรวมของ itemcode นี้
  final double totalBalance;

  /// ยอดคงเหลือแยกตาม shop (shopId -> balance)
  final Map<String, double> balanceByShop;

  /// ยอดคงเหลือแบบ formatted string จาก Backend (เช่น "1 กล่อง x 2 โหล x 3 ชิ้น")
  final String? balanceFormatted;

  /// ยอดคงเหลือแยกตาม warehouse/location จาก Backend API
  final List<WarehouseBalanceModel> warehouses;

  /// ยอดค้างรับ (PO ที่ยังไม่ได้รับของ)
  final double pendingRecvQty;
  final String pendingRecvWord;

  /// ยอดค้างส่ง (SO ที่ยังไม่ได้ส่งของ)
  final double pendingSendQty;
  final String pendingSendWord;

  const ProductCardData({
    required this.product,
    required this.allUnits,
    required this.totalBalance,
    required this.balanceByShop,
    this.balanceFormatted,
    this.warehouses = const [],
    this.pendingRecvQty = 0.0,
    this.pendingRecvWord = '',
    this.pendingSendQty = 0.0,
    this.pendingSendWord = '',
  });

  /// สร้าง ProductCardData จาก product, units, และ balance model
  factory ProductCardData.fromData({
    required ProductSearchModel product,
    required List<ProductSearchModel> allUnits,
    ProductBalanceModel? balance,
    String? balanceFormatted,
    List<WarehouseBalanceModel>? warehouses,
  }) {
    return ProductCardData(
      product: product,
      allUnits: allUnits,
      totalBalance: balance?.totalBalance ?? 0.0,
      balanceByShop: balance?.balanceByShop ?? {},
      balanceFormatted: balanceFormatted,
      warehouses: warehouses ?? balance?.warehouses ?? const [],
      pendingRecvQty: balance?.pendingRecvQty ?? 0.0,
      pendingRecvWord: balance?.pendingRecvWord ?? '',
      pendingSendQty: balance?.pendingSendQty ?? 0.0,
      pendingSendWord: balance?.pendingSendWord ?? '',
    );
  }

  /// สร้าง ProductCardData แบบไม่มียอดคงเหลือ
  factory ProductCardData.withoutBalance({
    required ProductSearchModel product,
    required List<ProductSearchModel> allUnits,
  }) {
    return ProductCardData(
      product: product,
      allUnits: allUnits,
      totalBalance: 0.0,
      balanceByShop: {},
      warehouses: const [],
    );
  }

  /// ตรวจสอบว่ามียอดคงเหลือหรือไม่
  bool get hasBalance => totalBalance > 0;

  /// ตรวจสอบว่ามี formatted balance จาก Backend หรือไม่
  bool get hasFormattedBalance => balanceFormatted != null && balanceFormatted!.isNotEmpty;

  /// ตรวจสอบว่ามีข้อมูล warehouse/location หรือไม่
  bool get hasWarehouses => warehouses.isNotEmpty;

  /// จำนวน warehouses ทั้งหมด
  int get warehouseCount => warehouses.length;

  /// จำนวน units ทั้งหมด
  int get unitCount => allUnits.length;

  /// มียอดค้างรับหรือไม่
  bool get hasPendingRecv => pendingRecvQty > 0;

  /// มียอดค้างส่งหรือไม่
  bool get hasPendingSend => pendingSendQty > 0;

  /// มียอดค้างรับหรือค้างส่งหรือไม่
  bool get hasPending => hasPendingRecv || hasPendingSend;

  /// itemCode ของสินค้า
  String get itemCode => product.itemCode;

  /// barcode ของสินค้าหลัก
  String get barcode => product.barcode;
}
