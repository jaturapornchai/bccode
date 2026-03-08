import 'dart:convert';

import 'package:smlaicloud/model/master_category_model.dart';
import 'package:smlaicloud/repositories/master_category_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'master_category_event.dart';
part 'master_category_state.dart';

class MasterCategoryBloc
    extends Bloc<MasterCategoryEvent, MasterCategoryState> {
  final MasterCategoryRepository _masterCategoryRepository;

  MasterCategoryBloc(
      {required MasterCategoryRepository masterCategoryRepository})
      : _masterCategoryRepository = masterCategoryRepository,
        super(MasterCategoryInitial()) {
    on<MasterCategoryLoadList>(onMasterCategoryLoad);
    on<MasterCategorySave>(onMasterCategorySave);
    on<MasterCategoryUpdate>(onMasterCategoryUpdate);
    on<MasterCategoryDelete>(onMasterCategoryDelete);
    on<MasterCategoryDeleteMany>(onMasterCategoryDeleteMany);
    on<MasterCategoryGet>(onMasterCategoryGet);
    on<MasterCategoryGetByCode>(onMasterCategoryGetByCode);
  }

  void onMasterCategoryLoad(
      MasterCategoryLoadList event, Emitter<MasterCategoryState> emit) async {
    emit(MasterCategoryInProgress());

    try {
      final results = await _masterCategoryRepository.getCategoryList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<MasterCategoryModel> categories = (results.data as List)
            .map((category) => MasterCategoryModel.fromJson(category))
            .toList();
        emit(MasterCategoryLoadSuccess(categories: categories));
      } else {
        emit(const MasterCategoryLoadFailed(message: 'Category Not Found'));
      }
    } catch (e) {
      emit(MasterCategoryLoadFailed(message: e.toString()));
    }
  }

  void onMasterCategoryDelete(
      MasterCategoryDelete event, Emitter<MasterCategoryState> emit) async {
    emit(MasterCategoryDeleteInProgress());
    try {
      await _masterCategoryRepository.deleteCategory(event.guid);

      emit(MasterCategoryDeleteSuccess());
    } catch (e) {
      emit(MasterCategoryDeleteFailed());
    }
  }

  void onMasterCategoryDeleteMany(
      MasterCategoryDeleteMany event, Emitter<MasterCategoryState> emit) async {
    emit(MasterCategoryDeleteManyInProgress());
    try {
      await _masterCategoryRepository.deleteCategoryMany(event.guid);

      emit(MasterCategoryDeleteManySuccess());
    } catch (e) {
      emit(MasterCategoryDeleteManyFailed());
    }
  }

  void onMasterCategorySave(
      MasterCategorySave event, Emitter<MasterCategoryState> emit) async {
    emit(MasterCategorySaveInProgress());
    try {
      await _masterCategoryRepository.saveCategory(event.categoryModel);
      emit(MasterCategorySaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(MasterCategorySaveFailed(message: error['message']));
    }
  }

  void onMasterCategoryUpdate(
      MasterCategoryUpdate event, Emitter<MasterCategoryState> emit) async {
    emit(MasterCategoryUpdateInProgress());
    try {
      await _masterCategoryRepository.updateCategory(
          event.guid, event.categoryModel);
      emit(MasterCategoryUpdateSuccess());
    } catch (e) {
      emit(MasterCategoryUpdateFailed(message: e.toString()));
    }
  }

  void onMasterCategoryGet(
      MasterCategoryGet event, Emitter<MasterCategoryState> emit) async {
    emit(MasterCategoryGetInProgress());
    try {
      final result = await _masterCategoryRepository.getCategory(event.guid);
      if (result.success) {
        MasterCategoryModel category =
            MasterCategoryModel.fromJson(result.data);
        emit(MasterCategoryGetSuccess(category: category));
      } else {
        emit(const MasterCategoryGetFailed(message: 'Category Not Found'));
      }
    } catch (e) {
      emit(MasterCategoryGetFailed(message: e.toString()));
    }
  }

  void onMasterCategoryGetByCode(
      MasterCategoryGetByCode event, Emitter<MasterCategoryState> emit) async {
    emit(MasterCategoryGetInProgress());
    try {
      final result =
          await _masterCategoryRepository.getCategoryByCode(event.code);
      if (result.success) {
        MasterCategoryModel category =
            MasterCategoryModel.fromJson(result.data);
        emit(MasterCategoryGetSuccess(category: category));
      } else {
        emit(const MasterCategoryGetFailed(message: 'Category Not Found'));
      }
    } catch (e) {
      emit(MasterCategoryGetFailed(message: e.toString()));
    }
  }
}
