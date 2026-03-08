import 'dart:convert';
import 'package:smlaicloud/model/master_design_model.dart';
import 'package:smlaicloud/repositories/master_design_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'master_design_event.dart';
part 'master_design_state.dart';

class MasterDesignBloc extends Bloc<MasterDesignEvent, MasterDesignState> {
  final MasterDesignRepository _masterDesignRepository;

  MasterDesignBloc({required MasterDesignRepository masterDesignRepository})
      : _masterDesignRepository = masterDesignRepository,
        super(MasterDesignInitial()) {
    on<MasterDesignLoadList>(onMasterDesignLoad);
    on<MasterDesignSave>(onMasterDesignSave);
    on<MasterDesignUpdate>(onMasterDesignUpdate);
    on<MasterDesignDelete>(onMasterDesignDelete);
    on<MasterDesignDeleteMany>(onMasterDesignDeleteMany);
    on<MasterDesignGet>(onMasterDesignGet);
    on<MasterDesignGetByCode>(onMasterDesignGetByCode);
  }

  void onMasterDesignLoad(
      MasterDesignLoadList event, Emitter<MasterDesignState> emit) async {
    emit(MasterDesignInProgress());
    try {
      final results = await _masterDesignRepository.getDesignList(
          offset: event.offset, limit: event.limit, search: event.search);
      if (results.success) {
        List<MasterDesignModel> designs = (results.data as List)
            .map((design) => MasterDesignModel.fromJson(design))
            .toList();
        emit(MasterDesignLoadSuccess(designs: designs));
      } else {
        emit(const MasterDesignLoadFailed(message: 'Design Not Found'));
      }
    } catch (e) {
      emit(MasterDesignLoadFailed(message: e.toString()));
    }
  }

  void onMasterDesignDelete(
      MasterDesignDelete event, Emitter<MasterDesignState> emit) async {
    emit(MasterDesignDeleteInProgress());
    try {
      await _masterDesignRepository.deleteDesign(event.guid);
      emit(MasterDesignDeleteSuccess());
    } catch (e) {
      emit(MasterDesignDeleteFailed());
    }
  }

  void onMasterDesignDeleteMany(
      MasterDesignDeleteMany event, Emitter<MasterDesignState> emit) async {
    emit(MasterDesignDeleteManyInProgress());
    try {
      await _masterDesignRepository.deleteDesignMany(event.guid);
      emit(MasterDesignDeleteManySuccess());
    } catch (e) {
      emit(MasterDesignDeleteManyFailed());
    }
  }

  void onMasterDesignSave(
      MasterDesignSave event, Emitter<MasterDesignState> emit) async {
    emit(MasterDesignSaveInProgress());
    try {
      await _masterDesignRepository.saveDesign(event.designModel);
      emit(MasterDesignSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(MasterDesignSaveFailed(message: error['message']));
    }
  }

  void onMasterDesignUpdate(
      MasterDesignUpdate event, Emitter<MasterDesignState> emit) async {
    emit(MasterDesignUpdateInProgress());
    try {
      await _masterDesignRepository.updateDesign(event.guid, event.designModel);
      emit(MasterDesignUpdateSuccess());
    } catch (e) {
      emit(MasterDesignUpdateFailed(message: e.toString()));
    }
  }

  void onMasterDesignGet(
      MasterDesignGet event, Emitter<MasterDesignState> emit) async {
    emit(MasterDesignGetInProgress());
    try {
      final result = await _masterDesignRepository.getDesign(event.guid);
      if (result.success) {
        MasterDesignModel design = MasterDesignModel.fromJson(result.data);
        emit(MasterDesignGetSuccess(design: design));
      } else {
        emit(const MasterDesignGetFailed(message: 'Design Not Found'));
      }
    } catch (e) {
      emit(MasterDesignGetFailed(message: e.toString()));
    }
  }

  void onMasterDesignGetByCode(
      MasterDesignGetByCode event, Emitter<MasterDesignState> emit) async {
    emit(MasterDesignGetInProgress());
    try {
      final result = await _masterDesignRepository.getDesignByCode(event.code);
      if (result.success) {
        MasterDesignModel design = MasterDesignModel.fromJson(result.data);
        emit(MasterDesignGetSuccess(design: design));
      } else {
        emit(const MasterDesignGetFailed(message: 'Design Not Found'));
      }
    } catch (e) {
      emit(MasterDesignGetFailed(message: e.toString()));
    }
  }
}
