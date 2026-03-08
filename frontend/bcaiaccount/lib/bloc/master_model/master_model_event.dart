part of 'master_model_bloc.dart';

abstract class MasterModelEvent extends Equatable {
  const MasterModelEvent();

  @override
  List<Object> get props => [];
}

class MasterModelGet extends MasterModelEvent {
  final String guid;

  const MasterModelGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class MasterModelLoadList extends MasterModelEvent {
  final int limit;
  final int offset;
  final String search;

  const MasterModelLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [offset, limit, search];
}

class MasterModelGetByCode extends MasterModelEvent {
  final String code;

  const MasterModelGetByCode({required this.code});

  @override
  List<Object> get props => [code];
}

class MasterModelDelete extends MasterModelEvent {
  final String guid;

  const MasterModelDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterModelDeleteMany extends MasterModelEvent {
  final List<String> guid;

  const MasterModelDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterModelSave extends MasterModelEvent {
  final MasterModelModel modelModel;

  const MasterModelSave({
    required this.modelModel,
  });

  @override
  List<Object> get props => [modelModel];
}

class MasterModelUpdate extends MasterModelEvent {
  final String guid;
  final MasterModelModel modelModel;

  const MasterModelUpdate({
    required this.guid,
    required this.modelModel,
  });

  @override
  List<Object> get props => [guid, modelModel];
}
