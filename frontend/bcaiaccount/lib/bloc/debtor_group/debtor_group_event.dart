part of 'debtor_group_bloc.dart';

abstract class DebtorGroupEvent extends Equatable {
  const DebtorGroupEvent();

  @override
  List<Object> get props => [];
}

class DebtorGroupGet extends DebtorGroupEvent {
  final String guid;

  const DebtorGroupGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class DebtorGroupLoadList extends DebtorGroupEvent {
  final int limit;
  final int offset;
  final String search;

  const DebtorGroupLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class DebtorGroupDelete extends DebtorGroupEvent {
  final String guid;

  const DebtorGroupDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class DebtorGroupDeleteMany extends DebtorGroupEvent {
  final List<String> guid;

  const DebtorGroupDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class DebtorGroupSave extends DebtorGroupEvent {
  final DebtorGroupModel debtorGroups;

  const DebtorGroupSave({
    required this.debtorGroups,
  });

  @override
  List<Object> get props => [debtorGroups];
}

class DebtorGroupUpdate extends DebtorGroupEvent {
  final String guid;
  final DebtorGroupModel debtorGroups;

  const DebtorGroupUpdate({
    required this.guid,
    required this.debtorGroups,
  });

  @override
  List<Object> get props => [debtorGroups];
}
