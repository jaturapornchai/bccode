part of 'shelf_product_bloc.dart';

abstract class ShelfProductEvent extends Equatable {
  const ShelfProductEvent();

  @override
  List<Object> get props => [];
}

/// Event สำหรับเพิ่มสินค้าเข้าชั้นวาง
class AddProductsToShelf extends ShelfProductEvent {
  final String warehouseCode;
  final String locationCode;
  final String shelfCode;
  final ShelfProductBulkAddModel products;

  const AddProductsToShelf({
    required this.warehouseCode,
    required this.locationCode,
    required this.shelfCode,
    required this.products,
  });

  @override
  List<Object> get props => [warehouseCode, locationCode, shelfCode, products];
}

/// Event สำหรับลบสินค้าออกจากชั้นวาง
class RemoveProductsFromShelf extends ShelfProductEvent {
  final String warehouseCode;
  final String locationCode;
  final String shelfCode;
  final ShelfProductBulkDeleteModel productGuids;

  const RemoveProductsFromShelf({
    required this.warehouseCode,
    required this.locationCode,
    required this.shelfCode,
    required this.productGuids,
  });

  @override
  List<Object> get props =>
      [warehouseCode, locationCode, shelfCode, productGuids];
}

/// Event สำหรับดึงรายการสินค้าที่มีอยู่ในชั้นวาง
class LoadProductsInShelf extends ShelfProductEvent {
  final String warehouseCode;
  final String locationCode;
  final String shelfCode;
  final int limit;
  final int offset;
  final String search;

  const LoadProductsInShelf({
    required this.warehouseCode,
    required this.locationCode,
    required this.shelfCode,
    required this.limit,
    required this.offset,
    required this.search,
  });

  @override
  List<Object> get props =>
      [warehouseCode, locationCode, shelfCode, limit, offset, search];
}
