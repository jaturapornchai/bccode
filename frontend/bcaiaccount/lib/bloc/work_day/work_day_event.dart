part of 'work_day_bloc.dart';

abstract class WorkDayEvent extends Equatable {
  const WorkDayEvent();

  @override
  List<Object> get props => [];
}

class WorkDayGet extends WorkDayEvent {
  final String guid;

  const WorkDayGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class WorkDayLoad extends WorkDayEvent {
  const WorkDayLoad();

  @override
  List<Object> get props => [];
}

class WorkDayDelete extends WorkDayEvent {
  final String guid;

  const WorkDayDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class WorkDayDeleteMany extends WorkDayEvent {
  final List<String> guid;

  const WorkDayDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class WorkDaySave extends WorkDayEvent {
  final WorkDayListModel workDays;

  const WorkDaySave({
    required this.workDays,
  });

  @override
  List<Object> get props => [workDays];
}

class WorkDayUpdate extends WorkDayEvent {
  final String guid;
  final WorkDayListModel workDays;

  const WorkDayUpdate({
    required this.guid,
    required this.workDays,
  });

  @override
  List<Object> get props => [workDays];
}
