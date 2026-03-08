part of 'product_type_bloc.dart';

abstract class ProductTypeEvent extends Equatable {
  const ProductTypeEvent();

  @override
  List<Object> get props => [];
}

class ProductTypeGet extends ProductTypeEvent {
  final String guid;

  const ProductTypeGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class ProductTypeLoadList extends ProductTypeEvent {
  final int limit;
  final int offset;
  final String search;

  const ProductTypeLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class ProductTypeDelete extends ProductTypeEvent {
  final String guid;

  const ProductTypeDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ProductTypeDeleteMany extends ProductTypeEvent {
  final List<String> guid;

  const ProductTypeDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ProductTypeSave extends ProductTypeEvent {
  final ProductTypeModel productTypemodel;

  const ProductTypeSave({
    required this.productTypemodel,
  });

  @override
  List<Object> get props => [productTypemodel];
}

class ProductTypeUpdate extends ProductTypeEvent {
  final String guid;
  final ProductTypeModel productTypemodel;

  const ProductTypeUpdate({
    required this.guid,
    required this.productTypemodel,
  });

  @override
  List<Object> get props => [productTypemodel];
}
