import 'package:equatable/equatable.dart';
import '../models/product_search_model.dart';

/// Base class สำหรับ Product Search States
abstract class ProductSearchState extends Equatable {
  const ProductSearchState();

  @override
  List<Object?> get props => [];
}

/// สถานะเริ่มต้น
class ProductSearchInitial extends ProductSearchState {}

/// กำลังค้นหา
class ProductSearchLoading extends ProductSearchState {}

/// ค้นหาสำเร็จ
class ProductSearchLoaded extends ProductSearchState {
  final List<ProductSearchModel> products;
  final String searchKeyword;

  /// units ทั้งหมดของแต่ละ itemcode (จาก Backend Unified API)
  /// key = itemcode, value = List<ProductSearchModel>
  final Map<String, List<ProductSearchModel>>? unitsMap;

  /// ยอดคงเหลือของแต่ละ itemcode (จาก Backend Unified API)
  /// key = itemcode, value = ProductBalanceModel
  final Map<String, ProductBalanceModel>? balancesMap;

  /// ยอดคงเหลือแบบ formatted string จาก Backend (เช่น "1 กล่อง x 2 โหล x 3 ชิ้น")
  /// key = itemcode, value = formatted string
  final Map<String, String>? balanceFormattedMap;

  // Pagination info สำหรับ infinite scroll
  final int total; // จำนวนทั้งหมดที่พบ
  final bool hasMore; // มีข้อมูลเพิ่มหรือไม่
  final int currentOffset; // offset ปัจจุบัน

  const ProductSearchLoaded({
    required this.products,
    this.searchKeyword = '',
    this.unitsMap,
    this.balancesMap,
    this.balanceFormattedMap,
    this.total = 0,
    this.hasMore = false,
    this.currentOffset = 0,
  });

  @override
  List<Object?> get props => [products, searchKeyword, unitsMap, balancesMap, balanceFormattedMap, total, hasMore, currentOffset];

  // Helper methods
  int get productCount => products.length;
  bool get isEmpty => products.isEmpty;
  bool get isNotEmpty => products.isNotEmpty;

  /// ตรวจสอบว่ามีข้อมูลจาก Unified API หรือไม่
  bool get hasUnifiedData => unitsMap != null && balancesMap != null;

  ProductSearchLoaded copyWith({
    List<ProductSearchModel>? products,
    String? searchKeyword,
    Map<String, List<ProductSearchModel>>? unitsMap,
    Map<String, ProductBalanceModel>? balancesMap,
    Map<String, String>? balanceFormattedMap,
    int? total,
    bool? hasMore,
    int? currentOffset,
  }) {
    return ProductSearchLoaded(
      products: products ?? this.products,
      searchKeyword: searchKeyword ?? this.searchKeyword,
      unitsMap: unitsMap ?? this.unitsMap,
      balancesMap: balancesMap ?? this.balancesMap,
      balanceFormattedMap: balanceFormattedMap ?? this.balanceFormattedMap,
      total: total ?? this.total,
      hasMore: hasMore ?? this.hasMore,
      currentOffset: currentOffset ?? this.currentOffset,
    );
  }
}

/// กำลังโหลดเพิ่ม (infinite scroll)
class ProductSearchLoadingMore extends ProductSearchState {
  final List<ProductSearchModel> currentProducts;
  final String searchKeyword;
  final Map<String, List<ProductSearchModel>>? unitsMap;
  final Map<String, ProductBalanceModel>? balancesMap;
  final Map<String, String>? balanceFormattedMap;
  final int total;
  final int currentOffset;

  const ProductSearchLoadingMore({
    required this.currentProducts,
    required this.searchKeyword,
    this.unitsMap,
    this.balancesMap,
    this.balanceFormattedMap,
    this.total = 0,
    this.currentOffset = 0,
  });

  @override
  List<Object?> get props => [currentProducts, searchKeyword, total, currentOffset];
}

/// พบสินค้า 1 รายการจาก barcode
class ProductFoundByBarcode extends ProductSearchState {
  final ProductSearchModel product;

  const ProductFoundByBarcode({required this.product});

  @override
  List<Object?> get props => [product];
}

/// ไม่พบสินค้า
class ProductNotFound extends ProductSearchState {
  final String searchKeyword;

  const ProductNotFound({required this.searchKeyword});

  @override
  List<Object?> get props => [searchKeyword];
}

/// เกิดข้อผิดพลาด
class ProductSearchError extends ProductSearchState {
  final String error;

  const ProductSearchError({required this.error});

  @override
  List<Object?> get props => [error];
}
