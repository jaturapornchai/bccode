import 'dart:convert';

import 'package:smlaicloud/model/doc_format_model.dart';
import 'package:smlaicloud/repositories/document_formate_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'doc_format_event.dart';
part 'doc_format_state.dart';

class DocFormatBloc extends Bloc<DocFormatEvent, DocFormatState> {
  final DocumentFormateRepository _docFormatRepository;
  DocFormatBloc({required DocumentFormateRepository documentFormateRepository})
      : _docFormatRepository = documentFormateRepository,
        super(DocFormatInitial()) {
    on<DocFormatLoadDefault>(onDocFormatLoad);
    on<DocFormatLoadList>(onDocFormatLoadList);
    on<DocFormatSave>(onDocFormatSave);
    on<DocFormatDelete>(onDocFormatDelete);
    on<DocFormatDeleteMany>(onDocFormatDeleteMany);
    on<DocFormatGet>(onDocFormatGet);
    on<DocFormatUpdate>(onDocFormatUpdate);
    on<DocFormatBulkSave>(onDocFormatBulk);
  }

  void onDocFormatLoad(
      DocFormatLoadDefault event, Emitter<DocFormatState> emit) async {
    emit(DocFormatLoadDefaultInProgress());

    try {
      final results = await _docFormatRepository.getDocFormatDefault();

      if (results.success) {
        List<DefaultDocFormatModel> docFormats = (results.data as List)
            .map((docFormats) => DefaultDocFormatModel.fromJson(docFormats))
            .toList();
        emit(DocFormatLoadDefaultSuccess(docFormats: docFormats));
      } else {
        emit(const DocFormatLoadDefaultFailed(
            message: 'Document Format Not Found'));
      }
    } catch (e) {
      emit(DocFormatLoadDefaultFailed(message: e.toString()));
    }
  }

  void onDocFormatLoadList(
      DocFormatLoadList event, Emitter<DocFormatState> emit) async {
    emit(DocFormatLoadListInProgres());
    try {
      final results = await _docFormatRepository.getDocFormatList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<DocFormatModel> docFormat = (results.data as List)
            .map((docFormat) => DocFormatModel.fromJson(docFormat))
            .toList();
        emit(DocFormatLoadListSuccess(docFormat: docFormat));
      } else {
        emit(const DocFormatLoadListFailed(message: 'Doc Format Not Found'));
      }
    } catch (e) {
      emit(DocFormatLoadListFailed(message: e.toString()));
    }
  }

  void onDocFormatSave(
      DocFormatSave event, Emitter<DocFormatState> emit) async {
    emit(DocFormatSaveInProgress());
    try {
      // print(event.docFormatModel.toJson());
      await _docFormatRepository.saveDocFormat(event.docFormatModel);
      emit(DocFormatSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(DocFormatSaveFailed(message: error['message']));
    }
  }

  void onDocFormatDelete(
      DocFormatDelete event, Emitter<DocFormatState> emit) async {
    emit(DocFormatDeleteInProgress());
    try {
      await _docFormatRepository.deleteDocFormat(event.guid);

      emit(DocFormatDeleteSuccess());
    } catch (e) {
      emit(DocFormatDeleteFailed(message: e.toString()));
    }
  }

  void onDocFormatDeleteMany(
      DocFormatDeleteMany event, Emitter<DocFormatState> emit) async {
    emit(DocFormatDeleteManyInProgress());
    try {
      await _docFormatRepository.deleteDocFormatMany(event.guid);

      emit(DocFormatDeleteManySuccess());
    } catch (e) {
      emit(DocFormatDeleteManyFailed(message: e.toString()));
    }
  }

  void onDocFormatGet(DocFormatGet event, Emitter<DocFormatState> emit) async {
    emit(DocFormatGetInProgress());
    try {
      final result = await _docFormatRepository.getDocFormat(event.guid);
      if (result.success) {
        DocFormatModel docFormat = DocFormatModel.fromJson(result.data);
        emit(DocFormatGetSuccess(docFormat: docFormat));
      } else {
        emit(const DocFormatGetFailed(message: 'DocFormat Not Found'));
      }
    } catch (e) {
      emit(DocFormatGetFailed(message: e.toString()));
    }
  }

  void onDocFormatUpdate(
      DocFormatUpdate event, Emitter<DocFormatState> emit) async {
    emit(DocFormatUpdateInProgress());
    try {
      await _docFormatRepository.updateDocFormat(
          event.guid, event.docFormatModel);
      emit(DocFormatUpdateSuccess());
    } catch (e) {
      emit(DocFormatUpdateFailed(message: e.toString()));
    }
  }

  void onDocFormatBulk(
      DocFormatBulkSave event, Emitter<DocFormatState> emit) async {
    emit(DocFormatSaveBulkInProgress());
    try {
      await _docFormatRepository.saveDocFormatBulk(event.docFormatModel);
      emit(DocFormatSaveBulkSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(DocFormatSaveBulkFailed(message: error['message']));
    }
  }
}
