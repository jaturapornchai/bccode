part of 'doc_format_bloc.dart';

abstract class DocFormatState extends Equatable {
  const DocFormatState();

  @override
  List<Object> get props => [];
}

class DocFormatInitial extends DocFormatState {}

class DocFormatLoadDefaultInProgress extends DocFormatState {}

class DocFormatLoadDefaultSuccess extends DocFormatState {
  final List<DefaultDocFormatModel> docFormats;

  const DocFormatLoadDefaultSuccess({required this.docFormats});

  DocFormatLoadDefaultSuccess copyWith({
    List<DefaultDocFormatModel>? docFormats,
  }) =>
      DocFormatLoadDefaultSuccess(docFormats: docFormats ?? this.docFormats);

  @override
  List<Object> get props => [docFormats];
}

class DocFormatLoadDefaultFailed extends DocFormatState {
  final String message;

  const DocFormatLoadDefaultFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DocFormatLoadListInitial extends DocFormatState {}

class DocFormatLoadListInProgres extends DocFormatState {}

class DocFormatLoadListSuccess extends DocFormatState {
  final List<DocFormatModel> docFormat;

  const DocFormatLoadListSuccess({required this.docFormat});

  DocFormatLoadListSuccess copyWith({
    List<DocFormatModel>? docFormat,
  }) =>
      DocFormatLoadListSuccess(docFormat: docFormat ?? this.docFormat);

  @override
  List<Object> get props => [docFormat];
}

class DocFormatLoadListFailed extends DocFormatState {
  final String message;

  const DocFormatLoadListFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DocFormatSaveInitial extends DocFormatState {}

class DocFormatSaveInProgress extends DocFormatState {}

class DocFormatSaveSuccess extends DocFormatState {}

class DocFormatSaveFailed extends DocFormatState {
  final String message;

  const DocFormatSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DocFormatSaveBulkInitial extends DocFormatState {}

class DocFormatSaveBulkInProgress extends DocFormatState {}

class DocFormatSaveBulkSuccess extends DocFormatState {}

class DocFormatSaveBulkFailed extends DocFormatState {
  final String message;

  const DocFormatSaveBulkFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DocFormatDeleteInProgress extends DocFormatState {}

class DocFormatDeleteSuccess extends DocFormatState {}

class DocFormatDeleteFailed extends DocFormatState {
  final String message;

  const DocFormatDeleteFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DocFormatDeleteManyInProgress extends DocFormatState {}

class DocFormatDeleteManySuccess extends DocFormatState {}

class DocFormatDeleteManyFailed extends DocFormatState {
  final String message;

  const DocFormatDeleteManyFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DocFormatGetInProgress extends DocFormatState {}

class DocFormatGetSuccess extends DocFormatState {
  final DocFormatModel docFormat;

  const DocFormatGetSuccess({required this.docFormat});

  DocFormatGetSuccess copyWith({
    DocFormatModel? docFormat,
  }) =>
      DocFormatGetSuccess(docFormat: docFormat ?? this.docFormat);

  @override
  List<Object> get props => [docFormat];
}

class DocFormatGetFailed extends DocFormatState {
  final String message;

  const DocFormatGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class DocFormatUpdateInitial extends DocFormatState {}

class DocFormatUpdateInProgress extends DocFormatState {}

class DocFormatUpdateSuccess extends DocFormatState {}

class DocFormatUpdateFailed extends DocFormatState {
  final String message;

  const DocFormatUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
