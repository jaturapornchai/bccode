part of 'qr_bloc.dart';

abstract class QrEvent extends Equatable {
  const QrEvent();

  @override
  List<Object> get props => [];
}

class QrGet extends QrEvent {
  final String guid;

  const QrGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class QrLoadList extends QrEvent {
  final int limit;
  final int offset;
  final String search;

  const QrLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class QrDelete extends QrEvent {
  final String guid;

  const QrDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class QrDeleteMany extends QrEvent {
  final List<String> guid;

  const QrDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class QrSave extends QrEvent {
  final QrModel qr;

  const QrSave({
    required this.qr,
  });

  @override
  List<Object> get props => [qr];
}

class QrBulkSave extends QrEvent {
  final List<QrModel> qrs;

  const QrBulkSave({
    required this.qrs,
  });

  @override
  List<Object> get props => [qrs];
}

class QrSaveWithImage extends QrEvent {
  final QrModel qr;
  final File imageFile;
  final Uint8List imageWeb;

  const QrSaveWithImage({
    required this.imageWeb,
    required this.imageFile,
    required this.qr,
  });

  @override
  List<Object> get props => [qr, imageFile];
}

class QrUpdate extends QrEvent {
  final String guid;
  final QrModel qrModel;

  const QrUpdate({
    required this.guid,
    required this.qrModel,
  });

  @override
  List<Object> get props => [qrModel];
}

class QrWithImageUpdate extends QrEvent {
  final String guid;
  final QrModel qr;
  final File imageFile;
  final Uint8List imageWeb;

  const QrWithImageUpdate({
    required this.guid,
    required this.qr,
    required this.imageFile,
    required this.imageWeb,
  });

  @override
  List<Object> get props => [qr];
}
