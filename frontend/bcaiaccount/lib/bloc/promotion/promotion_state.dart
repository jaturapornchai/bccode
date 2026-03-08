part of 'promotion_bloc.dart';

abstract class PromotionState extends Equatable {
  const PromotionState();

  @override
  List<Object> get props => [];
}

class PromotionInitial extends PromotionState {}

class PromotionInProgress extends PromotionState {}

class PromotionLoadSuccess extends PromotionState {
  final List<PromotionModel> promotions;

  const PromotionLoadSuccess({required this.promotions});

  PromotionLoadSuccess copyWith({
    List<PromotionModel>? promotions,
  }) =>
      PromotionLoadSuccess(promotions: promotions ?? this.promotions);

  @override
  List<Object> get props => [promotions];
}

class PromotionLoadFailed extends PromotionState {
  final String message;

  const PromotionLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class PromotionSaveInitial extends PromotionState {}

class PromotionSaveInProgress extends PromotionState {}

class PromotionSaveSuccess extends PromotionState {}

class PromotionSaveFailed extends PromotionState {
  final String message;

  const PromotionSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class PromotionDeleteInProgress extends PromotionState {}

class PromotionDeleteSuccess extends PromotionState {}

class PromotionDeleteFailed extends PromotionState {
  final String message;

  const PromotionDeleteFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class PromotionDeleteManyInProgress extends PromotionState {}

class PromotionDeleteManySuccess extends PromotionState {}

class PromotionDeleteManyFailed extends PromotionState {
  final String message;

  const PromotionDeleteManyFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class PromotionGetInProgress extends PromotionState {}

class PromotionGetSuccess extends PromotionState {
  final PromotionModel promotions;

  const PromotionGetSuccess({required this.promotions});

  PromotionGetSuccess copyWith({
    PromotionModel? promotions,
  }) =>
      PromotionGetSuccess(promotions: promotions ?? this.promotions);

  @override
  List<Object> get props => [promotions];
}

class PromotionGetFailed extends PromotionState {
  final String message;

  const PromotionGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class PromotionUpdateInitial extends PromotionState {}

class PromotionUpdateInProgress extends PromotionState {}

class PromotionUpdateSuccess extends PromotionState {}

class PromotionUpdateFailed extends PromotionState {
  final String message;

  const PromotionUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
