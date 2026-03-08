part of 'import_product_bloc.dart';

abstract class ImportProductEvent extends Equatable {
  const ImportProductEvent();

  @override
  List<Object> get props => [];
}

class UploadFileExcel extends ImportProductEvent {
  final String filename;
  final Uint8List file;

  const UploadFileExcel({required this.file, required this.filename});

  @override
  List<Object> get props => [file, filename];
}

class LoadImportProductByTaskid extends ImportProductEvent {
  final String taskid;
  final int limit;
  final int page;
  final String q;

  const LoadImportProductByTaskid({
    required this.taskid,
    required this.limit,
    required this.page,
    required this.q,
  });

  @override
  List<Object> get props => [taskid];
}

class UpdateDetail extends ImportProductEvent {
  final String guid;
  final ImportProductModel importProductModel;

  const UpdateDetail({required this.guid, required this.importProductModel});

  @override
  List<Object> get props => [guid, importProductModel];
}

class AddDetail extends ImportProductEvent {
  final ImportProductModel importProductModel;

  const AddDetail({required this.importProductModel});

  @override
  List<Object> get props => [importProductModel];
}

class DeleteDetailByGuid extends ImportProductEvent {
  final String guid;

  const DeleteDetailByGuid({required this.guid});

  @override
  List<Object> get props => [guid];
}

class SaveTaskid extends ImportProductEvent {
  final String taskid;
  final String languangecode;

  const SaveTaskid({required this.taskid, required this.languangecode});

  @override
  List<Object> get props => [taskid, languangecode];
}

class VerifyTaskid extends ImportProductEvent {
  final String taskid;

  const VerifyTaskid({required this.taskid});

  @override
  List<Object> get props => [taskid];
}

class GetStatusTaskid extends ImportProductEvent {
  final String taskid;

  const GetStatusTaskid({required this.taskid});

  @override
  List<Object> get props => [taskid];
}

class CompareDetail extends ImportProductEvent {
  final String taskid;
  final int page;
  final int limit;
  final bool fast;

  const CompareDetail({
    required this.taskid,
    required this.page,
    required this.limit,
    required this.fast,
  });

  @override
  List<Object> get props => [taskid, page, limit, fast];
}

class ApplyImport extends ImportProductEvent {
  final String taskid;
  final bool async;
  final int batchSize;
  final bool forceUpdate;
  final String importMode;

  const ApplyImport({
    required this.taskid,
    required this.async,
    required this.batchSize,
    required this.forceUpdate,
    required this.importMode,
  });

  @override
  List<Object> get props => [taskid, async, batchSize, forceUpdate, importMode];
}
