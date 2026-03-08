part of 'product_dimension_bloc.dart';

abstract class ProductDimensionEvent extends Equatable {
  const ProductDimensionEvent();

  @override
  List<Object> get props => [];
}

class ProductDimensionGet extends ProductDimensionEvent {
  final String guid;

  const ProductDimensionGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class ProductDimensionLoadList extends ProductDimensionEvent {
  final int limit;
  final int offset;
  final String search;

  const ProductDimensionLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class ProductDimensionDelete extends ProductDimensionEvent {
  final String guid;

  const ProductDimensionDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ProductDimensionDeleteMany extends ProductDimensionEvent {
  final List<String> guid;

  const ProductDimensionDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ProductDimensionSave extends ProductDimensionEvent {
  final DimensionModel productDimensionmodel;

  const ProductDimensionSave({
    required this.productDimensionmodel,
  });

  @override
  List<Object> get props => [productDimensionmodel];
}

class ProductDimensionUpdate extends ProductDimensionEvent {
  final String guid;
  final DimensionModel productDimensionmodel;

  const ProductDimensionUpdate({
    required this.guid,
    required this.productDimensionmodel,
  });

  @override
  List<Object> get props => [productDimensionmodel];
}
