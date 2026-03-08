part of 'creditor_bloc.dart';

abstract class CreditorEvent extends Equatable {
  const CreditorEvent();

  @override
  List<Object> get props => [];
}

class CreditorGet extends CreditorEvent {
  final String guid;

  const CreditorGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class CreditorLoadList extends CreditorEvent {
  final int limit;
  final int offset;
  final String search;
  final List<String>? groups;

  const CreditorLoadList(
      {required this.offset,
      required this.limit,
      required this.search,
      required this.groups});

  @override
  List<Object> get props => [];
}

class CreditorDelete extends CreditorEvent {
  final String guid;

  const CreditorDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class CreditorDeleteMany extends CreditorEvent {
  final List<String> guid;

  const CreditorDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class CreditorSave extends CreditorEvent {
  final CreditorModel creditor;

  const CreditorSave({
    required this.creditor,
  });

  @override
  List<Object> get props => [creditor];
}

class CreditorWithImageSave extends CreditorEvent {
  final List<File> imageFile;
  final CreditorModel creditor;
  final List<Uint8List> imageWeb;
  const CreditorWithImageSave({
    required this.imageWeb,
    required this.imageFile,
    required this.creditor,
  });

  @override
  List<Object> get props => [creditor, imageFile];
}

class CreditorUpdate extends CreditorEvent {
  final String guid;
  final CreditorModel creditorModel;

  const CreditorUpdate({
    required this.guid,
    required this.creditorModel,
  });

  @override
  List<Object> get props => [creditorModel];
}

class CreditorWithImageUpdate extends CreditorEvent {
  final String guid;
  final CreditorModel creditor;
  final List<File> imageFiles;
  final List<Uint8List> imageWeb;
  final List<ImagesModel> imagesUris;

  const CreditorWithImageUpdate({
    required this.guid,
    required this.creditor,
    required this.imageFiles,
    required this.imageWeb,
    required this.imagesUris,
  });

  @override
  List<Object> get props => [creditor];
}

class CreditorGetBycode extends CreditorEvent {
  final String custcode;

  const CreditorGetBycode({required this.custcode});

  @override
  List<Object> get props => [custcode];
}

class CreditorBulkSave extends CreditorEvent {
  final List<CreditorRequestModel> creditors;

  const CreditorBulkSave({
    required this.creditors,
  });

  @override
  List<Object> get props => [creditors];
}
