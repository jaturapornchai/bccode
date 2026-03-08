part of 'shelf_product_bloc.dart';

abstract class ShelfProductState extends Equatable {
  const ShelfProductState();

  @override
  List<Object> get props => [];
}

/// Initial state
class ShelfProductInitial extends ShelfProductState {}

/// Loading states
class ShelfProductAddInProgress extends ShelfProductState {}

class ShelfProductRemoveInProgress extends ShelfProductState {}

class ShelfProductLoadInProgress extends ShelfProductState {}

/// Success states
class ShelfProductAddSuccess extends ShelfProductState {
  final String message;

  const ShelfProductAddSuccess({this.message = 'Products added successfully'});

  @override
  List<Object> get props => [message];
}

class ShelfProductRemoveSuccess extends ShelfProductState {
  final String message;

  const ShelfProductRemoveSuccess(
      {this.message = 'Products removed successfully'});

  @override
  List<Object> get props => [message];
}

class ShelfProductLoadSuccess extends ShelfProductState {
  final List<ShelfProductDisplayModel> products;

  const ShelfProductLoadSuccess({required this.products});

  ShelfProductLoadSuccess copyWith({
    List<ShelfProductDisplayModel>? products,
  }) =>
      ShelfProductLoadSuccess(products: products ?? this.products);

  @override
  List<Object> get props => [products];
}

/// Failure states
class ShelfProductAddFailed extends ShelfProductState {
  final String message;

  const ShelfProductAddFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class ShelfProductRemoveFailed extends ShelfProductState {
  final String message;

  const ShelfProductRemoveFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class ShelfProductLoadFailed extends ShelfProductState {
  final String message;

  const ShelfProductLoadFailed({required this.message});

  @override
  List<Object> get props => [message];
}
