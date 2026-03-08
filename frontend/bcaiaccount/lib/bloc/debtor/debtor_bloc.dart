import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'dart:convert';

import 'package:smlaicloud/model/debtor_model.dart';
import 'package:smlaicloud/model/point_transaction_model.dart';
import 'package:smlaicloud/model/bulk_points_request_model.dart';
import 'package:smlaicloud/repositories/debtor_repository.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/global_model.dart';
import '../../global.dart' as global;
import 'package:smlaicloud/model/master_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'dart:io';

part 'debtor_event.dart';
part 'debtor_state.dart';

class DebtorBloc extends Bloc<DebtorEvent, DebtorState> {
  final DebtorRepository _debtorRepository;

  DebtorBloc({required DebtorRepository debtorRepository})
    : _debtorRepository = debtorRepository,
      super(DebtorInitial()) {
    on<DebtorLoadList>(onDebtorLoad);
    on<DebtorSave>(onDebtorSave);
    on<DebtorBulkSave>(onDebtorBulkSave);
    on<DebtorWithImageSave>(onDebtorWithImageSave);
    on<DebtorUpdate>(onDebtorUpdate);
    on<DebtorWithImageUpdate>(onDebtorWithImageUpdate);
    on<DebtorDelete>(onDebtorDelete);
    on<DebtorDeleteMany>(onDebtorDeleteMany);
    on<DebtorGet>(onDebtorGet);
    on<DebtorGetBycode>(onDebtorGetBycode);
    on<DebtorAddPoint>(onDebtorAddPoint);
    on<DebtorBulkAddPoints>(onDebtorBulkAddPoints);
  }

  void onDebtorLoad(DebtorLoadList event, Emitter<DebtorState> emit) async {
    emit(DebtorInProgress());

    try {
      final results = await _debtorRepository.getDebtorList(
        offset: event.offset,
        limit: event.limit,
        search: event.search,
        groups: event.groups,
      );

      if (results.success) {
        List<DebtorModel> debtors = (results.data as List)
            .map((debtor) => DebtorModel.fromJson(debtor))
            .toList();
        emit(DebtorLoadSuccess(debtors: debtors));
      } else {
        emit(const DebtorLoadFailed(message: 'Debtor Not Found'));
      }
    } catch (e) {
      emit(DebtorLoadFailed(message: e.toString()));
    }
  }

  void onDebtorDelete(DebtorDelete event, Emitter<DebtorState> emit) async {
    emit(DebtorDeleteInProgress());
    try {
      await _debtorRepository.deleteDebtor(event.guid);

      emit(DebtorDeleteSuccess());
    } catch (e) {
      // emit(DebtorDeleteFailure(message: e.toString()));
    }
  }

  void onDebtorDeleteMany(
    DebtorDeleteMany event,
    Emitter<DebtorState> emit,
  ) async {
    emit(DebtorDeleteManyInProgress());
    try {
      await _debtorRepository.deleteDebtorMany(event.guid);

      emit(DebtorDeleteManySuccess());
    } catch (e) {
      // emit(DebtorDeleteFailure(message: e.toString()));
    }
  }

  void onDebtorSave(DebtorSave event, Emitter<DebtorState> emit) async {
    emit(DebtorSaveInProgress());
    try {
      DebtorModel debtor = event.debtor;

      DebtorRequestModel debtorRequestModel = DebtorRequestModel(
        guidfixed: debtor.guidfixed,
        code: debtor.code,
        names: debtor.names,
        customertype: debtor.customertype,
        branchnumber: debtor.branchnumber,
        personaltype: debtor.personaltype,
        ismember: debtor.ismember,
        addressforbilling: debtor.addressforbilling,
        addressforshipping: debtor.addressforshipping,
        images: debtor.images,
        taxid: debtor.taxid,
        email: debtor.email,
        creditday: debtor.creditday,
        fundcode: debtor.fundcode,
        pointscode: debtor.pointscode,
        pricelevel: debtor.pricelevel,
      );

      for (var element in debtor.groups) {
        debtorRequestModel.groups.add(element.guidfixed);
      }

      await _debtorRepository.saveDebtor(debtorRequestModel);
      emit(DebtorSaveSuccess());
    } catch (e) {
      emit(DebtorSaveFailed(message: e.toString()));
    }
  }

  void onDebtorBulkSave(DebtorBulkSave event, Emitter<DebtorState> emit) async {
    emit(DebtorBulkSaveInProgress());
    try {
      final result = await _debtorRepository.saveBulkDebtors(event.debtors);
      if (result.success) {
        emit(DebtorBulkSaveSuccess(savedCount: event.debtors.length));
      } else {
        emit(DebtorBulkSaveFailed(message: result.message));
      }
    } catch (e) {
      emit(DebtorBulkSaveFailed(message: e.toString()));
    }
  }

  void onDebtorWithImageSave(
    DebtorWithImageSave event,
    Emitter<DebtorState> emit,
  ) async {
    emit(DebtorSaveInProgress());
    try {
      List<ImagesModel> images = [];
      if (event.imageFile.isNotEmpty) {
        for (int i = 0; i < event.imageFile.length; i++) {
          if (event.imageFile[i].uri.toString() != '') {
            ApiResponse result = await _debtorRepository.uploadImage(
              event.imageFile[i],
              event.imageWeb[i],
            );
            if (result.success) {
              UploadImageModel uploadImage = UploadImageModel.fromJson(
                result.data,
              );
              images.add(ImagesModel(uri: uploadImage.uri, xorder: i));
            } else {
              emit(DebtorSaveFailed(message: result.message));
            }
          }
        }

        if (images.isNotEmpty) {
          DebtorModel debtor = event.debtor;

          DebtorRequestModel debtorRequestModel = DebtorRequestModel(
            guidfixed: debtor.guidfixed,
            code: debtor.code,
            names: debtor.names,
            customertype: debtor.customertype,
            branchnumber: debtor.branchnumber,
            personaltype: debtor.personaltype,
            ismember: debtor.ismember,
            addressforbilling: debtor.addressforbilling,
            addressforshipping: debtor.addressforshipping,
            images: images,
            taxid: debtor.taxid,
            email: debtor.email,
            creditday: debtor.creditday,
            fundcode: debtor.fundcode,
            pointscode: debtor.pointscode,
            pricelevel: debtor.pricelevel,
          );

          for (var element in debtor.groups) {
            debtorRequestModel.groups.add(element.guidfixed);
          }

          await _debtorRepository.saveDebtor(debtorRequestModel);
          emit(DebtorSaveSuccess());
        } else {
          emit(const DebtorSaveFailed(message: 'image upload failed'));
        }
      } else {
        emit(const DebtorSaveFailed(message: 'no image found'));
      }
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(DebtorSaveFailed(message: error['message']));
    }
  }

  void onDebtorUpdate(DebtorUpdate event, Emitter<DebtorState> emit) async {
    emit(DebtorUpdateInProgress());
    try {
      DebtorModel debtor = event.debtorModel;

      DebtorRequestModel debtorRequestModel = DebtorRequestModel(
        guidfixed: debtor.guidfixed,
        code: debtor.code,
        names: debtor.names,
        customertype: debtor.customertype,
        branchnumber: debtor.branchnumber,
        personaltype: debtor.personaltype,
        ismember: debtor.ismember,
        addressforbilling: debtor.addressforbilling,
        addressforshipping: debtor.addressforshipping,
        images: debtor.images,
        taxid: debtor.taxid,
        email: debtor.email,
        creditday: debtor.creditday,
        fundcode: debtor.fundcode,
        pointscode: debtor.pointscode,
        pricelevel: debtor.pricelevel,
      );

      for (var element in debtor.groups) {
        debtorRequestModel.groups.add(element.guidfixed);
      }

      await _debtorRepository.updateDebtor(event.guid, debtorRequestModel);
      emit(DebtorUpdateSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(DebtorUpdateFailed(message: error['message']));
    }
  }

  void onDebtorWithImageUpdate(
    DebtorWithImageUpdate event,
    Emitter<DebtorState> emit,
  ) async {
    emit(DebtorUpdateInProgress());
    try {
      List<ImagesModel> images = [];
      if (event.imagesUris.isNotEmpty) {
        for (int i = 0; i < event.imagesUris.length; i++) {
          if (event.imageWeb[i].isNotEmpty) {
            ApiResponse result = await _debtorRepository.uploadImage(
              event.imageFiles[i],
              event.imageWeb[i],
            );
            if (result.success) {
              UploadImageModel uploadImage = UploadImageModel.fromJson(
                result.data,
              );
              images.add(ImagesModel(uri: uploadImage.uri, xorder: i));
            } else {
              emit(DebtorUpdateFailed(message: result.message));
            }
          } else if (event.imagesUris[i].uri != '') {
            images.add(ImagesModel(uri: event.imagesUris[i].uri, xorder: i));
          }
        }

        if (images.isNotEmpty) {
          DebtorModel debtor = event.debtor;

          DebtorRequestModel debtorRequestModel = DebtorRequestModel(
            guidfixed: debtor.guidfixed,
            code: debtor.code,
            names: debtor.names,
            customertype: debtor.customertype,
            branchnumber: debtor.branchnumber,
            personaltype: debtor.personaltype,
            ismember: debtor.ismember,
            addressforbilling: debtor.addressforbilling,
            addressforshipping: debtor.addressforshipping,
            images: images,
            taxid: debtor.taxid,
            email: debtor.email,
            creditday: debtor.creditday,
            fundcode: debtor.fundcode,
            pointscode: debtor.pointscode,
            pricelevel: debtor.pricelevel,
          );

          for (var element in debtor.groups) {
            debtorRequestModel.groups.add(element.guidfixed);
          }

          await _debtorRepository.updateDebtor(event.guid, debtorRequestModel);
          emit(DebtorUpdateSuccess());
        } else {
          emit(const DebtorUpdateFailed(message: 'image upload failed'));
        }
      } else {
        emit(const DebtorUpdateFailed(message: 'no image found'));
      }
    } catch (e) {
      emit(DebtorUpdateFailed(message: e.toString()));
    }
  }

  void onDebtorGet(DebtorGet event, Emitter<DebtorState> emit) async {
    emit(DebtorGetInProgress());
    try {
      final result = await _debtorRepository.getDebtor(event.guid);
      if (result.success) {
        DebtorModel debtor = DebtorModel.fromJson(result.data);
        // print(debtor.toJson());
        emit(DebtorGetSuccess(debtors: debtor));
      } else {
        emit(const DebtorGetFailed(message: 'Debtor Not Found'));
      }
    } catch (e) {
      // emit(DebtorDeleteFailure(message: e.toString()));
    }
  }

  void onDebtorGetBycode(
    DebtorGetBycode event,
    Emitter<DebtorState> emit,
  ) async {
    emit(DebtorGetBycodeInProgress());
    try {
      final result = await _debtorRepository.getDebtorByCode(event.custcode);
      if (result.success) {
        DebtorModel debtor = DebtorModel.fromJson(result.data);
        emit(DebtorGetBycodeSuccess(debtors: debtor));
      } else {
        emit(const DebtorGetBycodeFailed(message: 'Debtor Not Found'));
      }
    } catch (e) {
      emit(DebtorGetBycodeFailed(message: e.toString()));
    }
  }

  void onDebtorAddPoint(DebtorAddPoint event, Emitter<DebtorState> emit) async {
    emit(DebtorAddPointInProgress());
    try {
      final result = await _debtorRepository.addPointToDebtor(
        event.pointscode,
        event.pointTransaction,
      );
      if (result.success) {
        emit(DebtorAddPointSuccess());
      } else {
        emit(
          DebtorAddPointFailed(
            message: result.message.isNotEmpty
                ? result.message
                : 'เพิ่มแต้มไม่สำเร็จ',
          ),
        );
      }
    } catch (e) {
      emit(DebtorAddPointFailed(message: e.toString()));
    }
  }

  void onDebtorBulkAddPoints(
    DebtorBulkAddPoints event,
    Emitter<DebtorState> emit,
  ) async {
    emit(DebtorBulkAddPointsInProgress());

    AppLogger.debug(
      '🔄 Bloc: Processing bulk add points - ${event.pointsList.length} items',
    );

    try {
      final result = await _debtorRepository.bulkAddPoints(event.pointsList);

      AppLogger.debug(
        '📊 Bulk add points result: success=${result.success}, message=${result.message}',
      );
      if (result.data != null) {
        AppLogger.debug(
          '   Success: ${result.data!.success}, Failed: ${result.data!.failed}, Total: ${result.data!.total}',
        );
      }

      if (result.success && result.data != null) {
        emit(
          DebtorBulkAddPointsSuccess(
            successCount: result.data!.success,
            failedCount: result.data!.failed,
            totalCount: result.data!.total,
            successItems: result.data!.successItems,
            failedItems: result.data!.failedItems,
          ),
        );
      } else {
        emit(
          DebtorBulkAddPointsFailed(
            message: result.message.isNotEmpty
                ? result.message
                : 'นำเข้าแต้มไม่สำเร็จ',
          ),
        );
      }
    } catch (e) {
      AppLogger.err('Error in bulk add points', err: e);
      emit(DebtorBulkAddPointsFailed(message: e.toString()));
    }
  }
}
