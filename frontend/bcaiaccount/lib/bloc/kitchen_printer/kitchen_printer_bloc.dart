import 'dart:convert';

import 'package:smlaicloud/repositories/json_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/kitchen_printer_model.dart';

part 'kitchen_printer_event.dart';
part 'kitchen_printer_state.dart';

class KitchenPrinterBloc
    extends Bloc<KitchenPrinterEvent, KitchenPrinterState> {
  final JsonRepository _jsonRepository;

  KitchenPrinterBloc({required JsonRepository jsonRepository})
      : _jsonRepository = jsonRepository,
        super(KitchenPrinterInitial()) {
    on<KitchenPrinterLoadList>(onKitchenPrinterLoad);
    on<KitchenPrinterSave>(onKitchenPrinterSave);
    on<KitchenPrinterUpdate>(onKitchenPrinterUpdate);
    on<KitchenPrinterDelete>(onKitchenPrinterDelete);
    on<KitchenPrinterDeleteMany>(onKitchenPrinterDeleteMany);
  }

  void onKitchenPrinterLoad(
      KitchenPrinterLoadList event, Emitter<KitchenPrinterState> emit) async {
    emit(KitchenPrinterInProgress());

    try {
      final results =
          await _jsonRepository.getSetting('kitchenprinter', event.search);

      if (results.success) {
        if (results.data.length > 0) {
          List<KitchenPrinterSaveModel> kitchenPrinterList = [];

          for (int i = 0; i < results.data.length; i++) {
            KitchenPrinterSaveModel kitchenPrinter =
                KitchenPrinterSaveModel.fromJson(
                    json.decode(results.data[i]['body']));
            kitchenPrinter.guidfixed = results.data[i]['guidfixed'];
            kitchenPrinterList.add(kitchenPrinter);
          }
          // print(kitchenPrinterList);
          emit(KitchenPrinterLoadSuccess(kitchenPrinters: kitchenPrinterList));
        } else {
          List<KitchenPrinterSaveModel> kitchenPrinterList = [];
          emit(KitchenPrinterLoadSuccess(kitchenPrinters: kitchenPrinterList));
        }
      } else {
        emit(const KitchenPrinterLoadFailed(
            message: 'KitchenPrinter Not Found'));
      }
    } catch (e) {
      emit(KitchenPrinterLoadFailed(message: e.toString()));
    }
  }

  void onKitchenPrinterSave(
      KitchenPrinterSave event, Emitter<KitchenPrinterState> emit) async {
    emit(KitchenPrinterSaveInProgress());

    List<String> printers = [];
    printers.add(event.kitchenPrinter.primary);
    printers.add(event.kitchenPrinter.spare);
    KitchenPrinterSaveModel kitchenPrinterSave = KitchenPrinterSaveModel(
        guidfixed: "",
        guidkitchen: event.kitchenPrinter.guidkitchen,
        kitchencode: event.kitchenPrinter.kitchencode,
        kitchenname: event.kitchenPrinter.kitchenname,
        printers: printers);
    try {
      final postData = {
        "code": 'kitchenprinter',
        "body": jsonEncode(kitchenPrinterSave)
      };
      // print(postData);
      await _jsonRepository.saveSetting(postData);
      emit(KitchenPrinterSaveSuccess());
    } catch (e) {
      emit(KitchenPrinterSaveFailed(message: e.toString()));
    }
  }

  void onKitchenPrinterUpdate(
      KitchenPrinterUpdate event, Emitter<KitchenPrinterState> emit) async {
    emit(KitchenPrinterUpdateInProgress());
    try {
      List<String> printers = [];
      printers.add(event.kitchenPrinter.primary);
      printers.add(event.kitchenPrinter.spare);
      KitchenPrinterSaveModel kitchenPrinterSave = KitchenPrinterSaveModel(
          guidfixed: "",
          guidkitchen: event.kitchenPrinter.guidkitchen,
          kitchencode: event.kitchenPrinter.kitchencode,
          kitchenname: event.kitchenPrinter.kitchenname,
          printers: printers);

      final postData = {
        "code": 'kitchenprinter',
        "body": jsonEncode(kitchenPrinterSave)
      };
      // print(postData);
      await _jsonRepository.updateSetting(event.guid, postData);
      emit(KitchenPrinterUpdateSuccess());
    } catch (e) {
      emit(KitchenPrinterUpdateFailed(message: e.toString()));
    }
  }

  void onKitchenPrinterDelete(
      KitchenPrinterDelete event, Emitter<KitchenPrinterState> emit) async {
    emit(KitchenPrinterDeleteInProgress());
    try {
      await _jsonRepository.deleteSetting(event.guid);

      emit(KitchenPrinterDeleteSuccess());
    } catch (e) {
      emit(KitchenPrinterDeleteFailed());
    }
  }

  void onKitchenPrinterDeleteMany(
      KitchenPrinterDeleteMany event, Emitter<KitchenPrinterState> emit) async {
    emit(KitchenPrinterDeleteManyInProgress());
    try {
      await _jsonRepository.deleteManySetting(event.guid);

      emit(KitchenPrinterDeleteManySuccess());
    } catch (e) {
      emit(KitchenPrinterDeleteManyFailed());
    }
  }
}
