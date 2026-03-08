part of 'doc_format_bloc.dart';

abstract class DocFormatEvent extends Equatable {
  const DocFormatEvent();

  @override
  List<Object> get props => [];
}

class DocFormatLoadDefault extends DocFormatEvent {
  const DocFormatLoadDefault();

  @override
  List<Object> get props => [];
}

class DocFormatLoadList extends DocFormatEvent {
  final int limit;
  final int offset;
  final String search;

  const DocFormatLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class DocFormatSave extends DocFormatEvent {
  final DocFormatModel docFormatModel;

  const DocFormatSave({
    required this.docFormatModel,
  });

  @override
  List<Object> get props => [docFormatModel];
}

class DocFormatBulkSave extends DocFormatEvent {
  final List<DocFormatModel> docFormatModel;

  const DocFormatBulkSave({
    required this.docFormatModel,
  });

  @override
  List<Object> get props => [docFormatModel];
}

class DocFormatDelete extends DocFormatEvent {
  final String guid;

  const DocFormatDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class DocFormatDeleteMany extends DocFormatEvent {
  final List<String> guid;

  const DocFormatDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class DocFormatGet extends DocFormatEvent {
  final String guid;

  const DocFormatGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class DocFormatUpdate extends DocFormatEvent {
  final String guid;
  final DocFormatModel docFormatModel;

  const DocFormatUpdate({
    required this.guid,
    required this.docFormatModel,
  });

  @override
  List<Object> get props => [DocFormatModel];
}
