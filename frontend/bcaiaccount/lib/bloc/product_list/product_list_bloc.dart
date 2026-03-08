import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/services/product_service.dart';

part 'product_list_event.dart';
part 'product_list_state.dart';

class ProductListBloc extends Bloc<ProductListEvent, ProductListState> {
  final ProductService _productService;
  CancelToken? _currentCancelToken;

  ProductListBloc({required ProductService productService})
      : _productService = productService,
        super(const ProductListInitial()) {
    on<LoadProducts>(_onLoadProducts);
    on<RefreshProducts>(_onRefreshProducts);
    on<SearchProducts>(_onSearchProducts);
    on<LoadMoreProducts>(_onLoadMoreProducts);
    on<CancelCurrentRequest>(_onCancelCurrentRequest);
    on<ClearProductList>(_onClearProductList);
  }

  void _onLoadProducts(
    LoadProducts event,
    Emitter<ProductListState> emit,
  ) async {
    emit(const ProductListLoading());
    await _loadProducts(
      emit: emit,
      offset: 0,
      search: event.search,
      branchcode: event.branchcode,
      businesstypecode: event.businesstypecode,
      isRefresh: true,
    );
  }

  void _onRefreshProducts(
    RefreshProducts event,
    Emitter<ProductListState> emit,
  ) async {
    final currentState = state;
    if (currentState is ProductListLoaded) {
      emit(currentState.copyWith(isRefreshing: true));
    } else {
      emit(const ProductListLoading());
    }

    await _loadProducts(
      emit: emit,
      offset: 0,
      search: event.search,
      branchcode: event.branchcode,
      businesstypecode: event.businesstypecode,
      isRefresh: true,
    );
  }

  void _onSearchProducts(
    SearchProducts event,
    Emitter<ProductListState> emit,
  ) async {
    // Cancel current request if any
    _cancelCurrentRequest();

    emit(const ProductListLoading());
    await _loadProducts(
      emit: emit,
      offset: 0,
      search: event.search,
      branchcode: event.branchcode,
      businesstypecode: event.businesstypecode,
      isRefresh: true,
    );
  }

  void _onLoadMoreProducts(
    LoadMoreProducts event,
    Emitter<ProductListState> emit,
  ) async {
    final currentState = state;
    if (currentState is ProductListLoaded && !currentState.isLoadingMore && currentState.hasMore) {
      emit(currentState.copyWith(isLoadingMore: true));

      await _loadProducts(
        emit: emit,
        offset: currentState.products.length,
        search: event.search,
        branchcode: event.branchcode,
        businesstypecode: event.businesstypecode,
        currentProducts: currentState.products,
      );
    }
  }

  void _onCancelCurrentRequest(
    CancelCurrentRequest event,
    Emitter<ProductListState> emit,
  ) {
    _cancelCurrentRequest();
  }

  Future<void> _loadProducts({
    required Emitter<ProductListState> emit,
    required int offset,
    required String search,
    required String branchcode,
    required String businesstypecode,
    List<ProductBarcodeModel>? currentProducts,
    bool isRefresh = false,
  }) async {
    // Cancel previous request
    _cancelCurrentRequest();

    // Create new cancel token
    _currentCancelToken = CancelToken();

    try {
      final products = await _productService.getProducts(
        offset: offset,
        limit: 20, // Load 20 items per page
        search: search,
        branchcode: branchcode,
        businesstypecode: businesstypecode,
        cancelToken: _currentCancelToken,
      );

      if (_currentCancelToken?.isCancelled == true) {
        return; // Request was cancelled
      }

      final allProducts = isRefresh || currentProducts == null ? products : [...currentProducts, ...products];

      emit(ProductListLoaded(
        products: allProducts,
        hasMore: products.length >= 20, // Assume no more data if less than page size
        isLoadingMore: false,
        isRefreshing: false,
        searchQuery: search,
      ));

      _currentCancelToken = null;
    } catch (e) {
      if (_currentCancelToken?.isCancelled == true) {
        return; // Request was cancelled, don't emit error
      }

      _currentCancelToken = null;

      if (e is CancelledException) {
        return; // Don't emit error for cancelled requests
      }

      final currentState = state;
      if (currentState is ProductListLoaded) {
        emit(currentState.copyWith(
          error: e.toString(),
          isLoadingMore: false,
          isRefreshing: false,
        ));
      } else {
        emit(ProductListError(e.toString()));
      }
    }
  }

  void _cancelCurrentRequest() {
    _currentCancelToken?.cancel();
    _currentCancelToken = null;
  }

  void _onClearProductList(
    ClearProductList event,
    Emitter<ProductListState> emit,
  ) {
    _cancelCurrentRequest();
    emit(const ProductListInitial());
  }

  @override
  Future<void> close() {
    _cancelCurrentRequest();
    _productService.dispose();
    return super.close();
  }
}
