part of 'bank_bloc.dart';

abstract class BankState extends Equatable {
  const BankState();

  @override
  List<Object> get props => [];
}

class BankInitial extends BankState {}

class BankInProgress extends BankState {}

class BankLoadSuccess extends BankState {
  final List<BankModel> banks;

  const BankLoadSuccess({required this.banks});

  BankLoadSuccess copyWith({
    List<BankModel>? banks,
  }) =>
      BankLoadSuccess(banks: banks ?? this.banks);

  @override
  List<Object> get props => [banks];
}

class BankLoadFailed extends BankState {
  final String message;

  const BankLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class BankSaveInitial extends BankState {}

class BankSaveInProgress extends BankState {}

class BankSaveSuccess extends BankState {}

class BankSaveFailed extends BankState {
  final String message;

  const BankSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class BankDeleteInProgress extends BankState {}

class BankDeleteSuccess extends BankState {}

class BankDeleteFailed extends BankState {}

class BankDeleteManyInProgress extends BankState {}

class BankDeleteManySuccess extends BankState {}

class BankDeleteManyFailed extends BankState {}

class BankGetInProgress extends BankState {}

class BankGetSuccess extends BankState {
  final BankModel bank;

  const BankGetSuccess({required this.bank});

  BankGetSuccess copyWith({
    BankModel? bank,
  }) =>
      BankGetSuccess(bank: bank ?? this.bank);

  @override
  List<Object> get props => [bank];
}

class BankGetFailed extends BankState {
  final String message;

  const BankGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class BankUpdateInitial extends BankState {}

class BankUpdateInProgress extends BankState {}

class BankUpdateSuccess extends BankState {}

class BankUpdateFailed extends BankState {
  final String message;

  const BankUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
