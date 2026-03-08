part of 'zone_bloc.dart';

abstract class ZoneEvent extends Equatable {
  const ZoneEvent();

  @override
  List<Object> get props => [];
}

class ZoneGet extends ZoneEvent {
  final String guid;

  const ZoneGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class ZoneLoadList extends ZoneEvent {
  final int limit;
  final int offset;
  final String search;
  final int groupNumber;

  const ZoneLoadList({
    required this.offset,
    required this.limit,
    required this.search,
    required this.groupNumber,
  });

  @override
  List<Object> get props => [];
}

class ZoneLoadBarcodeAllList extends ZoneEvent {
  final int limit;
  final int offset;
  final String search;

  const ZoneLoadBarcodeAllList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class ZoneDelete extends ZoneEvent {
  final String guid;

  const ZoneDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ZoneDeleteMany extends ZoneEvent {
  final List<String> guid;

  const ZoneDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class ZoneSave extends ZoneEvent {
  final ZoneDataModel zoneDataModel;

  const ZoneSave({
    required this.zoneDataModel,
  });

  @override
  List<Object> get props => [zoneDataModel];
}

class ZoneUpdate extends ZoneEvent {
  final String guid;
  final ZoneDataModel zoneDataModel;

  const ZoneUpdate({
    required this.guid,
    required this.zoneDataModel,
  });

  @override
  List<Object> get props => [zoneDataModel];
}
