part of 'line_notify_bloc.dart';

abstract class LineNotifyState extends Equatable {
  const LineNotifyState();

  @override
  List<Object> get props => [];
}

class LineNotifyInitial extends LineNotifyState {}

class LineNotifyInProgress extends LineNotifyState {}

class LineNotifyLoadSuccess extends LineNotifyState {
  final List<LineNotifyModel> lineNotifys;

  const LineNotifyLoadSuccess({required this.lineNotifys});

  LineNotifyLoadSuccess copyWith({
    List<LineNotifyModel>? lineNotifys,
  }) =>
      LineNotifyLoadSuccess(lineNotifys: lineNotifys ?? this.lineNotifys);

  @override
  List<Object> get props => [lineNotifys];
}

class LineNotifyLoadFailed extends LineNotifyState {
  final String message;

  const LineNotifyLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class LineNotifySaveInitial extends LineNotifyState {}

class LineNotifySaveInProgress extends LineNotifyState {}

class LineNotifySaveSuccess extends LineNotifyState {}

class LineNotifySaveFailed extends LineNotifyState {
  final String message;

  const LineNotifySaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class LineNotifyDeleteInProgress extends LineNotifyState {}

class LineNotifyDeleteSuccess extends LineNotifyState {}

class LineNotifyDeleteFailed extends LineNotifyState {
  final String message;

  const LineNotifyDeleteFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class LineNotifyDeleteManyInProgress extends LineNotifyState {}

class LineNotifyDeleteManySuccess extends LineNotifyState {}

class LineNotifyDeleteManyFailed extends LineNotifyState {}

class LineNotifyGetInProgress extends LineNotifyState {}

class LineNotifyGetSuccess extends LineNotifyState {
  final LineNotifyModel lineNotify;

  const LineNotifyGetSuccess({required this.lineNotify});

  LineNotifyGetSuccess copyWith({
    LineNotifyModel? lineNotify,
  }) =>
      LineNotifyGetSuccess(lineNotify: lineNotify ?? this.lineNotify);

  @override
  List<Object> get props => [lineNotify];
}

class LineNotifyGetFailed extends LineNotifyState {
  final String message;

  const LineNotifyGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class LineNotifyUpdateInitial extends LineNotifyState {}

class LineNotifyUpdateInProgress extends LineNotifyState {}

class LineNotifyUpdateSuccess extends LineNotifyState {}

class LineNotifyUpdateFailed extends LineNotifyState {
  final String message;

  const LineNotifyUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class LineNotifyTestInitial extends LineNotifyState {}

class LineNotifyTestInProgress extends LineNotifyState {}

class LineNotifyTestSuccess extends LineNotifyState {}

class LineNotifyTestFailed extends LineNotifyState {
  final String message;

  const LineNotifyTestFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
