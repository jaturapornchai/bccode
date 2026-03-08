part of 'order_type_bloc.dart';

abstract class OrderTypeEvent extends Equatable {
  const OrderTypeEvent();

  @override
  List<Object> get props => [];
}

class OrderTypeGet extends OrderTypeEvent {
  final String guid;

  const OrderTypeGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class OrderTypeLoadList extends OrderTypeEvent {
  final int limit;
  final int offset;
  final String search;

  const OrderTypeLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class OrderTypeDelete extends OrderTypeEvent {
  final String guid;

  const OrderTypeDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class OrderTypeDeleteMany extends OrderTypeEvent {
  final List<String> guid;

  const OrderTypeDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class OrderTypeSave extends OrderTypeEvent {
  final OrderTypeModel ordertypemodel;

  const OrderTypeSave({
    required this.ordertypemodel,
  });

  @override
  List<Object> get props => [ordertypemodel];
}

class OrderTypeUpdate extends OrderTypeEvent {
  final String guid;
  final OrderTypeModel ordertypemodel;

  const OrderTypeUpdate({
    required this.guid,
    required this.ordertypemodel,
  });

  @override
  List<Object> get props => [ordertypemodel];
}
