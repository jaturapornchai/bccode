part of 'company_branch_bloc.dart';

abstract class CompanyBranchEvent extends Equatable {
  const CompanyBranchEvent();

  @override
  List<Object> get props => [];
}

class CompanyBranchGet extends CompanyBranchEvent {
  final String guid;

  const CompanyBranchGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class CompanyBranchGetBycode extends CompanyBranchEvent {
  final String code;

  const CompanyBranchGetBycode({required this.code});

  @override
  List<Object> get props => [code];
}

class CompanyBranchLoadList extends CompanyBranchEvent {
  final int limit;
  final int offset;
  final String search;

  const CompanyBranchLoadList({
    required this.offset,
    required this.limit,
    required this.search,
  }); // Default to empty string if null

  @override
  List<Object> get props => [offset, limit, search];
}

class CompanyBranchByBusinessTypeLoadList extends CompanyBranchEvent {
  final int limit;
  final int offset;
  final String search;
  final String businesstypecode; // Removed the nullable indicator

  const CompanyBranchByBusinessTypeLoadList({
    required this.offset,
    required this.limit,
    required this.search,
    String? businesstypecode,
  }) : businesstypecode =
            businesstypecode ?? ""; // Default to empty string if null

  @override
  List<Object> get props => [offset, limit, search, businesstypecode];
}

class CompanyBranchDelete extends CompanyBranchEvent {
  final String guid;

  const CompanyBranchDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class CompanyBranchDeleteMany extends CompanyBranchEvent {
  final List<String> guid;

  const CompanyBranchDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class CompanyBranchSave extends CompanyBranchEvent {
  final CompanyBranchModel companyBranch;

  const CompanyBranchSave({
    required this.companyBranch,
  });

  @override
  List<Object> get props => [companyBranch];
}

class CompanyBranchUpdate extends CompanyBranchEvent {
  final String guid;
  final CompanyBranchModel companyBranch;

  const CompanyBranchUpdate({
    required this.guid,
    required this.companyBranch,
  });

  @override
  List<Object> get props => [companyBranch];
}
