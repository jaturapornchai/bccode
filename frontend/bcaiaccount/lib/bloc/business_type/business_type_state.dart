part of 'business_type_bloc.dart';

abstract class BusinessTypeState extends Equatable {
  const BusinessTypeState();

  @override
  List<Object> get props => [];
}

class BusinessTypeInitial extends BusinessTypeState {}

class BusinessTypeInProgress extends BusinessTypeState {}

class BusinessTypeLoadSuccess extends BusinessTypeState {
  final List<BusinessTypeModel> businessType;

  const BusinessTypeLoadSuccess({required this.businessType});

  BusinessTypeLoadSuccess copyWith({
    List<BusinessTypeModel>? businessType,
  }) =>
      BusinessTypeLoadSuccess(businessType: businessType ?? this.businessType);

  @override
  List<Object> get props => [businessType];
}

class BusinessTypeLoadFailed extends BusinessTypeState {
  final String message;

  const BusinessTypeLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class BusinessTypeSaveInitial extends BusinessTypeState {}

class BusinessTypeSaveInProgress extends BusinessTypeState {}

class BusinessTypeSaveSuccess extends BusinessTypeState {}

class BusinessTypeSaveFailed extends BusinessTypeState {
  final String message;

  const BusinessTypeSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class BusinessTypeDeleteInProgress extends BusinessTypeState {}

class BusinessTypeDeleteSuccess extends BusinessTypeState {}

class BusinessTypeDeleteFailed extends BusinessTypeState {}

class BusinessTypeDeleteManyInProgress extends BusinessTypeState {}

class BusinessTypeDeleteManySuccess extends BusinessTypeState {}

class BusinessTypeDeleteManyFailed extends BusinessTypeState {}

class BusinessTypeGetInProgress extends BusinessTypeState {}

class BusinessTypeGetSuccess extends BusinessTypeState {
  final BusinessTypeModel businessType;

  const BusinessTypeGetSuccess({required this.businessType});

  BusinessTypeGetSuccess copyWith({
    BusinessTypeModel? businessType,
  }) =>
      BusinessTypeGetSuccess(businessType: businessType ?? this.businessType);

  @override
  List<Object> get props => [businessType];
}

class BusinessTypeGetFailed extends BusinessTypeState {
  final String message;

  const BusinessTypeGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class BusinessTypeUpdateInitial extends BusinessTypeState {}

class BusinessTypeUpdateInProgress extends BusinessTypeState {}

class BusinessTypeUpdateSuccess extends BusinessTypeState {}

class BusinessTypeUpdateFailed extends BusinessTypeState {
  final String message;

  const BusinessTypeUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
