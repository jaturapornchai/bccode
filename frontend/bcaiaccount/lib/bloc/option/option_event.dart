part of 'option_bloc.dart';

abstract class OptionEvent extends Equatable {
  const OptionEvent();

  @override
  List<Object> get props => [];
}

class OptionLoad extends OptionEvent {
  const OptionLoad();

  @override
  List<Object> get props => [];
}

// ignore: must_be_immutable
class ListOptionLoad extends OptionEvent {
  int page;
  int perPage;
  String search;
  bool nextPage;

  ListOptionLoad({
    required this.page,
    required this.perPage,
    required this.search,
    required this.nextPage,
  });

  @override
  List<Object> get props => [];
}

// ignore: must_be_immutable
class ListOptionLoadById extends OptionEvent {
  String id;
  ListOptionLoadById({required this.id});

  @override
  List<Object> get props => [];
}

class ListOptionLoadSelect extends OptionEvent {
  const ListOptionLoadSelect();

  @override
  List<Object> get props => [];
}

class OptionSaved extends OptionEvent {
  final OptionModel option;

  const OptionSaved({
    required this.option,
  });

  @override
  List<Object> get props => [option];
}

class OptionUpdate extends OptionEvent {
  final OptionModel option;

  const OptionUpdate({
    required this.option,
  });

  @override
  List<Object> get props => [option];
}

class OptionDelete extends OptionEvent {
  final String id;

  const OptionDelete({
    required this.id,
  });

  @override
  List<Object> get props => [id];
}
