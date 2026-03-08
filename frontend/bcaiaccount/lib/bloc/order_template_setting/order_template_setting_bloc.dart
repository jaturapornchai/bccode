import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:smlaicloud/model/master_model.dart';
import 'package:smlaicloud/model/order_template_setting_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/order_template_setting_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'order_template_setting_event.dart';
part 'order_template_setting_state.dart';

class OrderTemplateSettingBloc
    extends Bloc<OrderTemplateSettingEvent, OrderTemplateSettingState> {
  final OrderTemplateSettingRepository _orderTemplateSettingRepository;

  OrderTemplateSettingBloc(
      {required OrderTemplateSettingRepository orderTemplateSettingRepository})
      : _orderTemplateSettingRepository = orderTemplateSettingRepository,
        super(OrderTemplateSettingInitial()) {
    on<OrderTemplateSettingLoadList>(onOrderTemplateSettingLoad);
    on<OrderTemplateSettingSave>(onOrderTemplateSettingSave);
    on<OrderTemplateSettingUpdate>(onOrderTemplateSettingUpdate);
    on<OrderTemplateSettingDelete>(onOrderTemplateSettingDelete);
    on<OrderTemplateSettingDeleteMany>(onOrderTemplateSettingDeleteMany);
    on<OrderTemplateSettingGet>(onOrderTemplateSettingGet);
    on<OrderTemplateSettingWithImageSave>(onOrderTemplateSettingWithImageSave);
    on<OrderTemplateSettingWithImageUpdate>(
        onOrderTemplateSettingWithImageUpdate);
  }

  void onOrderTemplateSettingLoad(OrderTemplateSettingLoadList event,
      Emitter<OrderTemplateSettingState> emit) async {
    emit(OrderTemplateSettingInProgress());

    try {
      final results = await _orderTemplateSettingRepository.getOrderList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<OrderTemplateSettingModel> orderTemplateSettings =
            (results.data as List)
                .map((orderTemplateSettings) =>
                    OrderTemplateSettingModel.fromJson(orderTemplateSettings))
                .toList();
        emit(OrderTemplateSettingLoadSuccess(
            orderTemplateSettings: orderTemplateSettings));
      } else {
        emit(const OrderTemplateSettingLoadFailed(
            message: 'Order Station Settings Not Found'));
      }
    } catch (e) {
      emit(OrderTemplateSettingLoadFailed(message: e.toString()));
    }
  }

  void onOrderTemplateSettingDelete(OrderTemplateSettingDelete event,
      Emitter<OrderTemplateSettingState> emit) async {
    emit(OrderTemplateSettingDeleteInProgress());
    try {
      await _orderTemplateSettingRepository.deleteOrder(event.guid);
      emit(OrderTemplateSettingDeleteSuccess());
    } catch (e) {
      emit(OrderTemplateSettingDeleteFailed());
    }
  }

  void onOrderTemplateSettingDeleteMany(OrderTemplateSettingDeleteMany event,
      Emitter<OrderTemplateSettingState> emit) async {
    emit(OrderTemplateSettingDeleteManyInProgress());
    try {
      await _orderTemplateSettingRepository.deleteOrderMany(event.guid);
      emit(OrderTemplateSettingDeleteManySuccess());
    } catch (e) {
      emit(OrderTemplateSettingDeleteFailed());
    }
  }

  void onOrderTemplateSettingSave(OrderTemplateSettingSave event,
      Emitter<OrderTemplateSettingState> emit) async {
    emit(OrderTemplateSettingSaveInProgress());
    try {
      await _orderTemplateSettingRepository
          .saveOrder(event.orderTemplateSetting);
      emit(OrderTemplateSettingSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(OrderTemplateSettingSaveFailed(message: error['message']));
    }
  }

  void onOrderTemplateSettingUpdate(OrderTemplateSettingUpdate event,
      Emitter<OrderTemplateSettingState> emit) async {
    emit(OrderTemplateSettingUpdateInProgress());
    try {
      await _orderTemplateSettingRepository.updateOrder(
          event.guid, event.orderTemplateSetting);
      emit(OrderTemplateSettingUpdateSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(OrderTemplateSettingUpdateFailed(message: error['message']));
    }
  }

  void onOrderTemplateSettingGet(OrderTemplateSettingGet event,
      Emitter<OrderTemplateSettingState> emit) async {
    emit(OrderTemplateSettingGetInProgress());
    try {
      final result = await _orderTemplateSettingRepository.getOrder(event.guid);
      if (result.success) {
        OrderTemplateSettingModel orderTemplateSetting =
            OrderTemplateSettingModel.fromJson(result.data);
        emit(OrderTemplateSettingGetSuccess(
            orderTemplateSettings: orderTemplateSetting));
      } else {
        emit(const OrderTemplateSettingGetFailed(
            message: 'Order Template Setting Not Found'));
      }
    } catch (e) {
      emit(OrderTemplateSettingGetFailed(message: e.toString()));
    }
  }

  void onOrderTemplateSettingWithImageSave(
      OrderTemplateSettingWithImageSave event,
      Emitter<OrderTemplateSettingState> emit) async {
    emit(OrderTemplateSettingSaveInProgress());
    try {
      ApiResponse result = await _orderTemplateSettingRepository.uploadImage(
          event.imageFile, event.imageWeb!);
      if (result.success) {
        UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
        OrderTemplateSettingModel orderTemplateSettingModel =
            event.orderTemplateSetting;
        orderTemplateSettingModel.lineoaimg = uploadImage.uri;
        await _orderTemplateSettingRepository
            .saveOrder(orderTemplateSettingModel);
        emit(OrderTemplateSettingSaveSuccess());
      } else {
        emit(OrderTemplateSettingSaveFailed(message: result.message));
      }
    } catch (e) {
      emit(OrderTemplateSettingSaveFailed(message: e.toString()));
    }
  }

  void onOrderTemplateSettingWithImageUpdate(
      OrderTemplateSettingWithImageUpdate event,
      Emitter<OrderTemplateSettingState> emit) async {
    emit(OrderTemplateSettingUpdateInProgress());
    try {
      ApiResponse result = await _orderTemplateSettingRepository.uploadImage(
          event.imageFile, event.imageWeb);
      if (result.success) {
        UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
        OrderTemplateSettingModel orderTemplateSettingModel =
            event.orderTemplateSetting;
        orderTemplateSettingModel.lineoaimg = uploadImage.uri;
        await _orderTemplateSettingRepository.updateOrder(
            event.guid, orderTemplateSettingModel);
        emit(OrderTemplateSettingUpdateSuccess());
      } else {
        emit(OrderTemplateSettingUpdateFailed(message: result.message));
      }
    } catch (e) {
      emit(OrderTemplateSettingUpdateFailed(message: e.toString()));
    }
  }
}
