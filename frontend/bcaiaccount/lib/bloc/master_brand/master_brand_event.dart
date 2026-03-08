part of 'master_brand_bloc.dart';

abstract class MasterBrandEvent extends Equatable {
  const MasterBrandEvent();

  @override
  List<Object> get props => [];
}

class MasterBrandGet extends MasterBrandEvent {
  final String guid;

  const MasterBrandGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class MasterBrandLoadList extends MasterBrandEvent {
  final int limit;
  final int offset;
  final String search;

  const MasterBrandLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [offset, limit, search];
}

class MasterBrandGetByCode extends MasterBrandEvent {
  final String code;

  const MasterBrandGetByCode({required this.code});

  @override
  List<Object> get props => [code];
}

class MasterBrandDelete extends MasterBrandEvent {
  final String guid;

  const MasterBrandDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterBrandDeleteMany extends MasterBrandEvent {
  final List<String> guid;

  const MasterBrandDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterBrandSave extends MasterBrandEvent {
  final MasterBrandModel brandModel;

  const MasterBrandSave({
    required this.brandModel,
  });

  @override
  List<Object> get props => [brandModel];
}

class MasterBrandUpdate extends MasterBrandEvent {
  final String guid;
  final MasterBrandModel brandModel;

  const MasterBrandUpdate({
    required this.guid,
    required this.brandModel,
  });

  @override
  List<Object> get props => [guid, brandModel];
}
