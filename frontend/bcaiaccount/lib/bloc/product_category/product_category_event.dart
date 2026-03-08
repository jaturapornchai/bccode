part of 'product_category_bloc.dart';

abstract class ProductCategoryEvent extends Equatable {
  const ProductCategoryEvent();

  @override
  List<Object> get props => [];
}

class ProductCategoryGet extends ProductCategoryEvent {
  final String guid;

  const ProductCategoryGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class ProductCategoryLoadList extends ProductCategoryEvent {
  final int? limit;
  final int? offset;
  final String? search;
  final int? groupNumber;

  const ProductCategoryLoadList(
      {this.offset, this.limit, this.search, this.groupNumber});

  @override
  List<Object> get props => [];
}

class ProductCategoryDelete extends ProductCategoryEvent {
  final String guid;

  const ProductCategoryDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ProductCategoryDeleteMany extends ProductCategoryEvent {
  final List<String> guid;

  const ProductCategoryDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ProductCategorySave extends ProductCategoryEvent {
  final ProductCategoryModel category;

  const ProductCategorySave({
    required this.category,
  });

  @override
  List<Object> get props => [category];
}

class ProductCategoryUpdateXOrder extends ProductCategoryEvent {
  final List<XSortModel> orderLists;

  const ProductCategoryUpdateXOrder({
    required this.orderLists,
  });

  @override
  List<Object> get props => [List<XSortModel>];
}

class ProductCategoryUpdate extends ProductCategoryEvent {
  final String guid;
  final ProductCategoryModel category;

  const ProductCategoryUpdate({
    required this.guid,
    required this.category,
  });

  @override
  List<Object> get props => [category];
}

class ProductCategoryWithImageSave extends ProductCategoryEvent {
  final File imageFile;
  final ProductCategoryModel category;
  final Uint8List? imageWeb;
  const ProductCategoryWithImageSave({
    required this.imageWeb,
    required this.imageFile,
    required this.category,
  });

  @override
  List<Object> get props => [category, imageFile];
}

class ProductCategoryWithImageUpdate extends ProductCategoryEvent {
  final String guid;
  final ProductCategoryModel category;
  final File imageFile;
  final Uint8List imageWeb;
  const ProductCategoryWithImageUpdate({
    required this.guid,
    required this.imageFile,
    required this.imageWeb,
    required this.category,
  });

  @override
  List<Object> get props => [category, imageWeb, category];
}
