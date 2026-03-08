part of 'staff_bloc.dart';

abstract class StaffState extends Equatable {
  const StaffState();

  @override
  List<Object> get props => [];
}

class StaffInitial extends StaffState {}

class StaffInProgress extends StaffState {}

class StaffLoadSuccess extends StaffState {
  final List<StaffModel> staffs;

  const StaffLoadSuccess({required this.staffs});

  StaffLoadSuccess copyWith({
    List<StaffModel>? staffs,
  }) =>
      StaffLoadSuccess(staffs: staffs ?? this.staffs);

  @override
  List<Object> get props => [staffs];
}

class StaffLoadFailed extends StaffState {
  final String message;

  const StaffLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class StaffSaveInitial extends StaffState {}

class StaffSaveInProgress extends StaffState {}

class StaffSaveSuccess extends StaffState {}

class StaffSaveFailed extends StaffState {
  final String message;

  const StaffSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class StaffDeleteInProgress extends StaffState {}

class StaffDeleteSuccess extends StaffState {}

class StaffDeleteFailed extends StaffState {}

class StaffDeleteManyInProgress extends StaffState {}

class StaffDeleteManySuccess extends StaffState {}

class StaffDeleteManyFailed extends StaffState {}

class StaffGetInProgress extends StaffState {}

class StaffGetSuccess extends StaffState {
  final StaffModel staff;

  const StaffGetSuccess({required this.staff});

  StaffGetSuccess copyWith({
    StaffModel? staff,
  }) =>
      StaffGetSuccess(staff: staff ?? this.staff);

  @override
  List<Object> get props => [staff];
}

class StaffGetFailed extends StaffState {
  final String message;

  const StaffGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class StaffUpdateInitial extends StaffState {}

class StaffUpdateInProgress extends StaffState {}

class StaffUpdateSuccess extends StaffState {}

class StaffUpdateFailed extends StaffState {
  final String message;

  const StaffUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
