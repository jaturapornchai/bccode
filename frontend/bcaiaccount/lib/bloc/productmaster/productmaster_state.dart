part of 'productmaster_bloc.dart';

abstract class ProductMasterState extends Equatable {
  const ProductMasterState();

  @override
  List<Object> get props => [];
}

class ProductMasterInitial extends ProductMasterState {}

class ProductMasterInProgress extends ProductMasterState {}

class ProductMasterLoadSuccess extends ProductMasterState {
  final List<ProductMasterModel> productMasters;

  const ProductMasterLoadSuccess({required this.productMasters});

  ProductMasterLoadSuccess copyWith({
    List<ProductMasterModel>? productMasters,
  }) =>
      ProductMasterLoadSuccess(
          productMasters: productMasters ?? this.productMasters);

  @override
  List<Object> get props => [productMasters];
}

class ProductMasterLoadFailed extends ProductMasterState {
  final String message;

  const ProductMasterLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductMasterSaveInitial extends ProductMasterState {}

class ProductMasterSaveInProgress extends ProductMasterState {}

class ProductMasterSaveSuccess extends ProductMasterState {}

class ProductMasterSaveFailed extends ProductMasterState {
  final String message;

  const ProductMasterSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductMasterDeleteInProgress extends ProductMasterState {}

class ProductMasterDeleteSuccess extends ProductMasterState {}

class ProductMasterDeleteFailed extends ProductMasterState {}

class ProductMasterDeleteManyInProgress extends ProductMasterState {}

class ProductMasterDeleteManySuccess extends ProductMasterState {}

class ProductMasterDeleteManyFailed extends ProductMasterState {}

class ProductMasterGetInProgress extends ProductMasterState {}

class ProductMasterGetSuccess extends ProductMasterState {
  final ProductMasterModel productMaster;

  const ProductMasterGetSuccess({required this.productMaster});

  ProductMasterGetSuccess copyWith({
    ProductMasterModel? productMaster,
  }) =>
      ProductMasterGetSuccess(
          productMaster: productMaster ?? this.productMaster);

  @override
  List<Object> get props => [productMaster];
}

class ProductMasterGetFailed extends ProductMasterState {
  final String message;

  const ProductMasterGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductMasterUpdateInitial extends ProductMasterState {}

class ProductMasterUpdateInProgress extends ProductMasterState {}

class ProductMasterUpdateSuccess extends ProductMasterState {}

class ProductMasterUpdateFailed extends ProductMasterState {
  final String message;

  const ProductMasterUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
