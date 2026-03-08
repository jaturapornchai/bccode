part of 'pos_media_bloc.dart';

abstract class PosMediaEvent extends Equatable {
  const PosMediaEvent();

  @override
  List<Object> get props => [];
}

class PosMediaGet extends PosMediaEvent {
  final String guid;

  const PosMediaGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class PosMediaLoadList extends PosMediaEvent {
  final int limit;
  final int offset;
  final String search;

  const PosMediaLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class PosMediaDelete extends PosMediaEvent {
  final String guid;

  const PosMediaDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class PosMediaDeleteMany extends PosMediaEvent {
  final List<String> guid;

  const PosMediaDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class PosMediaSave extends PosMediaEvent {
  final PosMediaModel posMedia;

  const PosMediaSave({
    required this.posMedia,
  });

  @override
  List<Object> get props => [posMedia];
}

class PosMediaUpdate extends PosMediaEvent {
  final String guid;
  final PosMediaModel posMedia;

  const PosMediaUpdate({
    required this.guid,
    required this.posMedia,
  });

  @override
  List<Object> get props => [posMedia];
}
