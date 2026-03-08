part of 'shop_select_bloc.dart';

abstract class ShopSelectEvent extends Equatable {
  const ShopSelectEvent();

  @override
  List<Object> get props => [];
}

// ignore: must_be_immutable
class ShopSelect extends ShopSelectEvent {
  ShopListModel shop;
  ShopSelect({required this.shop});

  @override
  List<Object> get props => [];
}
