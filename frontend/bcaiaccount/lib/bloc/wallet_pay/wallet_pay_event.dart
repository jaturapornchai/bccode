part of 'wallet_pay_bloc.dart';

abstract class WalletPayEvent extends Equatable {
  const WalletPayEvent();

  @override
  List<Object> get props => [];
}

class WalletPayGet extends WalletPayEvent {
  final String guid;

  const WalletPayGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class WalletPayLoadList extends WalletPayEvent {
  final int limit;
  final int offset;
  final String search;

  const WalletPayLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class WalletPayDelete extends WalletPayEvent {
  final String guid;

  const WalletPayDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class WalletPayDeleteMany extends WalletPayEvent {
  final List<String> guid;

  const WalletPayDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class WalletPaySave extends WalletPayEvent {
  final WalletModel walletModel;

  const WalletPaySave({
    required this.walletModel,
  });

  @override
  List<Object> get props => [walletModel];
}

class WalletPayUpdate extends WalletPayEvent {
  final String guid;
  final WalletModel walletModel;

  const WalletPayUpdate({
    required this.guid,
    required this.walletModel,
  });

  @override
  List<Object> get props => [WalletModel];
}
