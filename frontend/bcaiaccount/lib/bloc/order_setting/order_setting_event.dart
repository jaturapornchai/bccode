part of 'order_setting_bloc.dart';

abstract class OrderSettingEvent extends Equatable {
  const OrderSettingEvent();

  @override
  List<Object> get props => [];
}

class OrderSettingGet extends OrderSettingEvent {
  final String guid;

  const OrderSettingGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class OrderSettingLoadList extends OrderSettingEvent {
  final int limit;
  final int offset;
  final String search;

  const OrderSettingLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class OrderSettingDelete extends OrderSettingEvent {
  final String guid;

  const OrderSettingDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class OrderSettingDeleteMany extends OrderSettingEvent {
  final List<String> guid;

  const OrderSettingDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class OrderSettingSave extends OrderSettingEvent {
  final OrderSettingModel orderSetting;

  const OrderSettingSave({
    required this.orderSetting,
  });

  @override
  List<Object> get props => [orderSetting];
}

class OrderSettingUpdate extends OrderSettingEvent {
  final String guid;
  final OrderSettingModel orderSetting;

  const OrderSettingUpdate({
    required this.guid,
    required this.orderSetting,
  });

  @override
  List<Object> get props => [orderSetting];
}
