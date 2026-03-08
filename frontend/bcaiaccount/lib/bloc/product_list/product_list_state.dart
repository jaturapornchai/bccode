part of 'product_list_bloc.dart';

abstract class ProductListState extends Equatable {
  const ProductListState();

  @override
  List<Object?> get props => [];
}

class ProductListInitial extends ProductListState {
  const ProductListInitial();
}

class ProductListLoading extends ProductListState {
  const ProductListLoading();
}

class ProductListLoaded extends ProductListState {
  final List<ProductBarcodeModel> products;
  final bool hasMore;
  final bool isLoadingMore;
  final bool isRefreshing;
  final String searchQuery;
  final String? error;

  const ProductListLoaded({
    required this.products,
    this.hasMore = true,
    this.isLoadingMore = false,
    this.isRefreshing = false,
    this.searchQuery = '',
    this.error,
  });

  @override
  List<Object?> get props => [
        products,
        hasMore,
        isLoadingMore,
        isRefreshing,
        searchQuery,
        error,
      ];

  ProductListLoaded copyWith({
    List<ProductBarcodeModel>? products,
    bool? hasMore,
    bool? isLoadingMore,
    bool? isRefreshing,
    String? searchQuery,
    String? error,
  }) {
    return ProductListLoaded(
      products: products ?? this.products,
      hasMore: hasMore ?? this.hasMore,
      isLoadingMore: isLoadingMore ?? this.isLoadingMore,
      isRefreshing: isRefreshing ?? this.isRefreshing,
      searchQuery: searchQuery ?? this.searchQuery,
      error: error,
    );
  }
}

class ProductListError extends ProductListState {
  final String message;

  const ProductListError(this.message);

  @override
  List<Object> get props => [message];
}
