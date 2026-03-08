part of 'product_barcode_bloc.dart';

abstract class ProductBarcodeEvent extends Equatable {
  const ProductBarcodeEvent();

  @override
  List<Object> get props => [];
}

class ProductBarcodeGet extends ProductBarcodeEvent {
  final String guid;

  const ProductBarcodeGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class ProductBarcodeGetRef extends ProductBarcodeEvent {
  final String guid;

  const ProductBarcodeGetRef({required this.guid});

  @override
  List<Object> get props => [guid];
}

class ProductBarcodeLoadList extends ProductBarcodeEvent {
  final int limit;
  final int offset;
  final String search;
  final String? itemtype;
  final String branchcode;
  final String businesstypecode;

  /// showbom | notshowbom | all
  final String? isbom;

  /// showsubbarcodes | notshowsubbarcodes | all
  final String? isusesubbarcodes;

  final String? shopsid;

  const ProductBarcodeLoadList({
    required this.offset,
    required this.limit,
    required this.search,
    this.itemtype = '',
    required this.branchcode,
    required this.businesstypecode,

    /// showbom | notshowbom | all
    this.isbom = '',

    /// showsubbarcodes | notshowsubbarcodes | all
    this.isusesubbarcodes = '',
    this.shopsid = '',
  });

  @override
  List<Object> get props => [
        offset,
        limit,
        search,
        itemtype!,
        branchcode,
        businesstypecode,
        shopsid!
      ];
}

class ProductBarcodeLoadListSearch extends ProductBarcodeEvent {
  final int limit;
  final int offset;
  final String search;
  final String? itemtype;
  final String branchcode;
  final String businesstypecode;

  /// showbom | notshowbom | all
  final String? isbom;

  /// showsubbarcodes | notshowsubbarcodes | all
  final String? isusesubbarcodes;

  final String? shopsid;

  const ProductBarcodeLoadListSearch({
    required this.offset,
    required this.limit,
    required this.search,
    this.itemtype = '',
    required this.branchcode,
    required this.businesstypecode,

    /// showbom | notshowbom | all
    this.isbom = '',

    /// showsubbarcodes | notshowsubbarcodes | all
    this.isusesubbarcodes = '',
    this.shopsid = '',
  });

  @override
  List<Object> get props => [
        offset,
        limit,
        search,
        itemtype!,
        branchcode,
        businesstypecode,
        shopsid!
      ];
}

class ProductBarcodeDelete extends ProductBarcodeEvent {
  final String guid;

  const ProductBarcodeDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ProductBarcodeDeleteMany extends ProductBarcodeEvent {
  final List<String> guid;

  const ProductBarcodeDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ProductBarcodeSave extends ProductBarcodeEvent {
  final ProductBarcodeModel productBarcode;

  const ProductBarcodeSave({
    required this.productBarcode,
  });

  @override
  List<Object> get props => [productBarcode];
}

class ProductBarcodeUpdate extends ProductBarcodeEvent {
  final String guid;
  final ProductBarcodeModel productBarcode;

  const ProductBarcodeUpdate({
    required this.guid,
    required this.productBarcode,
  });

  @override
  List<Object> get props => [productBarcode];
}

class ProductBarcodeWithImageSave extends ProductBarcodeEvent {
  final File imageFile;
  final ProductBarcodeModel productBarcode;
  final Uint8List? imageWeb;
  const ProductBarcodeWithImageSave({
    required this.imageWeb,
    required this.imageFile,
    required this.productBarcode,
  });

  @override
  List<Object> get props => [productBarcode, imageFile];
}

class ProductBarcodeWithImageUpdate extends ProductBarcodeEvent {
  final String guid;
  final ProductBarcodeModel productBarcode;
  final File imageFile;
  final Uint8List imageWeb;
  const ProductBarcodeWithImageUpdate({
    required this.guid,
    required this.imageFile,
    required this.imageWeb,
    required this.productBarcode,
  });

  @override
  List<Object> get props => [productBarcode, imageWeb, productBarcode];
}

class ProductBarcodeGetBom extends ProductBarcodeEvent {
  final String barcode;

  const ProductBarcodeGetBom({required this.barcode});

  @override
  List<Object> get props => [barcode];
}

class ProductBarcodeGetByBarcodeList extends ProductBarcodeEvent {
  final List<String> barcodes;

  const ProductBarcodeGetByBarcodeList({required this.barcodes});

  @override
  List<Object> get props => [barcodes];
}

class ProductBarcodeLoadListPg extends ProductBarcodeEvent {
  final int limit;
  final int offset;
  final String keyword;
  final String groupCode;
  final String brandCode;
  final String categoryCode;
  final String classCode;
  final String designCode;
  final String gradeCode;
  final String modelCode;
  final String patternCode;
  final double? priceMin;
  final double? priceMax;
  final String sortField;
  final String sortOrder;

  const ProductBarcodeLoadListPg({
    required this.offset,
    required this.limit,
    this.keyword = '',
    this.groupCode = '',
    this.brandCode = '',
    this.categoryCode = '',
    this.classCode = '',
    this.designCode = '',
    this.gradeCode = '',
    this.modelCode = '',
    this.patternCode = '',
    this.priceMin,
    this.priceMax,
    this.sortField = 'barcode',
    this.sortOrder = 'asc',
  });

  @override
  List<Object> get props => [
        offset,
        limit,
        keyword,
        groupCode,
        brandCode,
        categoryCode,
        sortField,
        sortOrder,
      ];
}

class ProductBarcodeGetByBarcode extends ProductBarcodeEvent {
  final String barcode;

  const ProductBarcodeGetByBarcode({required this.barcode});

  @override
  List<Object> get props => [barcode];
}

class ProductBarcodeGetPriceHistory extends ProductBarcodeEvent {
  final String barcode;
  final int page;
  final int limit;

  const ProductBarcodeGetPriceHistory({
    required this.barcode,
    this.page = 1,
    this.limit = 20,
  });

  @override
  List<Object> get props => [barcode, page, limit];
}
