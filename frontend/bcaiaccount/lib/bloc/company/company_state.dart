part of 'company_bloc.dart';

abstract class CompanyState extends Equatable {
  const CompanyState();

  @override
  List<Object> get props => [];
}

class CompanyInitial extends CompanyState {}

class CompanyInProgress extends CompanyState {}

class CompanyLoadSuccess extends CompanyState {
  final String guidFixed;
  final CompanyModel company;

  const CompanyLoadSuccess({required this.guidFixed, required this.company});

  @override
  List<Object> get props => [guidFixed, company];
}

class CompanyLoadNotFound extends CompanyState {
  final String message;
  final String shopdefault;

  const CompanyLoadNotFound({
    required this.message,
    required this.shopdefault,
  });

  @override
  List<Object> get props => [message, shopdefault];
}

class CompanyLoadFailed extends CompanyState {
  final String message;

  const CompanyLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CompanySaveInitial extends CompanyState {}

class CompanySaveInProgress extends CompanyState {}

class CompanySaveSuccess extends CompanyState {}

class CompanySaveFailed extends CompanyState {
  final String message;

  const CompanySaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CompanyDeleteInProgress extends CompanyState {}

class CompanyDeleteSuccess extends CompanyState {}

class CompanyDeleteFailed extends CompanyState {}

class CompanyDeleteManyInProgress extends CompanyState {}

class CompanyDeleteManySuccess extends CompanyState {}

class CompanyDeleteManyFailed extends CompanyState {}

class CompanyGetInProgress extends CompanyState {}

class CompanyGetSuccess extends CompanyState {
  final CompanyModel company;

  const CompanyGetSuccess({required this.company});

  CompanyGetSuccess copyWith({
    CompanyModel? company,
  }) =>
      CompanyGetSuccess(company: company ?? this.company);

  @override
  List<Object> get props => [company];
}

class CompanyGetFailed extends CompanyState {
  final String message;

  const CompanyGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CompanyUpdateInitial extends CompanyState {}

class CompanyUpdateInProgress extends CompanyState {}

class CompanyUpdateSuccess extends CompanyState {}

class CompanyUpdateFailed extends CompanyState {
  final String message;

  const CompanyUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
