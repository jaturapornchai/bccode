part of 'sale_channel_bloc.dart';

abstract class SaleChannelState extends Equatable {
  const SaleChannelState();

  @override
  List<Object> get props => [];
}

class SaleChannelInitial extends SaleChannelState {}

class SaleChannelInProgress extends SaleChannelState {}

class SaleChannelLoadSuccess extends SaleChannelState {
  final List<SaleChannelModel> salechannel;

  const SaleChannelLoadSuccess({required this.salechannel});

  SaleChannelLoadSuccess copyWith({
    List<SaleChannelModel>? salechannel,
  }) =>
      SaleChannelLoadSuccess(salechannel: salechannel ?? this.salechannel);

  @override
  List<Object> get props => [salechannel];
}

class SaleChannelLoadFailed extends SaleChannelState {
  final String message;

  const SaleChannelLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class SaleChannelSaveInitial extends SaleChannelState {}

class SaleChannelSaveInProgress extends SaleChannelState {}

class SaleChannelSaveSuccess extends SaleChannelState {}

class SaleChannelSaveFailed extends SaleChannelState {
  final String message;

  const SaleChannelSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class SaleChannelDeleteInProgress extends SaleChannelState {}

class SaleChannelDeleteSuccess extends SaleChannelState {}

class SaleChannelDeleteFailed extends SaleChannelState {}

class SaleChannelDeleteManyInProgress extends SaleChannelState {}

class SaleChannelDeleteManySuccess extends SaleChannelState {}

class SaleChannelDeleteManyFailed extends SaleChannelState {}

class SaleChannelGetInProgress extends SaleChannelState {}

class SaleChannelGetSuccess extends SaleChannelState {
  final SaleChannelModel salechannel;

  const SaleChannelGetSuccess({required this.salechannel});

  SaleChannelGetSuccess copyWith({
    SaleChannelModel? salechannel,
  }) =>
      SaleChannelGetSuccess(salechannel: salechannel ?? this.salechannel);

  @override
  List<Object> get props => [salechannel];
}

class SaleChannelGetFailed extends SaleChannelState {
  final String message;

  const SaleChannelGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class SaleChannelUpdateInitial extends SaleChannelState {}

class SaleChannelUpdateInProgress extends SaleChannelState {}

class SaleChannelUpdateSuccess extends SaleChannelState {}

class SaleChannelUpdateFailed extends SaleChannelState {
  final String message;

  const SaleChannelUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
