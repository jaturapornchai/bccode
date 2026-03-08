import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/master_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/json_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/repositories/bank_repository.dart';
import 'package:smlaicloud/model/bank_model.dart';

part 'bank_event.dart';
part 'bank_state.dart';

class BankBloc extends Bloc<BankEvent, BankState> {
  final BankRepository _bankRepository;
  final JsonRepository _jsonRepository;
  BankBloc(
      {required BankRepository bankRepository,
      required JsonRepository jsonRepository})
      : _bankRepository = bankRepository,
        _jsonRepository = jsonRepository,
        super(BankInitial()) {
    on<BankLoadList>(onBankLoad);
    on<BankSave>(onBankSave);
    on<BankBulkSave>(onBankBulkSave);
    on<BankSaveWithImage>(onBankSaveWithImage);
    on<BankUpdate>(onBankUpdate);
    on<BankWithImageUpdate>(onBankWithImageUpdate);
    on<BankDelete>(bankDelete);
    on<BankDeleteMany>(bankDeleteMany);
    on<BankGet>(onBankGet);
  }

  void onBankLoad(BankLoadList event, Emitter<BankState> emit) async {
    emit(BankInProgress());

    try {
      final results = await _bankRepository.getBankList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<BankModel> banks = (results.data as List)
            .map((bank) => BankModel.fromJson(bank))
            .toList();
        emit(BankLoadSuccess(banks: banks));
      } else {
        emit(const BankLoadFailed(message: 'Bank Not Found'));
      }
    } catch (e) {
      emit(BankLoadFailed(message: e.toString()));
    }
  }

  void bankDelete(BankDelete event, Emitter<BankState> emit) async {
    emit(BankDeleteInProgress());
    try {
      await _bankRepository.deleteBank(event.guid);

      emit(BankDeleteSuccess());
    } catch (e) {
      // emit(BankDeleteFailure(message: e.toString()));
    }
  }

  void bankDeleteMany(BankDeleteMany event, Emitter<BankState> emit) async {
    emit(BankDeleteManyInProgress());
    try {
      await _bankRepository.deleteBankMany(event.guid);

      emit(BankDeleteManySuccess());
    } catch (e) {
      // emit(BankDeleteFailure(message: e.toString()));
    }
  }

  void onBankSave(BankSave event, Emitter<BankState> emit) async {
    emit(BankSaveInProgress());
    try {
      await _bankRepository.saveBank(event.bank);
      emit(BankSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(BankSaveFailed(message: error['message']));
    }
  }

  void onBankSaveWithImage(
      BankSaveWithImage event, Emitter<BankState> emit) async {
    emit(BankSaveInProgress());
    try {
      ImagesModel images = ImagesModel(uri: "", xorder: 0);
      if (event.imageWeb.isNotEmpty) {
        ApiResponse result =
            await _jsonRepository.uploadImage(event.imageFile, event.imageWeb);
        if (result.success) {
          UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
          images = ImagesModel(uri: uploadImage.uri, xorder: 0);
        } else {
          emit(BankSaveFailed(message: result.message));
        }

        if (images.uri != "") {
          BankModel bankModel = event.bank;
          bankModel.logo = images.uri;
          // print(bankModel);
          await _bankRepository.saveBank(event.bank);
          emit(BankSaveSuccess());
        } else {
          emit(const BankSaveFailed(message: 'image upload failed'));
        }
      } else {
        emit(const BankSaveFailed(message: 'no image found'));
      }
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(BankSaveFailed(message: error['message']));
    }
  }

  void onBankWithImageUpdate(
      BankWithImageUpdate event, Emitter<BankState> emit) async {
    emit(BankUpdateInProgress());
    try {
      ImagesModel images = ImagesModel(uri: '', xorder: 0);

      if (event.imageWeb.isNotEmpty) {
        ApiResponse result =
            await _jsonRepository.uploadImage(event.imageFile, event.imageWeb);
        if (result.success) {
          UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
          images = ImagesModel(uri: uploadImage.uri, xorder: 0);
        } else {
          emit(BankUpdateFailed(message: result.message));
        }
      } else if (event.imagesUri.uri != '') {
        images = ImagesModel(uri: event.imagesUri.uri, xorder: 0);
      }

      if (images.uri != '') {
        BankModel bankModel = event.bank;
        bankModel.logo = images.uri;

        // print(bankModel);
        await _bankRepository.updateBank(event.guid, event.bank);
        emit(BankUpdateSuccess());
      } else {
        emit(const BankUpdateFailed(message: 'image upload failed'));
      }
    } catch (e) {
      emit(BankUpdateFailed(message: e.toString()));
    }
  }

  void onBankUpdate(BankUpdate event, Emitter<BankState> emit) async {
    emit(BankUpdateInProgress());
    try {
      await _bankRepository.updateBank(event.guid, event.bankModel);
      emit(BankUpdateSuccess());
    } catch (e) {
      emit(BankUpdateFailed(message: e.toString()));
    }
  }

  void onBankGet(BankGet event, Emitter<BankState> emit) async {
    emit(BankGetInProgress());
    try {
      final result = await _bankRepository.getBank(event.guid);
      if (result.success) {
        BankModel bank = BankModel.fromJson(result.data);
        emit(BankGetSuccess(bank: bank));
      } else {
        emit(const BankGetFailed(message: 'Bank Not Found'));
      }
    } catch (e) {
      // emit(BankDeleteFailure(message: e.toString()));
    }
  }

  void onBankBulkSave(BankBulkSave event, Emitter<BankState> emit) async {
    emit(BankSaveInProgress());
    try {
      await _bankRepository.saveBankBulk(event.banks);
      emit(BankSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(BankSaveFailed(message: error['message']));
    }
  }
}
