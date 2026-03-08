import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/master_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/json_repository.dart';
import 'package:smlaicloud/repositories/qrpayment_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/qr_model.dart';

part 'qr_event.dart';
part 'qr_state.dart';

class QrBloc extends Bloc<QrEvent, QrState> {
  final QrPaymentRepository _qrRepository;
  final JsonRepository _jsonRepository;
  QrBloc(
      {required QrPaymentRepository qrRepository,
      required JsonRepository jsonRepository})
      : _qrRepository = qrRepository,
        _jsonRepository = jsonRepository,
        super(QrInitial()) {
    on<QrLoadList>(onQrLoad);
    on<QrSave>(onQrSave);
    on<QrBulkSave>(onQrBulkSave);
    on<QrSaveWithImage>(onQrSaveWithImage);
    on<QrUpdate>(onQrUpdate);
    on<QrWithImageUpdate>(onQrWithImageUpdate);
    on<QrDelete>(onQrDelete);
    on<QrDeleteMany>(onQrDeleteMany);
    on<QrGet>(onQrGet);
  }

  void onQrLoad(QrLoadList event, Emitter<QrState> emit) async {
    emit(QrInProgress());

    try {
      final results = await _qrRepository.getQrPaymentList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<QrModel> qrs =
            (results.data as List).map((qr) => QrModel.fromJson(qr)).toList();
        emit(QrLoadSuccess(qrs: qrs));
      } else {
        emit(const QrLoadFailed(message: 'Qr Not Found'));
      }
    } catch (e) {
      emit(QrLoadFailed(message: e.toString()));
    }
  }

  void onQrDelete(QrDelete event, Emitter<QrState> emit) async {
    emit(QrDeleteInProgress());
    try {
      await _qrRepository.deleteQrPayment(event.guid);

      emit(QrDeleteSuccess());
    } catch (e) {
      // emit(QrDeleteFailure(message: e.toString()));
    }
  }

  void onQrDeleteMany(QrDeleteMany event, Emitter<QrState> emit) async {
    emit(QrDeleteManyInProgress());
    try {
      await _qrRepository.deleteQrPaymentMany(event.guid);

      emit(QrDeleteManySuccess());
    } catch (e) {
      // emit(QrDeleteFailure(message: e.toString()));
    }
  }

  void onQrSave(QrSave event, Emitter<QrState> emit) async {
    emit(QrSaveInProgress());
    try {
      await _qrRepository.saveQrPayment(event.qr);
      emit(QrSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(QrSaveFailed(message: error['message']));
    }
  }

  void onQrSaveWithImage(QrSaveWithImage event, Emitter<QrState> emit) async {
    emit(QrSaveInProgress());
    try {
      ImagesModel images = ImagesModel(uri: "", xorder: 0);
      if (event.imageWeb.isNotEmpty) {
        ApiResponse result =
            await _jsonRepository.uploadImage(event.imageFile, event.imageWeb);
        if (result.success) {
          UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
          images = ImagesModel(uri: uploadImage.uri, xorder: 0);
        } else {
          emit(QrSaveFailed(message: result.message));
        }

        if (images.uri != "") {
          QrModel qrModel = event.qr;
          qrModel.logo = images.uri;
          // print(QrModel);
          await _qrRepository.saveQrPayment(event.qr);
          emit(QrSaveSuccess());
        } else {
          emit(const QrSaveFailed(message: 'image upload failed'));
        }
      } else {
        emit(const QrSaveFailed(message: 'no image found'));
      }
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(QrSaveFailed(message: error['message']));
    }
  }

  void onQrWithImageUpdate(
      QrWithImageUpdate event, Emitter<QrState> emit) async {
    emit(QrUpdateInProgress());
    try {
      ImagesModel images = ImagesModel(uri: '', xorder: 0);

      if (event.imageWeb.isNotEmpty) {
        ApiResponse result =
            await _jsonRepository.uploadImage(event.imageFile, event.imageWeb);
        if (result.success) {
          UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
          images = ImagesModel(uri: uploadImage.uri, xorder: 0);
        } else {
          emit(QrUpdateFailed(message: result.message));
        }
      }

      if (images.uri != '') {
        QrModel qrModel = event.qr;
        qrModel.logo = images.uri;

        // print(QrModel);
        await _qrRepository.updateQrPayment(event.guid, event.qr);
        emit(QrUpdateSuccess());
      } else {
        emit(const QrUpdateFailed(message: 'image upload failed'));
      }
    } catch (e) {
      emit(QrUpdateFailed(message: e.toString()));
    }
  }

  void onQrUpdate(QrUpdate event, Emitter<QrState> emit) async {
    emit(QrUpdateInProgress());
    try {
      await _qrRepository.updateQrPayment(event.guid, event.qrModel);
      emit(QrUpdateSuccess());
    } catch (e) {
      emit(QrUpdateFailed(message: e.toString()));
    }
  }

  void onQrGet(QrGet event, Emitter<QrState> emit) async {
    emit(QrGetInProgress());
    try {
      final result = await _qrRepository.getQrPayment(event.guid);
      if (result.success) {
        QrModel qr = QrModel.fromJson(result.data);
        emit(QrGetSuccess(qrs: qr));
      } else {
        emit(const QrGetFailed(message: 'Qr Not Found'));
      }
    } catch (e) {
      // emit(QrDeleteFailure(message: e.toString()));
    }
  }

  void onQrBulkSave(QrBulkSave event, Emitter<QrState> emit) async {
    emit(QrSaveInProgress());
    try {
      await _qrRepository.saveQrPaymentBulk(event.qrs);
      emit(QrSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(QrSaveFailed(message: error['message']));
    }
  }
}
