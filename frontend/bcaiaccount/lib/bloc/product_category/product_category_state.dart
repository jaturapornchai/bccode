part of 'product_category_bloc.dart';

abstract class ProductCategoryState extends Equatable {
  const ProductCategoryState();

  @override
  List<Object> get props => [];
}

class ProductCategoryInitial extends ProductCategoryState {}

class ProductCategoryInProgress extends ProductCategoryState {}

class ProductCategoryLoadSuccess extends ProductCategoryState {
  final List<ProductCategoryModel> productCategorys;

  const ProductCategoryLoadSuccess({required this.productCategorys});

  ProductCategoryLoadSuccess copyWith({
    List<ProductCategoryModel>? categorys,
  }) =>
      ProductCategoryLoadSuccess(
          productCategorys: categorys ?? productCategorys);

  @override
  List<Object> get props => [productCategorys];
}

class ProductCategoryLoadFailed extends ProductCategoryState {
  final String message;

  const ProductCategoryLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductCategorySaveInitial extends ProductCategoryState {}

class ProductCategorySaveInProgress extends ProductCategoryState {}

class ProductCategorySaveSuccess extends ProductCategoryState {}

class ProductCategorySaveFailed extends ProductCategoryState {
  final String message;

  const ProductCategorySaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductCategoryDeleteInProgress extends ProductCategoryState {}

class ProductCategoryDeleteSuccess extends ProductCategoryState {}

class ProductCategoryDeleteFailed extends ProductCategoryState {}

class ProductCategoryDeleteManyInProgress extends ProductCategoryState {}

class ProductCategoryDeleteManySuccess extends ProductCategoryState {}

class ProductCategoryDeleteManyFailed extends ProductCategoryState {}

class ProductCategoryGetInProgress extends ProductCategoryState {}

class ProductCategoryGetSuccess extends ProductCategoryState {
  final ProductCategoryModel category;

  const ProductCategoryGetSuccess({required this.category});

  ProductCategoryGetSuccess copyWith({
    ProductCategoryModel? category,
  }) =>
      ProductCategoryGetSuccess(category: category ?? this.category);

  @override
  List<Object> get props => [category];
}

class ProductCategoryGetFailed extends ProductCategoryState {
  final String message;

  const ProductCategoryGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class ProductCategoryUpdateInitial extends ProductCategoryState {}

class ProductCategoryUpdateInProgress extends ProductCategoryState {}

class ProductCategoryUpdateSuccess extends ProductCategoryState {}

class ProductCategoryUpdateFailed extends ProductCategoryState {
  final String message;

  const ProductCategoryUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
