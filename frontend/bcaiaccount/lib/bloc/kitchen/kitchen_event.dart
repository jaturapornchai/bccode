part of 'kitchen_bloc.dart';

abstract class KitchenEvent extends Equatable {
  const KitchenEvent();

  @override
  List<Object> get props => [];
}

class KitchenGet extends KitchenEvent {
  final String guid;

  const KitchenGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class KitchenLoadList extends KitchenEvent {
  final int limit;
  final int offset;
  final String search;
  final int groupnumber;

  const KitchenLoadList(
      {required this.offset,
      required this.limit,
      required this.search,
      required this.groupnumber});

  @override
  List<Object> get props => [];
}

class KitchenProductsLoadList extends KitchenEvent {}

class KitchenDelete extends KitchenEvent {
  final String guid;

  const KitchenDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class KitchenDeleteMany extends KitchenEvent {
  final List<String> guid;

  const KitchenDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class KitchenSave extends KitchenEvent {
  final KitchenModel kitchenModel;

  const KitchenSave({
    required this.kitchenModel,
  });

  @override
  List<Object> get props => [kitchenModel];
}

class KitchenUpdate extends KitchenEvent {
  final String guid;
  final KitchenModel kitchenModel;

  const KitchenUpdate({
    required this.guid,
    required this.kitchenModel,
  });

  @override
  List<Object> get props => [kitchenModel];
}

class KitchenDetailBarcodeGet extends KitchenEvent {
  final String guid;

  const KitchenDetailBarcodeGet({required this.guid});

  @override
  List<Object> get props => [guid];
}
