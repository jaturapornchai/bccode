part of 'product_bloc.dart';

abstract class ProductState extends Equatable {
  const ProductState();

  @override
  List<Object> get props => [];
}

class ProductInitial extends ProductState {}

class ProductInProgress extends ProductState {}

class ProductLoadSuccess extends ProductState {
  final List<ProductModel> products;

  const ProductLoadSuccess({required this.products});

  ProductLoadSuccess copyWith({
    List<ProductModel>? products,
  }) =>
      ProductLoadSuccess(products: products ?? this.products);

  @override
  List<Object> get props => [products];
}

class ProductLoadFailed extends ProductState {
  final String message;

  const ProductLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductSaveInitial extends ProductState {}

class ProductSaveInProgress extends ProductState {}

class ProductSaveSuccess extends ProductState {}

class ProductSaveFailed extends ProductState {
  final String message;

  const ProductSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductDeleteInProgress extends ProductState {}

class ProductDeleteSuccess extends ProductState {}

class ProductDeleteFailed extends ProductState {}

class ProductDeleteManyInProgress extends ProductState {}

class ProductDeleteManySuccess extends ProductState {}

class ProductDeleteManyFailed extends ProductState {
  final String message;

  const ProductDeleteManyFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductGetInProgress extends ProductState {}

class ProductGetSuccess extends ProductState {
  final ProductModel product;

  const ProductGetSuccess({required this.product});

  ProductGetSuccess copyWith({
    ProductModel? product,
  }) =>
      ProductGetSuccess(product: product ?? this.product);

  @override
  List<Object> get props => [product];
}

class ProductGetFailed extends ProductState {
  final String message;

  const ProductGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductUpdateInitial extends ProductState {}

class ProductUpdateInProgress extends ProductState {}

class ProductUpdateSuccess extends ProductState {}

class ProductUpdateFailed extends ProductState {
  final String message;

  const ProductUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductBranchGetInProgress extends ProductState {}

class ProductBranchGetSuccess extends ProductState {
  final List<ProductBarcodeModel> products;

  const ProductBranchGetSuccess({required this.products});

  ProductBranchGetSuccess copyWith({
    List<ProductBarcodeModel>? products,
  }) =>
      ProductBranchGetSuccess(products: products ?? this.products);

  @override
  List<Object> get props => [products];
}

class ProductBranchGetFailed extends ProductState {
  final String message;

  const ProductBranchGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductBranchupdateInProgress extends ProductState {}

class ProductBranchUpdateSuccess extends ProductState {}

class ProductBranchUpdateFailed extends ProductState {
  final String message;

  const ProductBranchUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductDepartmentGetInProgress extends ProductState {}

class ProductDepartmentGetSuccess extends ProductState {
  final List<ProductBarcodeModel> products;

  const ProductDepartmentGetSuccess({required this.products});

  ProductDepartmentGetSuccess copyWith({
    List<ProductBarcodeModel>? products,
  }) =>
      ProductDepartmentGetSuccess(products: products ?? this.products);

  @override
  List<Object> get props => [products];
}

class ProductDepartmentGetFailed extends ProductState {
  final String message;

  const ProductDepartmentGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductDepartmentupdateInProgress extends ProductState {}

class ProductDepartmentUpdateSuccess extends ProductState {}

class ProductDepartmentUpdateFailed extends ProductState {
  final String message;

  const ProductDepartmentUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductGetByBarcodesInProgress extends ProductState {
  @override
  List<Object> get props => [];
}

class ProductGetByBarcodesSuccess extends ProductState {
  final List<ProductBarcodeModel> products;
  final String source;

  const ProductGetByBarcodesSuccess({required this.products, this.source = ''});

  @override
  List<Object> get props => [products, source];
}

class ProductGetByBarcodesFailed extends ProductState {
  final String message;

  const ProductGetByBarcodesFailed({required this.message});

  @override
  List<Object> get props => [message];
}
