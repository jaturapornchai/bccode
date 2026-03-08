part of 'list_shop_bloc.dart';

abstract class ListShopState extends Equatable {
  const ListShopState();

  @override
  List<Object> get props => [];
}

class ListShopInitial extends ListShopState {}

class ListShopInProgress extends ListShopState {}

// ignore: must_be_immutable
class ListShopLoadSuccess extends ListShopState {
  List<ShopListModel> shop;

  ListShopLoadSuccess({required this.shop});

  @override
  List<Object> get props => [shop];
}

class ListShopLoadFailed extends ListShopState {
  final String message;
  const ListShopLoadFailed({required this.message});

  @override
  List<Object> get props => [message];
}
