part of 'master_brand_bloc.dart';

abstract class MasterBrandState extends Equatable {
  const MasterBrandState();

  @override
  List<Object> get props => [];
}

class MasterBrandInitial extends MasterBrandState {}

class MasterBrandInProgress extends MasterBrandState {}

class MasterBrandLoadSuccess extends MasterBrandState {
  final List<MasterBrandModel> brands;

  const MasterBrandLoadSuccess({required this.brands});

  MasterBrandLoadSuccess copyWith({
    List<MasterBrandModel>? brands,
  }) =>
      MasterBrandLoadSuccess(brands: brands ?? this.brands);

  @override
  List<Object> get props => [brands];
}

class MasterBrandLoadFailed extends MasterBrandState {
  final String message;

  const MasterBrandLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterBrandSaveInitial extends MasterBrandState {}

class MasterBrandSaveInProgress extends MasterBrandState {}

class MasterBrandSaveSuccess extends MasterBrandState {}

class MasterBrandSaveFailed extends MasterBrandState {
  final String message;

  const MasterBrandSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterBrandDeleteInProgress extends MasterBrandState {}

class MasterBrandDeleteSuccess extends MasterBrandState {}

class MasterBrandDeleteFailed extends MasterBrandState {}

class MasterBrandDeleteManyInProgress extends MasterBrandState {}

class MasterBrandDeleteManySuccess extends MasterBrandState {}

class MasterBrandDeleteManyFailed extends MasterBrandState {}

class MasterBrandGetInProgress extends MasterBrandState {}

class MasterBrandGetSuccess extends MasterBrandState {
  final MasterBrandModel brand;

  const MasterBrandGetSuccess({required this.brand});

  MasterBrandGetSuccess copyWith({
    MasterBrandModel? brand,
  }) =>
      MasterBrandGetSuccess(brand: brand ?? this.brand);

  @override
  List<Object> get props => [brand];
}

class MasterBrandGetFailed extends MasterBrandState {
  final String message;

  const MasterBrandGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterBrandUpdateInitial extends MasterBrandState {}

class MasterBrandUpdateInProgress extends MasterBrandState {}

class MasterBrandUpdateSuccess extends MasterBrandState {}

class MasterBrandUpdateFailed extends MasterBrandState {
  final String message;

  const MasterBrandUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
