part of 'book_bank_bloc.dart';

abstract class BookBankState extends Equatable {
  const BookBankState();

  @override
  List<Object> get props => [];
}

class BookBankInitial extends BookBankState {}

class BookBankInProgress extends BookBankState {}

class BookBankLoadSuccess extends BookBankState {
  final List<BookBankModel> bookBanks;

  const BookBankLoadSuccess({required this.bookBanks});

  BookBankLoadSuccess copyWith({
    List<BookBankModel>? bookBanks,
  }) =>
      BookBankLoadSuccess(bookBanks: bookBanks ?? this.bookBanks);

  @override
  List<Object> get props => [bookBanks];
}

class BookBankLoadFailed extends BookBankState {
  final String message;

  const BookBankLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class BookBankSaveInitial extends BookBankState {}

class BookBankSaveInProgress extends BookBankState {}

class BookBankSaveSuccess extends BookBankState {}

class BookBankSaveFailed extends BookBankState {
  final String message;

  const BookBankSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class BookBankDeleteInProgress extends BookBankState {}

class BookBankDeleteSuccess extends BookBankState {}

class BookBankDeleteFailed extends BookBankState {}

class BookBankDeleteManyInProgress extends BookBankState {}

class BookBankDeleteManySuccess extends BookBankState {}

class BookBankDeleteManyFailed extends BookBankState {}

class BookBankGetInProgress extends BookBankState {}

class BookBankGetSuccess extends BookBankState {
  final BookBankModel bookBank;

  const BookBankGetSuccess({required this.bookBank});

  BookBankGetSuccess copyWith({
    BookBankModel? bookBank,
  }) =>
      BookBankGetSuccess(bookBank: bookBank ?? this.bookBank);

  @override
  List<Object> get props => [bookBank];
}

class BookBankGetFailed extends BookBankState {
  final String message;

  const BookBankGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class BookBankUpdateInitial extends BookBankState {}

class BookBankUpdateInProgress extends BookBankState {}

class BookBankUpdateSuccess extends BookBankState {}

class BookBankUpdateFailed extends BookBankState {
  final String message;

  const BookBankUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
