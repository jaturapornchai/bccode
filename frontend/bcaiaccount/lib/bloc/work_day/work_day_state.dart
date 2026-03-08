part of 'work_day_bloc.dart';

abstract class WorkDayState extends Equatable {
  const WorkDayState();

  @override
  List<Object> get props => [];
}

class WorkDayInitial extends WorkDayState {}

class WorkDayInProgress extends WorkDayState {}

class WorkDayLoadSuccess extends WorkDayState {
  final String guidfixed;
  final List<WorkDayModel> workDay;

  const WorkDayLoadSuccess({required this.guidfixed, required this.workDay});

  @override
  List<Object> get props => [guidfixed, workDay];
}

class WorkDayLoadFailed extends WorkDayState {
  final String message;

  const WorkDayLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class WorkDaySaveInitial extends WorkDayState {}

class WorkDaySaveInProgress extends WorkDayState {}

class WorkDaySaveSuccess extends WorkDayState {}

class WorkDaySaveFailed extends WorkDayState {
  final String message;

  const WorkDaySaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class WorkDayDeleteInProgress extends WorkDayState {}

class WorkDayDeleteSuccess extends WorkDayState {}

class WorkDayDeleteFailed extends WorkDayState {}

class WorkDayDeleteManyInProgress extends WorkDayState {}

class WorkDayDeleteManySuccess extends WorkDayState {}

class WorkDayDeleteManyFailed extends WorkDayState {}

class WorkDayGetInProgress extends WorkDayState {}

class WorkDayGetSuccess extends WorkDayState {
  final WorkDayModel workDay;

  const WorkDayGetSuccess({required this.workDay});

  WorkDayGetSuccess copyWith({
    WorkDayModel? workDay,
  }) =>
      WorkDayGetSuccess(workDay: workDay ?? this.workDay);

  @override
  List<Object> get props => [WorkDayModel];
}

class WorkDayGetFailed extends WorkDayState {
  final String message;

  const WorkDayGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class WorkDayUpdateInitial extends WorkDayState {}

class WorkDayUpdateInProgress extends WorkDayState {}

class WorkDayUpdateSuccess extends WorkDayState {}

class WorkDayUpdateFailed extends WorkDayState {
  final String message;

  const WorkDayUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
