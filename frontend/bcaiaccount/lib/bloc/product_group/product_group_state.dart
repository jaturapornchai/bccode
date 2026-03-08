part of 'product_group_bloc.dart';

abstract class ProductGroupState extends Equatable {
  const ProductGroupState();

  @override
  List<Object> get props => [];
}

class ProductGroupInitial extends ProductGroupState {}

class ProductGroupInProgress extends ProductGroupState {}

class ProductGroupLoadSuccess extends ProductGroupState {
  final List<ProductGroupModel> productGroups;

  const ProductGroupLoadSuccess({required this.productGroups});

  ProductGroupLoadSuccess copyWith({
    List<ProductGroupModel>? productGroups,
  }) =>
      ProductGroupLoadSuccess(
          productGroups: productGroups ?? this.productGroups);

  @override
  List<Object> get props => [productGroups];
}

class ProductGroupLoadFailed extends ProductGroupState {
  final String message;

  const ProductGroupLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductGroupSaveInitial extends ProductGroupState {}

class ProductGroupSaveInProgress extends ProductGroupState {}

class ProductGroupSaveSuccess extends ProductGroupState {}

class ProductGroupSaveFailed extends ProductGroupState {
  final String message;

  const ProductGroupSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductGroupDeleteInProgress extends ProductGroupState {}

class ProductGroupDeleteSuccess extends ProductGroupState {}

class ProductGroupDeleteFailed extends ProductGroupState {}

class ProductGroupDeleteManyInProgress extends ProductGroupState {}

class ProductGroupDeleteManySuccess extends ProductGroupState {}

class ProductGroupDeleteManyFailed extends ProductGroupState {}

class ProductGroupGetInProgress extends ProductGroupState {}

class ProductGroupGetSuccess extends ProductGroupState {
  final ProductGroupModel productGroup;

  const ProductGroupGetSuccess({required this.productGroup});

  ProductGroupGetSuccess copyWith({
    ProductGroupModel? productGroup,
  }) =>
      ProductGroupGetSuccess(productGroup: productGroup ?? this.productGroup);

  @override
  List<Object> get props => [productGroup];
}

class ProductGroupGetFailed extends ProductGroupState {
  final String message;

  const ProductGroupGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductGroupUpdateInitial extends ProductGroupState {}

class ProductGroupUpdateInProgress extends ProductGroupState {}

class ProductGroupUpdateSuccess extends ProductGroupState {}

class ProductGroupUpdateFailed extends ProductGroupState {
  final String message;

  const ProductGroupUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
