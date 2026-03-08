import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:smlaicloud/model/sale_channel_model.dart';
import 'package:smlaicloud/repositories/sale_channel_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/model/master_model.dart';
import 'package:smlaicloud/repositories/client.dart';

part 'sale_channel_event.dart';
part 'sale_channel_state.dart';

class SaleChannelBloc extends Bloc<SaleChannelEvent, SaleChannelState> {
  final SaleChannelRepository _salechannelRepository;

  SaleChannelBloc({required SaleChannelRepository saleChannelRepository})
      : _salechannelRepository = saleChannelRepository,
        super(SaleChannelInitial()) {
    on<SaleChannelLoadList>(onSaleChannelLoad);
    on<SaleChannelSave>(onSaleChannelSave);
    on<SaleChannelUpdate>(onSaleChannelUpdate);
    on<SaleChannelDelete>(onSaleChannelDelete);
    on<SaleChannelDeleteMany>(onSaleChannelDeleteMany);
    on<SaleChannelGet>(onSaleChannelGet);
    on<SaleChannelWithImageSave>(onSaleChannelWithImageSave);
    on<SaleChannelWithImageUpdate>(onSaleChannelWithImageUpdate);
    on<SaleChannelSaveBulk>(onSaleChannelSaveBulk);
  }

  void onSaleChannelLoad(
      SaleChannelLoadList event, Emitter<SaleChannelState> emit) async {
    emit(SaleChannelInProgress());
    try {
      final results = await _salechannelRepository.getSaleChannelList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<SaleChannelModel> salechannel = (results.data as List)
            .map((salechannel) => SaleChannelModel.fromJson(salechannel))
            .toList();
        emit(SaleChannelLoadSuccess(salechannel: salechannel));
      } else {
        emit(const SaleChannelLoadFailed(message: 'SaleChannel Not Found'));
      }
    } catch (e) {
      emit(SaleChannelLoadFailed(message: e.toString()));
    }
  }

  void onSaleChannelDelete(
      SaleChannelDelete event, Emitter<SaleChannelState> emit) async {
    emit(SaleChannelDeleteInProgress());
    try {
      await _salechannelRepository.deleteSaleChannel(event.guid);

      emit(SaleChannelDeleteSuccess());
    } catch (e) {
      // emit(SaleChannelDeleteFailure(message: e.toString()));
    }
  }

  void onSaleChannelDeleteMany(
      SaleChannelDeleteMany event, Emitter<SaleChannelState> emit) async {
    emit(SaleChannelDeleteManyInProgress());
    try {
      await _salechannelRepository.deleteSaleChannelMany(event.guid);

      emit(SaleChannelDeleteManySuccess());
    } catch (e) {
      // emit(SaleChannelDeleteFailure(message: e.toString()));
    }
  }

  void onSaleChannelSave(
      SaleChannelSave event, Emitter<SaleChannelState> emit) async {
    emit(SaleChannelSaveInProgress());
    try {
      await _salechannelRepository.saveSaleChannel(event.salechannelmodel);
      emit(SaleChannelSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(SaleChannelSaveFailed(message: error['message']));
    }
  }

  void onSaleChannelUpdate(
      SaleChannelUpdate event, Emitter<SaleChannelState> emit) async {
    emit(SaleChannelUpdateInProgress());
    try {
      await _salechannelRepository.updateSaleChannel(
          event.guid, event.salechannelmodel);
      emit(SaleChannelUpdateSuccess());
    } catch (e) {
      emit(SaleChannelUpdateFailed(message: e.toString()));
    }
  }

  void onSaleChannelGet(
      SaleChannelGet event, Emitter<SaleChannelState> emit) async {
    emit(SaleChannelGetInProgress());
    try {
      final result = await _salechannelRepository.getSaleChannel(event.guid);
      if (result.success) {
        SaleChannelModel salechannel = SaleChannelModel.fromJson(result.data);
        emit(SaleChannelGetSuccess(salechannel: salechannel));
      } else {
        emit(const SaleChannelGetFailed(message: 'SaleChannel Not Found'));
      }
    } catch (e) {
      // emit(SaleChannelDeleteFailure(message: e.toString()));
    }
  }

  void onSaleChannelWithImageSave(
      SaleChannelWithImageSave event, Emitter<SaleChannelState> emit) async {
    emit(SaleChannelSaveInProgress());
    try {
      ApiResponse result = await _salechannelRepository.uploadImage(
          event.imageFile, event.imageWeb!);
      if (result.success) {
        UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
        SaleChannelModel saleChannelModel = event.salechannel;
        saleChannelModel.imageuri = uploadImage.uri;
        await _salechannelRepository.saveSaleChannel(saleChannelModel);
        emit(SaleChannelSaveSuccess());
      } else {
        emit(SaleChannelSaveFailed(message: result.message));
      }
    } catch (e) {
      emit(SaleChannelSaveFailed(message: e.toString()));
    }
  }

  void onSaleChannelWithImageUpdate(
      SaleChannelWithImageUpdate event, Emitter<SaleChannelState> emit) async {
    emit(SaleChannelUpdateInProgress());
    try {
      ApiResponse result = await _salechannelRepository.uploadImage(
          event.imageFile, event.imageWeb);
      if (result.success) {
        UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
        SaleChannelModel saleChannelModel = event.salechannel;
        saleChannelModel.imageuri = uploadImage.uri;
        await _salechannelRepository.updateSaleChannel(
            event.guid, saleChannelModel);
        emit(SaleChannelUpdateSuccess());
      } else {
        emit(SaleChannelUpdateFailed(message: result.message));
      }
    } catch (e) {
      emit(SaleChannelUpdateFailed(message: e.toString()));
    }
  }

  void onSaleChannelSaveBulk(
      SaleChannelSaveBulk event, Emitter<SaleChannelState> emit) async {
    emit(SaleChannelSaveInProgress());
    try {
      await _salechannelRepository.saveSaleChannelBulk(event.salechannels);
      emit(SaleChannelSaveSuccess());
    } catch (e) {
      emit(SaleChannelSaveFailed(message: e.toString()));
    }
  }
}
