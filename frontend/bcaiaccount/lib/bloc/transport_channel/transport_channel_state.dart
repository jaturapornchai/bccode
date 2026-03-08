part of 'transport_channel_bloc.dart';

abstract class TransportChannelState extends Equatable {
  const TransportChannelState();

  @override
  List<Object> get props => [];
}

class TransportChannelInitial extends TransportChannelState {}

class TransportChannelInProgress extends TransportChannelState {}

class TransportChannelLoadSuccess extends TransportChannelState {
  final List<TransportChannelModel> transportchannel;

  const TransportChannelLoadSuccess({required this.transportchannel});

  TransportChannelLoadSuccess copyWith({
    List<TransportChannelModel>? transportchannel,
  }) =>
      TransportChannelLoadSuccess(
          transportchannel: transportchannel ?? this.transportchannel);

  @override
  List<Object> get props => [transportchannel];
}

class TransportChannelLoadFailed extends TransportChannelState {
  final String message;

  const TransportChannelLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransportChannelSaveInitial extends TransportChannelState {}

class TransportChannelSaveInProgress extends TransportChannelState {}

class TransportChannelSaveSuccess extends TransportChannelState {}

class TransportChannelSaveFailed extends TransportChannelState {
  final String message;

  const TransportChannelSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransportChannelDeleteInProgress extends TransportChannelState {}

class TransportChannelDeleteSuccess extends TransportChannelState {}

class TransportChannelDeleteFailed extends TransportChannelState {}

class TransportChannelDeleteManyInProgress extends TransportChannelState {}

class TransportChannelDeleteManySuccess extends TransportChannelState {}

class TransportChannelDeleteManyFailed extends TransportChannelState {}

class TransportChannelGetInProgress extends TransportChannelState {}

class TransportChannelGetSuccess extends TransportChannelState {
  final TransportChannelModel transportchannel;

  const TransportChannelGetSuccess({required this.transportchannel});

  TransportChannelGetSuccess copyWith({
    TransportChannelModel? transportchannel,
  }) =>
      TransportChannelGetSuccess(
          transportchannel: transportchannel ?? this.transportchannel);

  @override
  List<Object> get props => [transportchannel];
}

class TransportChannelGetFailed extends TransportChannelState {
  final String message;

  const TransportChannelGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class TransportChannelUpdateInitial extends TransportChannelState {}

class TransportChannelUpdateInProgress extends TransportChannelState {}

class TransportChannelUpdateSuccess extends TransportChannelState {}

class TransportChannelUpdateFailed extends TransportChannelState {
  final String message;

  const TransportChannelUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
