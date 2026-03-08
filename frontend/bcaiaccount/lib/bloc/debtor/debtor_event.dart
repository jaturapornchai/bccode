part of 'debtor_bloc.dart';

abstract class DebtorEvent extends Equatable {
  const DebtorEvent();

  @override
  List<Object> get props => [];
}

class DebtorGet extends DebtorEvent {
  final String guid;

  const DebtorGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class DebtorLoadList extends DebtorEvent {
  final int limit;
  final int offset;
  final String search;
  final List<String>? groups;

  const DebtorLoadList({
    required this.offset,
    required this.limit,
    required this.search,
    required this.groups,
  });

  @override
  List<Object> get props => [];
}

class DebtorDelete extends DebtorEvent {
  final String guid;

  const DebtorDelete({required this.guid});

  @override
  List<Object> get props => [guid];
}

class DebtorDeleteMany extends DebtorEvent {
  final List<String> guid;

  const DebtorDeleteMany({required this.guid});

  @override
  List<Object> get props => [guid];
}

class DebtorSave extends DebtorEvent {
  final DebtorModel debtor;

  const DebtorSave({required this.debtor});

  @override
  List<Object> get props => [debtor];
}

class DebtorWithImageSave extends DebtorEvent {
  final List<File> imageFile;
  final DebtorModel debtor;
  final List<Uint8List> imageWeb;
  const DebtorWithImageSave({
    required this.imageWeb,
    required this.imageFile,
    required this.debtor,
  });

  @override
  List<Object> get props => [debtor, imageFile];
}

class DebtorUpdate extends DebtorEvent {
  final String guid;
  final DebtorModel debtorModel;

  const DebtorUpdate({required this.guid, required this.debtorModel});

  @override
  List<Object> get props => [debtorModel];
}

class DebtorWithImageUpdate extends DebtorEvent {
  final String guid;
  final DebtorModel debtor;
  final List<File> imageFiles;
  final List<Uint8List> imageWeb;
  final List<ImagesModel> imagesUris;

  const DebtorWithImageUpdate({
    required this.guid,
    required this.debtor,
    required this.imageFiles,
    required this.imageWeb,
    required this.imagesUris,
  });

  @override
  List<Object> get props => [debtor];
}

class DebtorGetBycode extends DebtorEvent {
  final String custcode;

  const DebtorGetBycode({required this.custcode});

  @override
  List<Object> get props => [custcode];
}

class DebtorBulkSave extends DebtorEvent {
  final List<DebtorRequestModel> debtors;

  const DebtorBulkSave({required this.debtors});

  @override
  List<Object> get props => [debtors];
}

class DebtorAddPoint extends DebtorEvent {
  final String pointscode;
  final PointTransactionCreateModel pointTransaction;

  const DebtorAddPoint({
    required this.pointscode,
    required this.pointTransaction,
  });

  @override
  List<Object> get props => [pointscode, pointTransaction];
}

class DebtorBulkAddPoints extends DebtorEvent {
  final List<BulkPointsRequestModel> pointsList;

  const DebtorBulkAddPoints({required this.pointsList});

  @override
  List<Object> get props => [pointsList];
}
