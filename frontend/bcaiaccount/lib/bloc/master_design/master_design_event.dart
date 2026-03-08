part of 'master_design_bloc.dart';

abstract class MasterDesignEvent extends Equatable {
  const MasterDesignEvent();

  @override
  List<Object> get props => [];
}

class MasterDesignGet extends MasterDesignEvent {
  final String guid;

  const MasterDesignGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class MasterDesignLoadList extends MasterDesignEvent {
  final int limit;
  final int offset;
  final String search;

  const MasterDesignLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [offset, limit, search];
}

class MasterDesignGetByCode extends MasterDesignEvent {
  final String code;

  const MasterDesignGetByCode({required this.code});

  @override
  List<Object> get props => [code];
}

class MasterDesignDelete extends MasterDesignEvent {
  final String guid;

  const MasterDesignDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterDesignDeleteMany extends MasterDesignEvent {
  final List<String> guid;

  const MasterDesignDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterDesignSave extends MasterDesignEvent {
  final MasterDesignModel designModel;

  const MasterDesignSave({
    required this.designModel,
  });

  @override
  List<Object> get props => [designModel];
}

class MasterDesignUpdate extends MasterDesignEvent {
  final String guid;
  final MasterDesignModel designModel;

  const MasterDesignUpdate({
    required this.guid,
    required this.designModel,
  });

  @override
  List<Object> get props => [guid, designModel];
}
