part of 'productmaster_bloc.dart';

abstract class ProductMasterEvent extends Equatable {
  const ProductMasterEvent();

  @override
  List<Object> get props => [];
}

class ProductMasterGet extends ProductMasterEvent {
  final String guid;

  const ProductMasterGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class ProductMasterLoadList extends ProductMasterEvent {
  final int limit;
  final int page;
  final String search;

  const ProductMasterLoadList(
      {required this.page, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class ProductMasterDelete extends ProductMasterEvent {
  final String guid;

  const ProductMasterDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ProductMasterDeleteMany extends ProductMasterEvent {
  final List<String> guid;

  const ProductMasterDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ProductMasterSave extends ProductMasterEvent {
  final ProductMasterModel productMasterModel;

  const ProductMasterSave({
    required this.productMasterModel,
  });

  @override
  List<Object> get props => [productMasterModel];
}

class ProductMasterSaveBulk extends ProductMasterEvent {
  final List<ProductMasterModel> productMasters;

  const ProductMasterSaveBulk({
    required this.productMasters,
  });

  @override
  List<Object> get props => [productMasters];
}

class ProductMasterUpdate extends ProductMasterEvent {
  final String guid;
  final ProductMasterModel productMasterModel;

  const ProductMasterUpdate({
    required this.guid,
    required this.productMasterModel,
  });

  @override
  List<Object> get props => [productMasterModel];
}
