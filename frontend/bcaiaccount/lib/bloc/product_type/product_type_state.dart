part of 'product_type_bloc.dart';

abstract class ProductTypeState extends Equatable {
  const ProductTypeState();

  @override
  List<Object> get props => [];
}

class ProductTypeInitial extends ProductTypeState {}

class ProductTypeInProgress extends ProductTypeState {}

class ProductTypeLoadSuccess extends ProductTypeState {
  final List<ProductTypeModel> productType;

  const ProductTypeLoadSuccess({required this.productType});

  ProductTypeLoadSuccess copyWith({
    List<ProductTypeModel>? productType,
  }) =>
      ProductTypeLoadSuccess(productType: productType ?? this.productType);

  @override
  List<Object> get props => [productType];
}

class ProductTypeLoadFailed extends ProductTypeState {
  final String message;

  const ProductTypeLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductTypeSaveInitial extends ProductTypeState {}

class ProductTypeSaveInProgress extends ProductTypeState {}

class ProductTypeSaveSuccess extends ProductTypeState {}

class ProductTypeSaveFailed extends ProductTypeState {
  final String message;

  const ProductTypeSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductTypeDeleteInProgress extends ProductTypeState {}

class ProductTypeDeleteSuccess extends ProductTypeState {}

class ProductTypeDeleteFailed extends ProductTypeState {
  final String message;

  const ProductTypeDeleteFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductTypeDeleteManyInProgress extends ProductTypeState {}

class ProductTypeDeleteManySuccess extends ProductTypeState {}

class ProductTypeDeleteManyFailed extends ProductTypeState {
  final String message;

  const ProductTypeDeleteManyFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductTypeGetInProgress extends ProductTypeState {}

class ProductTypeGetSuccess extends ProductTypeState {
  final ProductTypeModel productType;

  const ProductTypeGetSuccess({required this.productType});

  ProductTypeGetSuccess copyWith({
    ProductTypeModel? productType,
  }) =>
      ProductTypeGetSuccess(productType: productType ?? this.productType);

  @override
  List<Object> get props => [productType];
}

class ProductTypeGetFailed extends ProductTypeState {
  final String message;

  const ProductTypeGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductTypeUpdateInitial extends ProductTypeState {}

class ProductTypeUpdateInProgress extends ProductTypeState {}

class ProductTypeUpdateSuccess extends ProductTypeState {}

class ProductTypeUpdateFailed extends ProductTypeState {
  final String message;

  const ProductTypeUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
