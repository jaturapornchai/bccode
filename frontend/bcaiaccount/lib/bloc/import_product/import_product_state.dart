part of 'import_product_bloc.dart';

abstract class ImportProductState extends Equatable {
  const ImportProductState();

  @override
  List<Object> get props => [];
}

class UploadFileExcelInProgress extends ImportProductState {}

class ImportProductInitial extends ImportProductState {}

class UploadFileExcelSuccess extends ImportProductState {
  final UploadSuccessModel response;

  const UploadFileExcelSuccess({required this.response});

  @override
  List<Object> get props => [response];
}

class UploadFileExcelFailed extends ImportProductState {
  final String message;

  const UploadFileExcelFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class LoadImportProductByTaskidInProgress extends ImportProductState {}

class LoadImportProductByTaskidSuccess extends ImportProductState {
  final List<ImportProductModel> data;
  final Pagination pagination;

  const LoadImportProductByTaskidSuccess({
    required this.data,
    required this.pagination,
  });

  @override
  List<Object> get props => [data, pagination];
}

class LoadImportProductByTaskidFailed extends ImportProductState {
  final String message;

  const LoadImportProductByTaskidFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class DeleteDetailByGuidInProgress extends ImportProductState {}

class DeleteDetailByGuidSuccess extends ImportProductState {}

class DeleteDetailByGuidFailed extends ImportProductState {
  final String message;

  const DeleteDetailByGuidFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class UpdateDetailInProgress extends ImportProductState {}

class UpdateDetailSuccess extends ImportProductState {}

class UpdateDetailFailed extends ImportProductState {
  final String message;

  const UpdateDetailFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class AddDetailInProgress extends ImportProductState {}

class AddDetailSuccess extends ImportProductState {}

class AddDetailFailed extends ImportProductState {
  final String message;

  const AddDetailFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class SaveTaskidInProgress extends ImportProductState {}

class SaveTaskidSuccess extends ImportProductState {}

class SaveTaskidFailed extends ImportProductState {
  final String message;

  const SaveTaskidFailed({required this.message});

  @override
  List<Object> get props => [message];
}

/// verify taskid
class VerifyTaskidInProgress extends ImportProductState {}

class VerifyTaskidSuccess extends ImportProductState {}

class VerifyTaskidFailed extends ImportProductState {
  final String message;

  const VerifyTaskidFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class GetStatusTaskidInProgress extends ImportProductState {}

class GetStatusTaskidSuccess extends ImportProductState {
  final ImportProductStatusModel data;

  const GetStatusTaskidSuccess({required this.data});

  @override
  List<Object> get props => [data];
}

class GetStatusTaskidFailed extends ImportProductState {
  final String message;

  const GetStatusTaskidFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class CompareDetailInProgress extends ImportProductState {}

class CompareDetailSuccess extends ImportProductState {
  final CompareDetailModel data;
  final Pagination pagination;

  const CompareDetailSuccess({required this.data, required this.pagination});

  @override
  List<Object> get props => [data, pagination];
}

class CompareDetailFailed extends ImportProductState {
  final String message;

  const CompareDetailFailed({required this.message});

  @override
  List<Object> get props => [message];
}

class ApplyImportInProgress extends ImportProductState {}

class ApplyImportSuccess extends ImportProductState {
  final ApplyImportResponseModel data;

  const ApplyImportSuccess({required this.data});

  @override
  List<Object> get props => [data];
}

class ApplyImportFailed extends ImportProductState {
  final String message;

  const ApplyImportFailed({required this.message});

  @override
  List<Object> get props => [message];
}
