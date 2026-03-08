import 'dart:convert';

import 'package:smlaicloud/model/order_setting_model.dart';
import 'package:smlaicloud/repositories/order_setting_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'order_setting_event.dart';
part 'order_setting_state.dart';

class OrderSettingBloc extends Bloc<OrderSettingEvent, OrderSettingState> {
  final OrderSettingRepository _orderSettingRepository;

  OrderSettingBloc({required OrderSettingRepository orderSettingRepository})
      : _orderSettingRepository = orderSettingRepository,
        super(OrderSettingInitial()) {
    on<OrderSettingLoadList>(onOrderSettingLoad);
    on<OrderSettingSave>(onOrderSettingSave);
    on<OrderSettingUpdate>(onOrderSettingUpdate);
    on<OrderSettingDelete>(onOrderSettingDelete);
    on<OrderSettingDeleteMany>(onOrderSettingDeleteMany);
    on<OrderSettingGet>(onOrderSettingGet);
  }

  void onOrderSettingLoad(
      OrderSettingLoadList event, Emitter<OrderSettingState> emit) async {
    emit(OrderSettingInProgress());

    try {
      final results = await _orderSettingRepository.getOrderList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<OrderSettingModel> orderSettings = (results.data as List)
            .map((orderSettings) => OrderSettingModel.fromJson(orderSettings))
            .toList();
        emit(OrderSettingLoadSuccess(orderSettings: orderSettings));
      } else {
        emit(const OrderSettingLoadFailed(message: 'Pos Settings Not Found'));
      }
    } catch (e) {
      emit(OrderSettingLoadFailed(message: e.toString()));
    }
  }

  void onOrderSettingDelete(
      OrderSettingDelete event, Emitter<OrderSettingState> emit) async {
    emit(OrderSettingDeleteInProgress());
    try {
      await _orderSettingRepository.deleteOrder(event.guid);

      emit(OrderSettingDeleteSuccess());
    } catch (e) {
      emit(OrderSettingDeleteFailed());
    }
  }

  void onOrderSettingDeleteMany(
      OrderSettingDeleteMany event, Emitter<OrderSettingState> emit) async {
    emit(OrderSettingDeleteManyInProgress());
    try {
      await _orderSettingRepository.deleteOrderMany(event.guid);

      emit(OrderSettingDeleteManySuccess());
    } catch (e) {
      emit(OrderSettingDeleteFailed());
    }
  }

  void onOrderSettingSave(
      OrderSettingSave event, Emitter<OrderSettingState> emit) async {
    emit(OrderSettingSaveInProgress());
    try {
      await _orderSettingRepository.saveOrder(event.orderSetting);
      emit(OrderSettingSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(OrderSettingSaveFailed(message: error['message']));
    }
  }

  void onOrderSettingUpdate(
      OrderSettingUpdate event, Emitter<OrderSettingState> emit) async {
    emit(OrderSettingUpdateInProgress());
    try {
      await _orderSettingRepository.updateOrder(event.guid, event.orderSetting);
      emit(OrderSettingUpdateSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(OrderSettingUpdateFailed(message: error['message']));
    }
  }

  void onOrderSettingGet(
      OrderSettingGet event, Emitter<OrderSettingState> emit) async {
    emit(OrderSettingGetInProgress());
    try {
      final result = await _orderSettingRepository.getOrder(event.guid);
      if (result.success) {
        OrderSettingModel orderSetting =
            OrderSettingModel.fromJson(result.data);
        emit(OrderSettingGetSuccess(orderSettings: orderSetting));
      } else {
        emit(const OrderSettingGetFailed(message: 'OrderSetting Not Found'));
      }
    } catch (e) {
      emit(OrderSettingGetFailed(message: e.toString()));
    }
  }
}
