part of 'holiday_bloc.dart';

abstract class HolidayState extends Equatable {
  const HolidayState();

  @override
  List<Object> get props => [];
}

class HolidayInitial extends HolidayState {}

class HolidayInProgress extends HolidayState {}

class HolidayLoadSuccess extends HolidayState {
  final List<HolidayModel> holidays;

  const HolidayLoadSuccess({required this.holidays});

  HolidayLoadSuccess copyWith({
    String guid = '',
    List<HolidayModel>? holidays,
  }) =>
      HolidayLoadSuccess(holidays: holidays ?? this.holidays);

  @override
  List<Object> get props => [holidays];
}

class HolidayLoadFailed extends HolidayState {
  final String message;

  const HolidayLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class HolidaySaveInitial extends HolidayState {}

class HolidaySaveInProgress extends HolidayState {}

class HolidaySaveSuccess extends HolidayState {}

class HolidaySaveFailed extends HolidayState {
  final String message;

  const HolidaySaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class HolidayDeleteInProgress extends HolidayState {}

class HolidayDeleteSuccess extends HolidayState {}

class HolidayDeleteFailed extends HolidayState {}

class HolidayDeleteManyInProgress extends HolidayState {}

class HolidayDeleteManySuccess extends HolidayState {}

class HolidayDeleteManyFailed extends HolidayState {}

class HolidayGetInProgress extends HolidayState {}

class HolidayGetSuccess extends HolidayState {
  final HolidayModel holiday;

  const HolidayGetSuccess({required this.holiday});

  HolidayGetSuccess copyWith({
    HolidayModel? holiday,
  }) =>
      HolidayGetSuccess(holiday: holiday ?? this.holiday);

  @override
  List<Object> get props => [holiday];
}

class HolidayGetFailed extends HolidayState {
  final String message;

  const HolidayGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class HolidayUpdateInitial extends HolidayState {}

class HolidayUpdateInProgress extends HolidayState {}

class HolidayUpdateSuccess extends HolidayState {}

class HolidayUpdateFailed extends HolidayState {
  final String message;

  const HolidayUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
