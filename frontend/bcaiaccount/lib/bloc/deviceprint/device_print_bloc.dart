import 'dart:convert';

import 'package:smlaicloud/repositories/json_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/device_printer_model.dart';

part 'device_print_event.dart';
part 'device_print_state.dart';

class DevicePrintBloc extends Bloc<DevicePrintEvent, DevicePrintState> {
  final JsonRepository _jsonRepository;

  DevicePrintBloc({required JsonRepository jsonRepository})
      : _jsonRepository = jsonRepository,
        super(DevicePrintInitial()) {
    on<DevicePrintLoadList>(onDevicePrintLoad);
    on<DevicePrintSave>(onDevicePrintSave);
    on<DevicePrintUpdate>(onDevicePrintUpdate);
    on<DevicePrintDelete>(onDevicePrintDelete);
    on<DevicePrintDeleteMany>(onDevicePrintDeleteMany);
  }

  void onDevicePrintLoad(
      DevicePrintLoadList event, Emitter<DevicePrintState> emit) async {
    emit(DevicePrintInProgress());

    try {
      final results =
          await _jsonRepository.getSetting('deviceprinter', event.search);

      if (results.success) {
        if (results.data.length > 0) {
          List<DevicePrinterSaveModel> devicePrints = [];
          for (int i = 0; i < results.data.length; i++) {
            DevicePrinterSaveModel devicePrint =
                DevicePrinterSaveModel.fromJson(
                    json.decode(results.data[i]['body']));
            devicePrint.guidfixed = results.data[i]['guidfixed'];
            devicePrints.add(devicePrint);
          }
          // print(devicePrints);
          emit(DevicePrintLoadSuccess(devicePrints: devicePrints));
        } else {
          List<DevicePrinterSaveModel> devicePrints = [];
          emit(DevicePrintLoadSuccess(devicePrints: devicePrints));
        }
      } else {
        emit(const DevicePrintLoadFailed(message: 'DevicePrint Not Found'));
      }
    } catch (e) {
      emit(DevicePrintLoadFailed(message: e.toString()));
    }
  }

  void onDevicePrintSave(
      DevicePrintSave event, Emitter<DevicePrintState> emit) async {
    emit(DevicePrintSaveInProgress());

    List<String> printers = [];
    printers.add(event.devicePrint.primary);
    printers.add(event.devicePrint.spare);
    DevicePrinterSaveModel devicePrinterSave = DevicePrinterSaveModel(
        guidfixed: "",
        guiddevice: event.devicePrint.guiddevice,
        devicecode: event.devicePrint.devicecode,
        devicename: event.devicePrint.devicename,
        printers: printers);
    try {
      final postData = {
        "code": 'deviceprinter',
        "body": jsonEncode(devicePrinterSave)
      };
      // print(postData);
      await _jsonRepository.saveSetting(postData);
      emit(DevicePrintSaveSuccess());
    } catch (e) {
      emit(DevicePrintSaveFailed(message: e.toString()));
    }
  }

  void onDevicePrintUpdate(
      DevicePrintUpdate event, Emitter<DevicePrintState> emit) async {
    emit(DevicePrintUpdateInProgress());
    try {
      List<String> printers = [];
      printers.add(event.devicePrint.primary);
      printers.add(event.devicePrint.spare);
      DevicePrinterSaveModel deviceprinterSave = DevicePrinterSaveModel(
          guidfixed: "",
          guiddevice: event.devicePrint.guiddevice,
          devicecode: event.devicePrint.devicecode,
          devicename: event.devicePrint.devicename,
          printers: printers);

      final postData = {
        "code": 'deviceprinter',
        "body": jsonEncode(deviceprinterSave)
      };
      // print(postData);
      await _jsonRepository.updateSetting(event.guid, postData);
      emit(DevicePrintUpdateSuccess());
    } catch (e) {
      emit(DevicePrintUpdateFailed(message: e.toString()));
    }
  }

  void onDevicePrintDelete(
      DevicePrintDelete event, Emitter<DevicePrintState> emit) async {
    emit(DevicePrintDeleteInProgress());
    try {
      await _jsonRepository.deleteSetting(event.guid);

      emit(DevicePrintDeleteSuccess());
    } catch (e) {
      emit(DevicePrintDeleteFailed());
    }
  }

  void onDevicePrintDeleteMany(
      DevicePrintDeleteMany event, Emitter<DevicePrintState> emit) async {
    emit(DevicePrintDeleteManyInProgress());
    try {
      await _jsonRepository.deleteManySetting(event.guid);

      emit(DevicePrintDeleteManySuccess());
    } catch (e) {
      emit(DevicePrintDeleteManyFailed());
    }
  }
}
