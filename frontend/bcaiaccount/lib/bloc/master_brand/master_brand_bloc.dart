import 'dart:convert';

import 'package:smlaicloud/model/master_brand_model.dart';
import 'package:smlaicloud/repositories/master_brand_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'master_brand_event.dart';
part 'master_brand_state.dart';

class MasterBrandBloc extends Bloc<MasterBrandEvent, MasterBrandState> {
  final MasterBrandRepository _masterBrandRepository;

  MasterBrandBloc({required MasterBrandRepository masterBrandRepository})
      : _masterBrandRepository = masterBrandRepository,
        super(MasterBrandInitial()) {
    on<MasterBrandLoadList>(onMasterBrandLoad);
    on<MasterBrandSave>(onMasterBrandSave);
    on<MasterBrandUpdate>(onMasterBrandUpdate);
    on<MasterBrandDelete>(onMasterBrandDelete);
    on<MasterBrandDeleteMany>(onMasterBrandDeleteMany);
    on<MasterBrandGet>(onMasterBrandGet);
    on<MasterBrandGetByCode>(onMasterBrandGetByCode);
  }

  void onMasterBrandLoad(
      MasterBrandLoadList event, Emitter<MasterBrandState> emit) async {
    emit(MasterBrandInProgress());

    try {
      final results = await _masterBrandRepository.getBrandList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<MasterBrandModel> brands = (results.data as List)
            .map((brand) => MasterBrandModel.fromJson(brand))
            .toList();
        emit(MasterBrandLoadSuccess(brands: brands));
      } else {
        emit(const MasterBrandLoadFailed(message: 'Brand Not Found'));
      }
    } catch (e) {
      emit(MasterBrandLoadFailed(message: e.toString()));
    }
  }

  void onMasterBrandDelete(
      MasterBrandDelete event, Emitter<MasterBrandState> emit) async {
    emit(MasterBrandDeleteInProgress());
    try {
      await _masterBrandRepository.deleteBrand(event.guid);

      emit(MasterBrandDeleteSuccess());
    } catch (e) {
      emit(MasterBrandDeleteFailed());
    }
  }

  void onMasterBrandDeleteMany(
      MasterBrandDeleteMany event, Emitter<MasterBrandState> emit) async {
    emit(MasterBrandDeleteManyInProgress());
    try {
      await _masterBrandRepository.deleteBrandMany(event.guid);

      emit(MasterBrandDeleteManySuccess());
    } catch (e) {
      emit(MasterBrandDeleteManyFailed());
    }
  }

  void onMasterBrandSave(
      MasterBrandSave event, Emitter<MasterBrandState> emit) async {
    emit(MasterBrandSaveInProgress());
    try {
      await _masterBrandRepository.saveBrand(event.brandModel);
      emit(MasterBrandSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(MasterBrandSaveFailed(message: error['message']));
    }
  }

  void onMasterBrandUpdate(
      MasterBrandUpdate event, Emitter<MasterBrandState> emit) async {
    emit(MasterBrandUpdateInProgress());
    try {
      await _masterBrandRepository.updateBrand(event.guid, event.brandModel);
      emit(MasterBrandUpdateSuccess());
    } catch (e) {
      emit(MasterBrandUpdateFailed(message: e.toString()));
    }
  }

  void onMasterBrandGet(
      MasterBrandGet event, Emitter<MasterBrandState> emit) async {
    emit(MasterBrandGetInProgress());
    try {
      final result = await _masterBrandRepository.getBrand(event.guid);
      if (result.success) {
        MasterBrandModel brand = MasterBrandModel.fromJson(result.data);
        emit(MasterBrandGetSuccess(brand: brand));
      } else {
        emit(const MasterBrandGetFailed(message: 'Brand Not Found'));
      }
    } catch (e) {
      emit(MasterBrandGetFailed(message: e.toString()));
    }
  }

  void onMasterBrandGetByCode(
      MasterBrandGetByCode event, Emitter<MasterBrandState> emit) async {
    emit(MasterBrandGetInProgress());
    try {
      final result = await _masterBrandRepository.getBrandByCode(event.code);
      if (result.success) {
        MasterBrandModel brand = MasterBrandModel.fromJson(result.data);
        emit(MasterBrandGetSuccess(brand: brand));
      } else {
        emit(const MasterBrandGetFailed(message: 'Brand Not Found'));
      }
    } catch (e) {
      emit(MasterBrandGetFailed(message: e.toString()));
    }
  }
}
