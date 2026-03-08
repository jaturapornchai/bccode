part of 'wallet_pay_bloc.dart';

abstract class WalletPayState extends Equatable {
  const WalletPayState();

  @override
  List<Object> get props => [];
}

class WalletPayInitial extends WalletPayState {}

class WalletPayInProgress extends WalletPayState {}

class WalletPayLoadSuccess extends WalletPayState {
  final List<WalletModel> walletPays;

  const WalletPayLoadSuccess({required this.walletPays});

  WalletPayLoadSuccess copyWith({
    String guid = '',
    List<WalletModel>? walletPays,
  }) =>
      WalletPayLoadSuccess(walletPays: walletPays ?? this.walletPays);

  @override
  List<Object> get props => [walletPays];
}

class WalletPayLoadFailed extends WalletPayState {
  final String message;

  const WalletPayLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class WalletPaySaveInitial extends WalletPayState {}

class WalletPaySaveInProgress extends WalletPayState {}

class WalletPaySaveSuccess extends WalletPayState {}

class WalletPaySaveFailed extends WalletPayState {
  final String message;

  const WalletPaySaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class WalletPayDeleteInProgress extends WalletPayState {}

class WalletPayDeleteSuccess extends WalletPayState {}

class WalletPayDeleteFailed extends WalletPayState {}

class WalletPayDeleteManyInProgress extends WalletPayState {}

class WalletPayDeleteManySuccess extends WalletPayState {}

class WalletPayDeleteManyFailed extends WalletPayState {}

class WalletPayGetInProgress extends WalletPayState {}

class WalletPayGetSuccess extends WalletPayState {
  final WalletModel walletPays;

  const WalletPayGetSuccess({required this.walletPays});

  WalletPayGetSuccess copyWith({
    WalletModel? walletPays,
  }) =>
      WalletPayGetSuccess(walletPays: walletPays ?? this.walletPays);

  @override
  List<Object> get props => [walletPays];
}

class WalletPayGetFailed extends WalletPayState {
  final String message;

  const WalletPayGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class WalletPayUpdateInitial extends WalletPayState {}

class WalletPayUpdateInProgress extends WalletPayState {}

class WalletPayUpdateSuccess extends WalletPayState {}

class WalletPayUpdateFailed extends WalletPayState {
  final String message;

  const WalletPayUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
