part of 'book_bank_bloc.dart';

abstract class BookBankEvent extends Equatable {
  const BookBankEvent();

  @override
  List<Object> get props => [];
}

class BookBankGet extends BookBankEvent {
  final String guid;

  const BookBankGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class BookBankLoadList extends BookBankEvent {
  final int limit;
  final int offset;
  final String search;

  const BookBankLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class BookBankDelete extends BookBankEvent {
  final String guid;

  const BookBankDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class BookBankDeleteMany extends BookBankEvent {
  final List<String> guid;

  const BookBankDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class BookBankSave extends BookBankEvent {
  final BookBankModel bookBank;

  const BookBankSave({
    required this.bookBank,
  });

  @override
  List<Object> get props => [bookBank];
}

class BookBankWithImageSave extends BookBankEvent {
  final List<File> imageFile;
  final BookBankModel bookBank;
  final List<Uint8List> imageWeb;
  const BookBankWithImageSave({
    required this.imageWeb,
    required this.imageFile,
    required this.bookBank,
  });

  @override
  List<Object> get props => [bookBank, imageFile];
}

class BookBankUpdate extends BookBankEvent {
  final String guid;
  final BookBankModel bookBank;

  const BookBankUpdate({
    required this.guid,
    required this.bookBank,
  });

  @override
  List<Object> get props => [bookBank];
}

class BookBankWithImageUpdate extends BookBankEvent {
  final String guid;
  final BookBankModel bookBank;
  final List<File> imageFiles;
  final List<Uint8List> imageWeb;
  final List<ImagesModel> imagesUris;

  const BookBankWithImageUpdate({
    required this.guid,
    required this.bookBank,
    required this.imageFiles,
    required this.imageWeb,
    required this.imagesUris,
  });

  @override
  List<Object> get props => [bookBank];
}
