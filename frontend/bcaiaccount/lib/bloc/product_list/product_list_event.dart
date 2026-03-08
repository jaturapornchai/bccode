part of 'product_list_bloc.dart';

abstract class ProductListEvent extends Equatable {
  const ProductListEvent();

  @override
  List<Object?> get props => [];
}

class LoadProducts extends ProductListEvent {
  final String search;
  final String branchcode;
  final String businesstypecode;

  const LoadProducts({
    required this.search,
    required this.branchcode,
    required this.businesstypecode,
  });

  @override
  List<Object> get props => [search, branchcode, businesstypecode];
}

class RefreshProducts extends ProductListEvent {
  final String search;
  final String branchcode;
  final String businesstypecode;

  const RefreshProducts({
    required this.search,
    required this.branchcode,
    required this.businesstypecode,
  });

  @override
  List<Object> get props => [search, branchcode, businesstypecode];
}

class SearchProducts extends ProductListEvent {
  final String search;
  final String branchcode;
  final String businesstypecode;

  const SearchProducts({
    required this.search,
    required this.branchcode,
    required this.businesstypecode,
  });

  @override
  List<Object> get props => [search, branchcode, businesstypecode];
}

class LoadMoreProducts extends ProductListEvent {
  final String search;
  final String branchcode;
  final String businesstypecode;

  const LoadMoreProducts({
    required this.search,
    required this.branchcode,
    required this.businesstypecode,
  });

  @override
  List<Object> get props => [search, branchcode, businesstypecode];
}

class CancelCurrentRequest extends ProductListEvent {
  const CancelCurrentRequest();
}

class ClearProductList extends ProductListEvent {
  const ClearProductList();
}
