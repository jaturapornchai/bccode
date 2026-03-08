part of 'product_group_bloc.dart';

abstract class ProductGroupEvent extends Equatable {
  const ProductGroupEvent();

  @override
  List<Object> get props => [];
}

class ProductGroupGet extends ProductGroupEvent {
  final String guid;

  const ProductGroupGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class ProductGroupLoadList extends ProductGroupEvent {
  final int limit;
  final int offset;
  final String search;

  const ProductGroupLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class ProductGroupDelete extends ProductGroupEvent {
  final String guid;

  const ProductGroupDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ProductGroupDeleteMany extends ProductGroupEvent {
  final List<String> guid;

  const ProductGroupDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ProductGroupSave extends ProductGroupEvent {
  final ProductGroupModel productGroup;

  const ProductGroupSave({
    required this.productGroup,
  });

  @override
  List<Object> get props => [productGroup];
}

class ProductGroupUpdateXOrder extends ProductGroupEvent {
  final List<XSortModel> orderLists;

  const ProductGroupUpdateXOrder({
    required this.orderLists,
  });

  @override
  List<Object> get props => [List<XSortModel>];
}

class ProductGroupUpdate extends ProductGroupEvent {
  final String guid;
  final ProductGroupModel productGroup;

  const ProductGroupUpdate({
    required this.guid,
    required this.productGroup,
  });

  @override
  List<Object> get props => [productGroup];
}

class ProductGroupWithImageSave extends ProductGroupEvent {
  final File imageFile;
  final ProductGroupModel productGroup;
  final Uint8List? imageWeb;
  const ProductGroupWithImageSave({
    required this.imageWeb,
    required this.imageFile,
    required this.productGroup,
  });

  @override
  List<Object> get props => [productGroup, imageFile];
}

class ProductGroupWithImageUpdate extends ProductGroupEvent {
  final String guid;
  final ProductGroupModel productGroup;
  final File imageFile;
  final Uint8List imageWeb;
  const ProductGroupWithImageUpdate({
    required this.guid,
    required this.imageFile,
    required this.imageWeb,
    required this.productGroup,
  });

  @override
  List<Object> get props => [productGroup, imageWeb, productGroup];
}
