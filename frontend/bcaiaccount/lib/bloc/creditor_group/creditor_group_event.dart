part of 'creditor_group_bloc.dart';

abstract class CreditorGroupEvent extends Equatable {
  const CreditorGroupEvent();

  @override
  List<Object> get props => [];
}

class CreditorGroupGet extends CreditorGroupEvent {
  final String guid;

  const CreditorGroupGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class CreditorGroupLoadList extends CreditorGroupEvent {
  final int limit;
  final int offset;
  final String search;

  const CreditorGroupLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class CreditorGroupDelete extends CreditorGroupEvent {
  final String guid;

  const CreditorGroupDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class CreditorGroupDeleteMany extends CreditorGroupEvent {
  final List<String> guid;

  const CreditorGroupDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class CreditorGroupSave extends CreditorGroupEvent {
  final CreditorGroupModel creditorGroup;

  const CreditorGroupSave({
    required this.creditorGroup,
  });

  @override
  List<Object> get props => [creditorGroup];
}

class CreditorGroupUpdate extends CreditorGroupEvent {
  final String guid;
  final CreditorGroupModel creditorGroup;

  const CreditorGroupUpdate({
    required this.guid,
    required this.creditorGroup,
  });

  @override
  List<Object> get props => [creditorGroup];
}
