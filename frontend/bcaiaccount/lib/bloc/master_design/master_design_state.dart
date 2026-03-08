part of 'master_design_bloc.dart';

abstract class MasterDesignState extends Equatable {
  const MasterDesignState();

  @override
  List<Object> get props => [];
}

class MasterDesignInitial extends MasterDesignState {}

class MasterDesignInProgress extends MasterDesignState {}

class MasterDesignLoadSuccess extends MasterDesignState {
  final List<MasterDesignModel> designs;

  const MasterDesignLoadSuccess({required this.designs});

  MasterDesignLoadSuccess copyWith({
    List<MasterDesignModel>? designs,
  }) =>
      MasterDesignLoadSuccess(designs: designs ?? this.designs);

  @override
  List<Object> get props => [designs];
}

class MasterDesignLoadFailed extends MasterDesignState {
  final String message;

  const MasterDesignLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterDesignSaveInitial extends MasterDesignState {}

class MasterDesignSaveInProgress extends MasterDesignState {}

class MasterDesignSaveSuccess extends MasterDesignState {}

class MasterDesignSaveFailed extends MasterDesignState {
  final String message;

  const MasterDesignSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterDesignDeleteInProgress extends MasterDesignState {}

class MasterDesignDeleteSuccess extends MasterDesignState {}

class MasterDesignDeleteFailed extends MasterDesignState {}

class MasterDesignDeleteManyInProgress extends MasterDesignState {}

class MasterDesignDeleteManySuccess extends MasterDesignState {}

class MasterDesignDeleteManyFailed extends MasterDesignState {}

class MasterDesignGetInProgress extends MasterDesignState {}

class MasterDesignGetSuccess extends MasterDesignState {
  final MasterDesignModel design;

  const MasterDesignGetSuccess({required this.design});

  MasterDesignGetSuccess copyWith({
    MasterDesignModel? design,
  }) =>
      MasterDesignGetSuccess(design: design ?? this.design);

  @override
  List<Object> get props => [design];
}

class MasterDesignGetFailed extends MasterDesignState {
  final String message;

  const MasterDesignGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterDesignUpdateInitial extends MasterDesignState {}

class MasterDesignUpdateInProgress extends MasterDesignState {}

class MasterDesignUpdateSuccess extends MasterDesignState {}

class MasterDesignUpdateFailed extends MasterDesignState {
  final String message;

  const MasterDesignUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
