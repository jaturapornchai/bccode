import 'dart:convert';

import 'package:smlaicloud/model/creditor_model.dart';
import 'package:smlaicloud/repositories/creditor_repository.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/master_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'dart:io';

part 'creditor_event.dart';
part 'creditor_state.dart';

class CreditorBloc extends Bloc<CreditorEvent, CreditorState> {
  final CreditorRepository _creditorRepository;

  CreditorBloc({required CreditorRepository creditorRepository})
      : _creditorRepository = creditorRepository,
        super(CreditorInitial()) {
    on<CreditorLoadList>(onCreditorLoad);
    on<CreditorSave>(onCreditorSave);
    on<CreditorBulkSave>(onCreditorBulkSave);
    on<CreditorUpdate>(onCreditorUpdate);
    on<CreditorWithImageSave>(onCreditorWithImageSave);
    on<CreditorWithImageUpdate>(onCreditorWithImageUpdate);
    on<CreditorDelete>(onCreditorDelete);
    on<CreditorDeleteMany>(onCreditorDeleteMany);
    on<CreditorGet>(onCreditorGet);
    on<CreditorGetBycode>(onCreditorGetBycode);
  }

  void onCreditorLoad(
      CreditorLoadList event, Emitter<CreditorState> emit) async {
    emit(CreditorInProgress());

    try {
      final results = await _creditorRepository.getCreditorList(
          offset: event.offset,
          limit: event.limit,
          search: event.search,
          groups: event.groups);

      if (results.success) {
        List<CreditorModel> creditors = (results.data as List)
            .map((creditor) => CreditorModel.fromJson(creditor))
            .toList();
        emit(CreditorLoadSuccess(creditors: creditors));
      } else {
        emit(const CreditorLoadFailed(message: 'Creditor Not Found'));
      }
    } catch (e) {
      emit(CreditorLoadFailed(message: e.toString()));
    }
  }

  void onCreditorDelete(
      CreditorDelete event, Emitter<CreditorState> emit) async {
    emit(CreditorDeleteInProgress());
    try {
      await _creditorRepository.deleteCreditor(event.guid);

      emit(CreditorDeleteSuccess());
    } catch (e) {
      // emit(CreditorDeleteFailure(message: e.toString()));
    }
  }

  void onCreditorDeleteMany(
      CreditorDeleteMany event, Emitter<CreditorState> emit) async {
    emit(CreditorDeleteManyInProgress());
    try {
      await _creditorRepository.deleteCreditorMany(event.guid);

      emit(CreditorDeleteManySuccess());
    } catch (e) {
      // emit(CreditorDeleteFailure(message: e.toString()));
    }
  }

  void onCreditorSave(CreditorSave event, Emitter<CreditorState> emit) async {
    emit(CreditorSaveInProgress());
    try {
      CreditorModel creditor = event.creditor;

      CreditorRequestModel creditorRequestModel = CreditorRequestModel(
        guidfixed: creditor.guidfixed,
        code: creditor.code,
        names: creditor.names,
        customertype: creditor.customertype,
        branchnumber: creditor.branchnumber,
        personaltype: creditor.personaltype,
        addressforbilling: creditor.addressforbilling,
        addressforshipping: creditor.addressforshipping,
        images: creditor.images,
        taxid: creditor.taxid,
        email: creditor.email,
        creditday: creditor.creditday,
        fundcode: creditor.fundcode,
      );

      for (var element in creditor.groups) {
        creditorRequestModel.groups.add(element.guidfixed);
      }

      await _creditorRepository.saveCreditor(creditorRequestModel);
      emit(CreditorSaveSuccess());
    } catch (e) {
      emit(CreditorSaveFailed(message: e.toString()));
    }
  }

  void onCreditorBulkSave(
      CreditorBulkSave event, Emitter<CreditorState> emit) async {
    emit(CreditorBulkSaveInProgress());
    try {
      final result =
          await _creditorRepository.saveBulkCreditors(event.creditors);
      if (result.success) {
        emit(CreditorBulkSaveSuccess(savedCount: event.creditors.length));
      } else {
        emit(CreditorBulkSaveFailed(message: result.message));
      }
    } catch (e) {
      emit(CreditorBulkSaveFailed(message: e.toString()));
    }
  }

  void onCreditorWithImageSave(
      CreditorWithImageSave event, Emitter<CreditorState> emit) async {
    emit(CreditorSaveInProgress());
    try {
      List<ImagesModel> images = [];
      if (event.imageFile.isNotEmpty) {
        for (int i = 0; i < event.imageFile.length; i++) {
          if (event.imageFile[i].uri.toString() != '') {
            ApiResponse result = await _creditorRepository.uploadImage(
                event.imageFile[i], event.imageWeb[i]);
            if (result.success) {
              UploadImageModel uploadImage =
                  UploadImageModel.fromJson(result.data);
              images.add(ImagesModel(uri: uploadImage.uri, xorder: i));
            } else {
              emit(CreditorSaveFailed(message: result.message));
            }
          }
        }

        if (images.isNotEmpty) {
          CreditorModel creditor = event.creditor;

          CreditorRequestModel creditorRequestModel = CreditorRequestModel(
            guidfixed: creditor.guidfixed,
            code: creditor.code,
            names: creditor.names,
            customertype: creditor.customertype,
            branchnumber: creditor.branchnumber,
            personaltype: creditor.personaltype,
            addressforbilling: creditor.addressforbilling,
            addressforshipping: creditor.addressforshipping,
            images: images,
            taxid: creditor.taxid,
            email: creditor.email,
            creditday: creditor.creditday,
            fundcode: creditor.fundcode,
          );

          for (var element in creditor.groups) {
            creditorRequestModel.groups.add(element.guidfixed);
          }

          await _creditorRepository.saveCreditor(creditorRequestModel);
          emit(CreditorSaveSuccess());
        } else {
          emit(const CreditorSaveFailed(message: 'image upload failed'));
        }
      } else {
        emit(const CreditorSaveFailed(message: 'no image found'));
      }
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(CreditorSaveFailed(message: error['message']));
    }
  }

  void onCreditorUpdate(
      CreditorUpdate event, Emitter<CreditorState> emit) async {
    emit(CreditorUpdateInProgress());
    try {
      CreditorModel creditor = event.creditorModel;

      CreditorRequestModel creditorRequestModel = CreditorRequestModel(
        guidfixed: creditor.guidfixed,
        code: creditor.code,
        names: creditor.names,
        customertype: creditor.customertype,
        branchnumber: creditor.branchnumber,
        personaltype: creditor.personaltype,
        addressforbilling: creditor.addressforbilling,
        addressforshipping: creditor.addressforshipping,
        images: creditor.images,
        taxid: creditor.taxid,
        email: creditor.email,
        creditday: creditor.creditday,
        fundcode: creditor.fundcode,
      );

      for (var element in creditor.groups) {
        creditorRequestModel.groups.add(element.guidfixed);
      }

      await _creditorRepository.updateCreditor(
          event.guid, creditorRequestModel);
      emit(CreditorUpdateSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(CreditorUpdateFailed(message: error['message']));
    }
  }

  void onCreditorWithImageUpdate(
      CreditorWithImageUpdate event, Emitter<CreditorState> emit) async {
    emit(CreditorUpdateInProgress());
    try {
      List<ImagesModel> images = [];
      if (event.imagesUris.isNotEmpty) {
        for (int i = 0; i < event.imagesUris.length; i++) {
          if (event.imageWeb[i].isNotEmpty) {
            ApiResponse result = await _creditorRepository.uploadImage(
                event.imageFiles[i], event.imageWeb[i]);
            if (result.success) {
              UploadImageModel uploadImage =
                  UploadImageModel.fromJson(result.data);
              images.add(ImagesModel(uri: uploadImage.uri, xorder: i));
            } else {
              emit(CreditorUpdateFailed(message: result.message));
            }
          } else if (event.imagesUris[i].uri != '') {
            images.add(ImagesModel(uri: event.imagesUris[i].uri, xorder: i));
          }
        }

        if (images.isNotEmpty) {
          CreditorModel creditor = event.creditor;

          CreditorRequestModel creditorRequestModel = CreditorRequestModel(
            guidfixed: creditor.guidfixed,
            code: creditor.code,
            names: creditor.names,
            customertype: creditor.customertype,
            branchnumber: creditor.branchnumber,
            personaltype: creditor.personaltype,
            addressforbilling: creditor.addressforbilling,
            addressforshipping: creditor.addressforshipping,
            images: images,
            taxid: creditor.taxid,
            email: creditor.email,
            creditday: creditor.creditday,
            fundcode: creditor.fundcode,
          );

          for (var element in creditor.groups) {
            creditorRequestModel.groups.add(element.guidfixed);
          }

          await _creditorRepository.updateCreditor(
              event.guid, creditorRequestModel);
          emit(CreditorUpdateSuccess());
        } else {
          emit(const CreditorUpdateFailed(message: 'image upload failed'));
        }
      } else {
        emit(const CreditorUpdateFailed(message: 'no image found'));
      }
    } catch (e) {
      emit(CreditorUpdateFailed(message: e.toString()));
    }
  }

  void onCreditorGet(CreditorGet event, Emitter<CreditorState> emit) async {
    emit(CreditorGetInProgress());
    try {
      final result = await _creditorRepository.getCreditor(event.guid);
      if (result.success) {
        CreditorModel creditor = CreditorModel.fromJson(result.data);
        // print(creditor.toJson());
        emit(CreditorGetSuccess(creditors: creditor));
      } else {
        emit(const CreditorGetFailed(message: 'Creditor Not Found'));
      }
    } catch (e) {
      emit(CreditorGetFailed(message: e.toString()));
    }
  }

  void onCreditorGetBycode(
      CreditorGetBycode event, Emitter<CreditorState> emit) async {
    emit(CreditorGetBycodeInProgress());
    try {
      final result =
          await _creditorRepository.getCreditorBycode(event.custcode);
      if (result.success) {
        CreditorModel creditor = CreditorModel.fromJson(result.data);
        emit(CreditorGetBycodeSuccess(creditors: creditor));
      } else {
        emit(const CreditorGetBycodeFailed(message: 'Creditor Not Found'));
      }
    } catch (e) {
      emit(CreditorGetBycodeFailed(message: e.toString()));
    }
  }
}
