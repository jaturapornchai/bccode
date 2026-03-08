import 'dart:convert';

import 'package:smlaicloud/global.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/master_model.dart';
import 'package:smlaicloud/repositories/json_repository.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/model/config_model.dart';

import 'dart:io';

part 'company_event.dart';
part 'company_state.dart';

class CompanyBloc extends Bloc<CompanyEvent, CompanyState> {
  final JsonRepository jsonRepo;

  CompanyBloc({required JsonRepository jsonRepository})
      : jsonRepo = jsonRepository,
        super(CompanyInitial()) {
    on<CompanyLoad>(onCompanyLoad);
    on<CompanySave>(onCompanySave);
    on<CompanyWithImageSave>(onCompanyWithImageSave);
    on<CompanyUpdate>(onCompanyUpdate);
    on<CompanyWithImageUpdate>(onCompanyWithImageUpdate);
  }

  void onCompanyLoad(CompanyLoad event, Emitter<CompanyState> emit) async {
    emit(CompanyInProgress());

    try {
      final results = await jsonRepo.getSetting("company", "");

      if (results.success) {
        if (results.data.length > 0) {
          CompanyModel company =
              CompanyModel.fromJson(json.decode(results.data[0]['body']));

          emit(CompanyLoadSuccess(
              guidFixed: results.data[0]['guidfixed'], company: company));
        } else {
          emit(CompanyLoadNotFound(
              message: 'Company Group Not Found',
              shopdefault: appConfig.getString("name") ?? ''));
        }
      } else {
        emit(const CompanyLoadFailed(message: 'Company Not Found'));
      }
    } catch (e) {
      emit(CompanyLoadFailed(message: e.toString()));
    }
  }

  void onCompanySave(CompanySave event, Emitter<CompanyState> emit) async {
    emit(CompanySaveInProgress());
    try {
      final data = event.company.toJson();

      final postData = {"code": 'company', "body": jsonEncode(data)};
      await jsonRepo.saveSetting(postData);
      emit(CompanySaveSuccess());
    } catch (e) {
      emit(CompanySaveFailed(message: e.toString()));
    }
  }

  void onCompanyWithImageSave(
      CompanyWithImageSave event, Emitter<CompanyState> emit) async {
    emit(CompanySaveInProgress());
    try {
      List<ImagesModel> images = [];
      if (event.imageFiles.isNotEmpty) {
        for (int i = 0; i < event.imageFiles.length; i++) {
          if (event.imageFiles[i].uri.toString() != '') {
            ApiResponse result = await jsonRepo.uploadImage(
                event.imageFiles[i], event.imageWeb[i]);
            if (result.success) {
              UploadImageModel uploadImage =
                  UploadImageModel.fromJson(result.data);
              images.add(ImagesModel(uri: uploadImage.uri, xorder: i));
            } else {
              emit(CompanySaveFailed(message: result.message));
            }
          }
        }

        if (images.length == event.imageFiles.length) {
          CompanyModel company = event.company;
          company.images = images;
          final data = company.toJson();
          final postData = {"code": 'company', "body": jsonEncode(data)};
          await jsonRepo.saveSetting(postData);
          emit(CompanySaveSuccess());
        } else {
          emit(const CompanySaveFailed(message: 'image upload failed'));
        }
      } else {
        emit(const CompanySaveFailed(message: 'no image found'));
      }
    } catch (e) {
      emit(CompanySaveFailed(message: e.toString()));
    }
  }

  void onCompanyUpdate(CompanyUpdate event, Emitter<CompanyState> emit) async {
    emit(CompanyUpdateInProgress());
    try {
      final data = event.company.toJson();

      final postData = {"code": 'company', "body": jsonEncode(data)};

      await jsonRepo.updateSetting(event.guid, postData);
      emit(CompanyUpdateSuccess());
    } catch (e) {
      emit(CompanyUpdateFailed(message: e.toString()));
    }
  }

  void onCompanyWithImageUpdate(
      CompanyWithImageUpdate event, Emitter<CompanyState> emit) async {
    emit(CompanyUpdateInProgress());
    try {
      List<ImagesModel> images = [];
      if (event.imagesUris.isNotEmpty) {
        for (int i = 0; i < event.imagesUris.length; i++) {
          if (event.imageWeb[i].isNotEmpty) {
            ApiResponse result = await jsonRepo.uploadImage(
                event.imageFiles[i], event.imageWeb[i]);
            if (result.success) {
              UploadImageModel uploadImage =
                  UploadImageModel.fromJson(result.data);
              images.add(ImagesModel(uri: uploadImage.uri, xorder: i));
            } else {
              emit(CompanyUpdateFailed(message: result.message));
            }
          } else if (event.imagesUris[i].uri != '') {
            images.add(ImagesModel(uri: event.imagesUris[i].uri, xorder: i));
          }
        }

        if (images.isNotEmpty) {
          CompanyModel company = event.company;
          company.images = images;
          final data = event.company.toJson();

          final postData = {"code": 'company', "body": jsonEncode(data)};

          await jsonRepo.updateSetting(event.guid, postData);
          emit(CompanyUpdateSuccess());
        } else {
          emit(const CompanyUpdateFailed(message: 'image upload failed'));
        }
      } else {
        emit(const CompanyUpdateFailed(message: 'no image found'));
      }
    } catch (e) {
      emit(CompanyUpdateFailed(message: e.toString()));
    }
  }
}
