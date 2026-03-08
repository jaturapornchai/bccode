part of 'product_dimension_bloc.dart';

abstract class ProductDimensionState extends Equatable {
  const ProductDimensionState();

  @override
  List<Object> get props => [];
}

class ProductDimensionInitial extends ProductDimensionState {}

class ProductDimensionInProgress extends ProductDimensionState {}

class ProductDimensionLoadSuccess extends ProductDimensionState {
  final List<DimensionModel> productDimension;

  const ProductDimensionLoadSuccess({required this.productDimension});

  ProductDimensionLoadSuccess copyWith({
    List<DimensionModel>? productDimension,
  }) =>
      ProductDimensionLoadSuccess(
          productDimension: productDimension ?? this.productDimension);

  @override
  List<Object> get props => [productDimension];
}

class ProductDimensionLoadFailed extends ProductDimensionState {
  final String message;

  const ProductDimensionLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductDimensionSaveInitial extends ProductDimensionState {}

class ProductDimensionSaveInProgress extends ProductDimensionState {}

class ProductDimensionSaveSuccess extends ProductDimensionState {}

class ProductDimensionSaveFailed extends ProductDimensionState {
  final String message;

  const ProductDimensionSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductDimensionDeleteInProgress extends ProductDimensionState {}

class ProductDimensionDeleteSuccess extends ProductDimensionState {}

class ProductDimensionDeleteFailed extends ProductDimensionState {
  final String message;

  const ProductDimensionDeleteFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductDimensionDeleteManyInProgress extends ProductDimensionState {}

class ProductDimensionDeleteManySuccess extends ProductDimensionState {}

class ProductDimensionDeleteManyFailed extends ProductDimensionState {
  final String message;

  const ProductDimensionDeleteManyFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductDimensionGetInProgress extends ProductDimensionState {}

class ProductDimensionGetSuccess extends ProductDimensionState {
  final DimensionModel productDimension;

  const ProductDimensionGetSuccess({required this.productDimension});

  ProductDimensionGetSuccess copyWith({
    DimensionModel? productDimension,
  }) =>
      ProductDimensionGetSuccess(
          productDimension: productDimension ?? this.productDimension);

  @override
  List<Object> get props => [productDimension];
}

class ProductDimensionGetFailed extends ProductDimensionState {
  final String message;

  const ProductDimensionGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductDimensionUpdateInitial extends ProductDimensionState {}

class ProductDimensionUpdateInProgress extends ProductDimensionState {}

class ProductDimensionUpdateSuccess extends ProductDimensionState {}

class ProductDimensionUpdateFailed extends ProductDimensionState {
  final String message;

  const ProductDimensionUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
