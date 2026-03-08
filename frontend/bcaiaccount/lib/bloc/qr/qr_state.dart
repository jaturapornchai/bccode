part of 'qr_bloc.dart';

abstract class QrState extends Equatable {
  const QrState();

  @override
  List<Object> get props => [];
}

class QrInitial extends QrState {}

class QrInProgress extends QrState {}

class QrLoadSuccess extends QrState {
  final List<QrModel> qrs;

  const QrLoadSuccess({required this.qrs});

  QrLoadSuccess copyWith({
    List<QrModel>? qrs,
  }) =>
      QrLoadSuccess(qrs: qrs ?? this.qrs);

  @override
  List<Object> get props => [qrs];
}

class QrLoadFailed extends QrState {
  final String message;

  const QrLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class QrSaveInitial extends QrState {}

class QrSaveInProgress extends QrState {}

class QrSaveSuccess extends QrState {}

class QrSaveFailed extends QrState {
  final String message;

  const QrSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class QrDeleteInProgress extends QrState {}

class QrDeleteSuccess extends QrState {}

class QrDeleteFailed extends QrState {}

class QrDeleteManyInProgress extends QrState {}

class QrDeleteManySuccess extends QrState {}

class QrDeleteManyFailed extends QrState {}

class QrGetInProgress extends QrState {}

class QrGetSuccess extends QrState {
  final QrModel qrs;

  const QrGetSuccess({required this.qrs});

  QrGetSuccess copyWith({
    QrModel? qrs,
  }) =>
      QrGetSuccess(qrs: qrs ?? this.qrs);

  @override
  List<Object> get props => [qrs];
}

class QrGetFailed extends QrState {
  final String message;

  const QrGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class QrUpdateInitial extends QrState {}

class QrUpdateInProgress extends QrState {}

class QrUpdateSuccess extends QrState {}

class QrUpdateFailed extends QrState {
  final String message;

  const QrUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
