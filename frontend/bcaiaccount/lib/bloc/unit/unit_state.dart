part of 'unit_bloc.dart';

abstract class UnitState extends Equatable {
  const UnitState();

  @override
  List<Object> get props => [];
}

class UnitInitial extends UnitState {}

class UnitInProgress extends UnitState {}

class UnitLoadSuccess extends UnitState {
  final List<UnitModel> units;

  const UnitLoadSuccess({required this.units});

  UnitLoadSuccess copyWith({
    List<UnitModel>? units,
  }) =>
      UnitLoadSuccess(units: units ?? this.units);

  @override
  List<Object> get props => [units];
}

class UnitLoadFailed extends UnitState {
  final String message;

  const UnitLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class UnitSaveInitial extends UnitState {}

class UnitSaveInProgress extends UnitState {}

class UnitSaveSuccess extends UnitState {}

class UnitSaveFailed extends UnitState {
  final String message;

  const UnitSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class UnitDeleteInProgress extends UnitState {}

class UnitDeleteSuccess extends UnitState {}

class UnitDeleteFailed extends UnitState {}

class UnitDeleteManyInProgress extends UnitState {}

class UnitDeleteManySuccess extends UnitState {}

class UnitDeleteManyFailed extends UnitState {}

class UnitGetInProgress extends UnitState {}

class UnitGetSuccess extends UnitState {
  final UnitModel unit;

  const UnitGetSuccess({required this.unit});

  UnitGetSuccess copyWith({
    UnitModel? unit,
  }) =>
      UnitGetSuccess(unit: unit ?? this.unit);

  @override
  List<Object> get props => [unit];
}

class UnitGetFailed extends UnitState {
  final String message;

  const UnitGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class UnitUpdateInitial extends UnitState {}

class UnitUpdateInProgress extends UnitState {}

class UnitUpdateSuccess extends UnitState {}

class UnitUpdateFailed extends UnitState {
  final String message;

  const UnitUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class UnitLoadFromMainShopInProgress extends UnitState {}

class UnitLoadFromMainShopSuccess extends UnitState {
  final List<UnitModel> units;

  const UnitLoadFromMainShopSuccess({required this.units});

  UnitLoadFromMainShopSuccess copyWith({
    List<UnitModel>? units,
  }) =>
      UnitLoadFromMainShopSuccess(units: units ?? this.units);

  @override
  List<Object> get props => [units];
}

class UnitLoadFromMainShopFailed extends UnitState {
  final String message;

  const UnitLoadFromMainShopFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class UnitUploadExcelInitial extends UnitState {}

class UnitUploadExcelInProgress extends UnitState {}

class UnitUploadExcelSuccess extends UnitState {
  final String uploadResult;

  const UnitUploadExcelSuccess({required this.uploadResult});

  @override
  List<Object> get props => [uploadResult];
}

class UnitUploadExcelFailed extends UnitState {
  final String message;

  const UnitUploadExcelFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
