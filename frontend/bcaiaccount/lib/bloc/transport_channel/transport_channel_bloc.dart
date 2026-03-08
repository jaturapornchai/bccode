import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:smlaicloud/model/transport_channel_model.dart';
import 'package:smlaicloud/repositories/transport_channel_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/model/master_model.dart';
import 'package:smlaicloud/repositories/client.dart';

part 'transport_channel_event.dart';
part 'transport_channel_state.dart';

class TransportChannelBloc
    extends Bloc<TransportChannelEvent, TransportChannelState> {
  final TransportChannelRepository _transportchannelRepository;

  TransportChannelBloc(
      {required TransportChannelRepository transportChannelRepository})
      : _transportchannelRepository = transportChannelRepository,
        super(TransportChannelInitial()) {
    on<TransportChannelLoadList>(onTransportChannelLoad);
    on<TransportChannelSave>(onTransportChannelSave);
    on<TransportChannelUpdate>(onTransportChannelUpdate);
    on<TransportChannelDelete>(onTransportChannelDelete);
    on<TransportChannelDeleteMany>(onTransportChannelDeleteMany);
    on<TransportChannelGet>(onTransportChannelGet);
    on<TransportChannelWithImageSave>(onTransportChannelWithImageSave);
    on<TransportChannelWithImageUpdate>(onTransportChannelWithImageUpdate);
    on<TransportChannelSaveBulk>(onTransportChannelSaveBulk);
  }

  void onTransportChannelLoad(TransportChannelLoadList event,
      Emitter<TransportChannelState> emit) async {
    emit(TransportChannelInProgress());
    try {
      final results = await _transportchannelRepository.getTransportChannelList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<TransportChannelModel> transportchannel = (results.data as List)
            .map((transportchannel) =>
                TransportChannelModel.fromJson(transportchannel))
            .toList();
        // // print(transportchannel.length);
        emit(TransportChannelLoadSuccess(transportchannel: transportchannel));
      } else {
        emit(const TransportChannelLoadFailed(
            message: 'TransportChannel Not Found'));
      }
    } catch (e) {
      emit(TransportChannelLoadFailed(message: e.toString()));
    }
  }

  void onTransportChannelDelete(
      TransportChannelDelete event, Emitter<TransportChannelState> emit) async {
    emit(TransportChannelDeleteInProgress());
    try {
      await _transportchannelRepository.deleteTransportChannel(event.guid);

      emit(TransportChannelDeleteSuccess());
    } catch (e) {
      // emit(TransportChannelDeleteFailure(message: e.toString()));
    }
  }

  void onTransportChannelDeleteMany(TransportChannelDeleteMany event,
      Emitter<TransportChannelState> emit) async {
    emit(TransportChannelDeleteManyInProgress());
    try {
      await _transportchannelRepository.deleteTransportChannelMany(event.guid);

      emit(TransportChannelDeleteManySuccess());
    } catch (e) {
      // emit(TransportChannelDeleteFailure(message: e.toString()));
    }
  }

  void onTransportChannelSave(
      TransportChannelSave event, Emitter<TransportChannelState> emit) async {
    emit(TransportChannelSaveInProgress());
    try {
      // // print(event.transportchannelmodel);
      await _transportchannelRepository
          .saveTransportChannel(event.transportchannelmodel);
      emit(TransportChannelSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(TransportChannelSaveFailed(message: error['message']));
    }
  }

  void onTransportChannelSaveBulk(TransportChannelSaveBulk event,
      Emitter<TransportChannelState> emit) async {
    emit(TransportChannelSaveInProgress());
    try {
      await _transportchannelRepository
          .saveTransportChannelBulk(event.transportchannels);
      emit(TransportChannelSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(TransportChannelSaveFailed(message: error['message']));
    }
  }

  void onTransportChannelUpdate(
      TransportChannelUpdate event, Emitter<TransportChannelState> emit) async {
    emit(TransportChannelUpdateInProgress());
    try {
      await _transportchannelRepository.updateTransportChannel(
          event.guid, event.transportchannelmodel);
      emit(TransportChannelUpdateSuccess());
    } catch (e) {
      emit(TransportChannelUpdateFailed(message: e.toString()));
    }
  }

  void onTransportChannelGet(
      TransportChannelGet event, Emitter<TransportChannelState> emit) async {
    emit(TransportChannelGetInProgress());
    try {
      final result =
          await _transportchannelRepository.getTransportChannel(event.guid);
      if (result.success) {
        TransportChannelModel transportchannel =
            TransportChannelModel.fromJson(result.data);
        emit(TransportChannelGetSuccess(transportchannel: transportchannel));
      } else {
        emit(const TransportChannelGetFailed(
            message: 'TransportChannel Not Found'));
      }
    } catch (e) {
      // emit(TransportChannelDeleteFailure(message: e.toString()));
    }
  }

  void onTransportChannelWithImageSave(TransportChannelWithImageSave event,
      Emitter<TransportChannelState> emit) async {
    emit(TransportChannelSaveInProgress());
    try {
      ApiResponse result = await _transportchannelRepository.uploadImage(
          event.imageFile, event.imageWeb!);
      if (result.success) {
        UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
        TransportChannelModel transportChannelModel = event.transportchannel;
        transportChannelModel.imageuri = uploadImage.uri;
        await _transportchannelRepository
            .saveTransportChannel(transportChannelModel);
        emit(TransportChannelSaveSuccess());
      } else {
        emit(TransportChannelSaveFailed(message: result.message));
      }
    } catch (e) {
      emit(TransportChannelSaveFailed(message: e.toString()));
    }
  }

  void onTransportChannelWithImageUpdate(TransportChannelWithImageUpdate event,
      Emitter<TransportChannelState> emit) async {
    emit(TransportChannelUpdateInProgress());
    try {
      ApiResponse result = await _transportchannelRepository.uploadImage(
          event.imageFile, event.imageWeb);
      if (result.success) {
        UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
        TransportChannelModel transportChannelModel = event.transportchannel;
        transportChannelModel.imageuri = uploadImage.uri;
        await _transportchannelRepository.updateTransportChannel(
            event.guid, transportChannelModel);
        emit(TransportChannelUpdateSuccess());
      } else {
        emit(TransportChannelUpdateFailed(message: result.message));
      }
    } catch (e) {
      emit(TransportChannelUpdateFailed(message: e.toString()));
    }
  }
}
