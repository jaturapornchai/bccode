part of 'company_branch_bloc.dart';

abstract class CompanyBranchState extends Equatable {
  const CompanyBranchState();

  @override
  List<Object> get props => [];
}

class CompanyBranchInitial extends CompanyBranchState {}

class CompanyBranchInProgress extends CompanyBranchState {}

class CompanyBranchLoadSuccess extends CompanyBranchState {
  final List<CompanyBranchModel> companyBranch;

  const CompanyBranchLoadSuccess({required this.companyBranch});

  CompanyBranchLoadSuccess copyWith({
    List<CompanyBranchModel>? companyBranch,
  }) =>
      CompanyBranchLoadSuccess(
          companyBranch: companyBranch ?? this.companyBranch);

  @override
  List<Object> get props => [companyBranch];
}

class CompanyBranchLoadFailed extends CompanyBranchState {
  final String message;

  const CompanyBranchLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CompanyBranchSaveInitial extends CompanyBranchState {}

class CompanyBranchSaveInProgress extends CompanyBranchState {}

class CompanyBranchSaveSuccess extends CompanyBranchState {}

class CompanyBranchSaveFailed extends CompanyBranchState {
  final String message;

  const CompanyBranchSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CompanyBranchDeleteInProgress extends CompanyBranchState {}

class CompanyBranchDeleteSuccess extends CompanyBranchState {}

class CompanyBranchDeleteFailed extends CompanyBranchState {}

class CompanyBranchDeleteManyInProgress extends CompanyBranchState {}

class CompanyBranchDeleteManySuccess extends CompanyBranchState {}

class CompanyBranchDeleteManyFailed extends CompanyBranchState {}

class CompanyBranchGetInProgress extends CompanyBranchState {}

class CompanyBranchGetSuccess extends CompanyBranchState {
  final CompanyBranchModel companyBranch;

  const CompanyBranchGetSuccess({required this.companyBranch});

  CompanyBranchGetSuccess copyWith({
    CompanyBranchModel? companyBranch,
  }) =>
      CompanyBranchGetSuccess(
          companyBranch: companyBranch ?? this.companyBranch);

  @override
  List<Object> get props => [companyBranch];
}

class CompanyBranchGetFailed extends CompanyBranchState {
  final String message;

  const CompanyBranchGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CompanyBranchUpdateInitial extends CompanyBranchState {}

class CompanyBranchUpdateInProgress extends CompanyBranchState {}

class CompanyBranchUpdateSuccess extends CompanyBranchState {}

class CompanyBranchUpdateFailed extends CompanyBranchState {
  final String message;

  const CompanyBranchUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CompanyBranchGetBycodeInProgress extends CompanyBranchState {}

class CompanyBranchGetBycodeSuccess extends CompanyBranchState {
  final CompanyBranchModel companyBranch;

  const CompanyBranchGetBycodeSuccess({required this.companyBranch});

  CompanyBranchGetBycodeSuccess copyWith({
    CompanyBranchModel? companyBranch,
  }) =>
      CompanyBranchGetBycodeSuccess(
          companyBranch: companyBranch ?? this.companyBranch);

  @override
  List<Object> get props => [companyBranch];
}

class CompanyBranchGetBycodeFailed extends CompanyBranchState {
  final String message;

  const CompanyBranchGetBycodeFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CompanyBranchByBusinessTypeInProgress extends CompanyBranchState {}

class CompanyBranchByBusinessTypeLoadSuccess extends CompanyBranchState {
  final List<CompanyBranchModel> companyBranch;

  const CompanyBranchByBusinessTypeLoadSuccess({required this.companyBranch});

  CompanyBranchByBusinessTypeLoadSuccess copyWith({
    List<CompanyBranchModel>? companyBranch,
  }) =>
      CompanyBranchByBusinessTypeLoadSuccess(
          companyBranch: companyBranch ?? this.companyBranch);

  @override
  List<Object> get props => [companyBranch];
}

class CompanyBranchByBusinessTypeLoadFailed extends CompanyBranchState {
  final String message;

  const CompanyBranchByBusinessTypeLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
