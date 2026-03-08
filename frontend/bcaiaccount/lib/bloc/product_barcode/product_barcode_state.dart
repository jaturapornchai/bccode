part of 'product_barcode_bloc.dart';

abstract class ProductBarcodeState extends Equatable {
  const ProductBarcodeState();

  @override
  List<Object> get props => [];
}

class ProductBarcodeInitial extends ProductBarcodeState {}

class ProductBarcodeInProgress extends ProductBarcodeState {}

class ProductBarcodeLoadSuccess extends ProductBarcodeState {
  final List<ProductBarcodeModel> productBarcodes;

  const ProductBarcodeLoadSuccess({required this.productBarcodes});

  ProductBarcodeLoadSuccess copyWith({
    List<ProductBarcodeModel>? productBarcodes,
  }) => ProductBarcodeLoadSuccess(
    productBarcodes: productBarcodes ?? this.productBarcodes,
  );

  @override
  List<Object> get props => [productBarcodes];
}

class ProductBarcodeLoadFailed extends ProductBarcodeState {
  final String message;

  const ProductBarcodeLoadFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class ProductBarcodeSearchInProgress extends ProductBarcodeState {}

class ProductBarcodeLoadSearchSuccess extends ProductBarcodeState {
  final List<ProductBarcodeModel> productBarcodes;

  const ProductBarcodeLoadSearchSuccess({required this.productBarcodes});

  ProductBarcodeLoadSearchSuccess copyWith({
    List<ProductBarcodeModel>? productBarcodes,
  }) => ProductBarcodeLoadSearchSuccess(
    productBarcodes: productBarcodes ?? this.productBarcodes,
  );

  @override
  List<Object> get props => [productBarcodes];
}

class ProductBarcodeLoadSearchFailed extends ProductBarcodeState {
  final String message;

  const ProductBarcodeLoadSearchFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class ProductBarcodeSaveInitial extends ProductBarcodeState {}

class ProductBarcodeSaveInProgress extends ProductBarcodeState {}

class ProductBarcodeSaveSuccess extends ProductBarcodeState {}

class ProductBarcodeSaveFailed extends ProductBarcodeState {
  final String message;

  const ProductBarcodeSaveFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class ProductBarcodeDeleteInProgress extends ProductBarcodeState {}

class ProductBarcodeDeleteSuccess extends ProductBarcodeState {}

class ProductBarcodeDeleteFailed extends ProductBarcodeState {
  Null get message => null;
}

class ProductBarcodeDeleteManyInProgress extends ProductBarcodeState {}

class ProductBarcodeDeleteManySuccess extends ProductBarcodeState {
  Null get message => null;
}

class ProductBarcodeDeleteManyFailed extends ProductBarcodeState {
  final String message;

  const ProductBarcodeDeleteManyFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class ProductBarcodeGetInProgress extends ProductBarcodeState {}

class ProductBarcodeGetSuccess extends ProductBarcodeState {
  final ProductBarcodeModel productBarcode;

  const ProductBarcodeGetSuccess({required this.productBarcode});

  ProductBarcodeGetSuccess copyWith({ProductBarcodeModel? productBarcode}) =>
      ProductBarcodeGetSuccess(
        productBarcode: productBarcode ?? this.productBarcode,
      );

  @override
  List<Object> get props => [productBarcode];
}

class ProductBarcodeGetFailed extends ProductBarcodeState {
  final String message;

  const ProductBarcodeGetFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class ProductBarcodeUpdateInitial extends ProductBarcodeState {}

class ProductBarcodeUpdateInProgress extends ProductBarcodeState {}

class ProductBarcodeUpdateSuccess extends ProductBarcodeState {}

class ProductBarcodeUpdateFailed extends ProductBarcodeState {
  final String message;

  const ProductBarcodeUpdateFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class ProductBarcodeGetRefInProgress extends ProductBarcodeState {}

class ProductBarcodeGetRefSuccess extends ProductBarcodeState {
  final List<ProductBarcodeModel> productBarcodes;

  const ProductBarcodeGetRefSuccess({required this.productBarcodes});

  ProductBarcodeGetRefSuccess copyWith({
    List<ProductBarcodeModel>? productBarcodes,
  }) => ProductBarcodeGetRefSuccess(
    productBarcodes: productBarcodes ?? this.productBarcodes,
  );

  @override
  List<Object> get props => [productBarcodes];
}

class ProductBarcodeGetRefFailed extends ProductBarcodeState {
  final String message;

  const ProductBarcodeGetRefFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class ProductBarcodeGetBomInProgress extends ProductBarcodeState {}

class ProductBarcodeGetBomSuccess extends ProductBarcodeState {
  final ProductBomModel productBom;

  const ProductBarcodeGetBomSuccess({required this.productBom});

  ProductBarcodeGetBomSuccess copyWith({ProductBomModel? productBom}) =>
      ProductBarcodeGetBomSuccess(productBom: productBom ?? this.productBom);

  @override
  List<Object> get props => [productBom];
}

class ProductBarcodeGetBomFailed extends ProductBarcodeState {
  final String message;

  const ProductBarcodeGetBomFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class ProductBarcodeGetByBarcodeListInProgress extends ProductBarcodeState {}

class ProductBarcodeGetByBarcodeListSuccess extends ProductBarcodeState {
  final List<ProductBarcodeModel> productBarcodes;

  const ProductBarcodeGetByBarcodeListSuccess({required this.productBarcodes});

  ProductBarcodeGetByBarcodeListSuccess copyWith({
    List<ProductBarcodeModel>? productBarcodes,
  }) => ProductBarcodeGetByBarcodeListSuccess(
    productBarcodes: productBarcodes ?? this.productBarcodes,
  );

  @override
  List<Object> get props => [productBarcodes];
}

class ProductBarcodeGetByBarcodeListFailed extends ProductBarcodeState {
  final String message;

  const ProductBarcodeGetByBarcodeListFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class ProductBarcodeLoadPgInProgress extends ProductBarcodeState {}

class ProductBarcodeLoadPgSuccess extends ProductBarcodeState {
  final List<ProductBarcodeModel> productBarcodes;
  final int total;

  const ProductBarcodeLoadPgSuccess({
    required this.productBarcodes,
    required this.total,
  });

  @override
  List<Object> get props => [productBarcodes, total];
}

class ProductBarcodeLoadPgFailed extends ProductBarcodeState {
  final String message;

  const ProductBarcodeLoadPgFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class ProductBarcodeGetPriceHistoryInProgress extends ProductBarcodeState {}

class ProductBarcodeGetPriceHistorySuccess extends ProductBarcodeState {
  final List<PriceHistoryModel> priceHistory;
  final PaginationModel? pagination;

  const ProductBarcodeGetPriceHistorySuccess({
    required this.priceHistory,
    this.pagination,
  });

  ProductBarcodeGetPriceHistorySuccess copyWith({
    List<PriceHistoryModel>? priceHistory,
    PaginationModel? pagination,
  }) => ProductBarcodeGetPriceHistorySuccess(
    priceHistory: priceHistory ?? this.priceHistory,
    pagination: pagination ?? this.pagination,
  );

  @override
  List<Object> get props => [priceHistory];
}

class ProductBarcodeGetPriceHistoryFailed extends ProductBarcodeState {
  final String message;

  const ProductBarcodeGetPriceHistoryFailed({required this.message});

  @override
  List<Object> get props => [message];
}
