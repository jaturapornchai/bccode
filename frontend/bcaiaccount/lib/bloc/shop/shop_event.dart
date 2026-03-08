part of 'shop_bloc.dart';

abstract class ShopEvent extends Equatable {
  const ShopEvent();

  @override
  List<Object> get props => [];
}

class GetShopInfo extends ShopEvent {
  final String shopid;

  const GetShopInfo({required this.shopid});

  @override
  List<Object> get props => [shopid];
}

class GetMainShopCenterTypes extends ShopEvent {
  final String mainShopId;

  const GetMainShopCenterTypes({required this.mainShopId});

  @override
  List<Object> get props => [mainShopId];
}

class ShopUpdate extends ShopEvent {
  final String shopid;
  final ShopModel shopdata;

  const ShopUpdate({
    required this.shopid,
    required this.shopdata,
  });

  @override
  List<Object> get props => [shopid, shopdata];
}
