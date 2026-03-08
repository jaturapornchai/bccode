part of 'bank_bloc.dart';

abstract class BankEvent extends Equatable {
  const BankEvent();

  @override
  List<Object> get props => [];
}

class BankGet extends BankEvent {
  final String guid;

  const BankGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class BankLoadList extends BankEvent {
  final int limit;
  final int offset;
  final String search;

  const BankLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class BankDelete extends BankEvent {
  final String guid;

  const BankDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class BankDeleteMany extends BankEvent {
  final List<String> guid;

  const BankDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class BankSave extends BankEvent {
  final BankModel bank;

  const BankSave({
    required this.bank,
  });

  @override
  List<Object> get props => [bank];
}

class BankBulkSave extends BankEvent {
  final List<BankModel> banks;

  const BankBulkSave({
    required this.banks,
  });

  @override
  List<Object> get props => [banks];
}

class BankSaveWithImage extends BankEvent {
  final BankModel bank;
  final File imageFile;
  final Uint8List imageWeb;

  const BankSaveWithImage({
    required this.imageWeb,
    required this.imageFile,
    required this.bank,
  });

  @override
  List<Object> get props => [bank, imageFile];
}

class BankUpdate extends BankEvent {
  final String guid;
  final BankModel bankModel;

  const BankUpdate({
    required this.guid,
    required this.bankModel,
  });

  @override
  List<Object> get props => [bankModel];
}

class BankWithImageUpdate extends BankEvent {
  final String guid;
  final BankModel bank;
  final File imageFile;
  final Uint8List imageWeb;
  final ImagesModel imagesUri;

  const BankWithImageUpdate({
    required this.guid,
    required this.bank,
    required this.imageFile,
    required this.imageWeb,
    required this.imagesUri,
  });

  @override
  List<Object> get props => [bank];
}
