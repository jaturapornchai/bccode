part of 'master_pattern_bloc.dart';

abstract class MasterPatternEvent extends Equatable {
  const MasterPatternEvent();

  @override
  List<Object> get props => [];
}

class MasterPatternGet extends MasterPatternEvent {
  final String guid;

  const MasterPatternGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class MasterPatternLoadList extends MasterPatternEvent {
  final int limit;
  final int offset;
  final String search;

  const MasterPatternLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [offset, limit, search];
}

class MasterPatternGetByCode extends MasterPatternEvent {
  final String code;

  const MasterPatternGetByCode({required this.code});

  @override
  List<Object> get props => [code];
}

class MasterPatternDelete extends MasterPatternEvent {
  final String guid;

  const MasterPatternDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterPatternDeleteMany extends MasterPatternEvent {
  final List<String> guid;

  const MasterPatternDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterPatternSave extends MasterPatternEvent {
  final MasterPatternModel patternModel;

  const MasterPatternSave({
    required this.patternModel,
  });

  @override
  List<Object> get props => [patternModel];
}

class MasterPatternUpdate extends MasterPatternEvent {
  final String guid;
  final MasterPatternModel patternModel;

  const MasterPatternUpdate({
    required this.guid,
    required this.patternModel,
  });

  @override
  List<Object> get props => [guid, patternModel];
}
