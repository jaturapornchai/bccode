part of 'selected_products_bloc.dart';

abstract class SelectedProductsState extends Equatable {
  const SelectedProductsState();

  @override
  List<Object?> get props => [];
}

class SelectedProductsInitial extends SelectedProductsState {
  const SelectedProductsInitial();
}

class SelectedProductsLoaded extends SelectedProductsState {
  final List<ProductWithCopies> selectedProducts;

  const SelectedProductsLoaded({
    required this.selectedProducts,
  });

  @override
  List<Object> get props => [selectedProducts];

  int get totalItems => selectedProducts.length;
  int get totalCopies => selectedProducts.fold<int>(0, (sum, item) => sum + item.copies);

  bool isProductSelected(String barcode) {
    return selectedProducts.any((item) => item.product.barcode == barcode);
  }

  SelectedProductsLoaded copyWith({
    List<ProductWithCopies>? selectedProducts,
  }) {
    return SelectedProductsLoaded(
      selectedProducts: selectedProducts ?? this.selectedProducts,
    );
  }
}

class SelectedProductsError extends SelectedProductsState {
  final String message;

  const SelectedProductsError(this.message);

  @override
  List<Object> get props => [message];
}
