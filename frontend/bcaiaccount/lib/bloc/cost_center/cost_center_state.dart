part of 'cost_center_bloc.dart';

abstract class CostCenterState extends Equatable {
  const CostCenterState();

  @override
  List<Object> get props => [];
}

class CostCenterInitial extends CostCenterState {}

class CostCenterInProgress extends CostCenterState {}

class CostCenterLoadSuccess extends CostCenterState {
  final List<CostCenterModel> costCenter;

  const CostCenterLoadSuccess({required this.costCenter});

  CostCenterLoadSuccess copyWith({
    List<CostCenterModel>? costCenter,
  }) =>
      CostCenterLoadSuccess(costCenter: costCenter ?? this.costCenter);

  @override
  List<Object> get props => [costCenter];
}

class CostCenterLoadFailed extends CostCenterState {
  final String message;

  const CostCenterLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CostCenterSaveInitial extends CostCenterState {}

class CostCenterSaveInProgress extends CostCenterState {}

class CostCenterSaveSuccess extends CostCenterState {
  final String responsesID;

  const CostCenterSaveSuccess({
    required this.responsesID,
  });

  @override
  List<Object> get props => [responsesID];
}

class CostCenterSaveFailed extends CostCenterState {
  final String message;

  const CostCenterSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CostCenterDeleteInProgress extends CostCenterState {}

class CostCenterDeleteSuccess extends CostCenterState {}

class CostCenterDeleteFailed extends CostCenterState {}

class CostCenterDeleteManyInProgress extends CostCenterState {}

class CostCenterDeleteManySuccess extends CostCenterState {}

class CostCenterDeleteManyFailed extends CostCenterState {}

class CostCenterGetInProgress extends CostCenterState {}

class CostCenterGetSuccess extends CostCenterState {
  final CostCenterModel costCenter;

  const CostCenterGetSuccess({required this.costCenter});

  CostCenterGetSuccess copyWith({
    CostCenterModel? costCenter,
  }) =>
      CostCenterGetSuccess(costCenter: costCenter ?? this.costCenter);

  @override
  List<Object> get props => [costCenter];
}

class CostCenterGetFailed extends CostCenterState {
  final String message;

  const CostCenterGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CostCenterUpdateInitial extends CostCenterState {}

class CostCenterUpdateInProgress extends CostCenterState {}

class CostCenterUpdateSuccess extends CostCenterState {}

class CostCenterUpdateFailed extends CostCenterState {
  final String message;

  const CostCenterUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
