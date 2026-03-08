part of 'product_bloc.dart';

abstract class ProductEvent extends Equatable {
  const ProductEvent();

  @override
  List<Object> get props => [];
}

class ProductGet extends ProductEvent {
  final String guid;

  const ProductGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class ProductLoadList extends ProductEvent {
  final int limit;
  final int offset;
  final String search;

  const ProductLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class ProductDelete extends ProductEvent {
  final String guid;

  const ProductDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ProductDeleteMany extends ProductEvent {
  final List<String> guid;

  const ProductDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ProductSave extends ProductEvent {
  final ProductModel product;

  const ProductSave({
    required this.product,
  });

  @override
  List<Object> get props => [ProductModel];
}

class ProductUpdate extends ProductEvent {
  final String guid;
  final ProductModel product;

  const ProductUpdate({
    required this.guid,
    required this.product,
  });

  @override
  List<Object> get props => [ProductModel];
}

class ProductWithImageSave extends ProductEvent {
  final List<File> imageFile;
  final ProductModel product;
  final List<Uint8List> imageWeb;
  const ProductWithImageSave({
    required this.imageWeb,
    required this.imageFile,
    required this.product,
  });

  @override
  List<Object> get props => [ProductModel, imageFile];
}

class ProductWithImageUpdate extends ProductEvent {
  final String guid;
  final ProductModel product;
  final List<File> imageFile;
  final List<Uint8List> imageWeb;
  final List<ImagesModel> imagesUri;

  const ProductWithImageUpdate({
    required this.guid,
    required this.product,
    required this.imageFile,
    required this.imageWeb,
    required this.imagesUri,
  });

  @override
  List<Object> get props => [product];
}

class ProductBranchGet extends ProductEvent {
  final String branchcode;

  const ProductBranchGet({required this.branchcode});

  @override
  List<Object> get props => [branchcode];
}

class ProductBranchSave extends ProductEvent {
  final ProductBranchModel productBranchModel;

  const ProductBranchSave({required this.productBranchModel});

  @override
  List<Object> get props => [ProductBranchModel];
}

class ProductDepartmentGet extends ProductEvent {
  final String branchcode;
  final String departmentcode;

  const ProductDepartmentGet({
    required this.branchcode,
    required this.departmentcode,
  });

  @override
  List<Object> get props => [branchcode, departmentcode];
}

class ProductDepartmentSave extends ProductEvent {
  final ProductDepartmentModel productDepartmentModel;

  const ProductDepartmentSave({required this.productDepartmentModel});

  @override
  List<Object> get props => [ProductDepartmentModel];
}

class ProductGetByBarcodes extends ProductEvent {
  final List<String> barcodes;
  final String source;

  const ProductGetByBarcodes({required this.barcodes, this.source = ''});

  @override
  List<Object> get props => [barcodes, source];
}
