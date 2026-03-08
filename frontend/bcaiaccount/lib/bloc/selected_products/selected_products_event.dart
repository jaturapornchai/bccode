part of 'selected_products_bloc.dart';

abstract class SelectedProductsEvent extends Equatable {
  const SelectedProductsEvent();

  @override
  List<Object?> get props => [];
}

class AddProductToSelection extends SelectedProductsEvent {
  final ProductBarcodeModel product;

  const AddProductToSelection(this.product);

  @override
  List<Object> get props => [product];
}

class AddProductsFromShelf extends SelectedProductsEvent {
  final List<ShelfProductItemModel> products;
  final String? shelfCode;
  final String? shelfName;

  const AddProductsFromShelf({
    required this.products,
    this.shelfCode,
    this.shelfName,
  });

  @override
  List<Object?> get props => [products, shelfCode, shelfName];
}

class AddAllProducts extends SelectedProductsEvent {
  final List<ProductBarcodeModel> products;

  const AddAllProducts(this.products);

  @override
  List<Object> get props => [products];
}

class RemoveProductFromSelection extends SelectedProductsEvent {
  final String barcode;

  const RemoveProductFromSelection(this.barcode);

  @override
  List<Object> get props => [barcode];
}

class UpdateProductCopies extends SelectedProductsEvent {
  final String barcode;
  final int copies;

  const UpdateProductCopies(this.barcode, this.copies);

  @override
  List<Object> get props => [barcode, copies];
}

class UpdateAllProductsCopies extends SelectedProductsEvent {
  final int copies;

  const UpdateAllProductsCopies(this.copies);

  @override
  List<Object> get props => [copies];
}

class ClearAllProducts extends SelectedProductsEvent {
  const ClearAllProducts();
}
