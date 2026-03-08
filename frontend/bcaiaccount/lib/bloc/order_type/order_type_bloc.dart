import 'dart:convert';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/order_type_model.dart';
import 'package:smlaicloud/repositories/order_type_repository.dart';

part 'order_type_event.dart';
part 'order_type_state.dart';

class OrderTypeBloc extends Bloc<OrderTypeEvent, OrderTypeState> {
  final OrderTypeRepository _orderTypeRepository;

  OrderTypeBloc({required OrderTypeRepository orderTypeRepository})
      : _orderTypeRepository = orderTypeRepository,
        super(OrderTypeInitial()) {
    on<OrderTypeLoadList>(onOrderTypeLoad);
    on<OrderTypeSave>(onOrderTypeSave);
    on<OrderTypeUpdate>(onOrderTypeUpdate);
    on<OrderTypeDelete>(onOrderTypeDelete);
    on<OrderTypeDeleteMany>(onOrderTypeDeleteMany);
    on<OrderTypeGet>(onOrderTypeGet);
  }

  void onOrderTypeLoad(
      OrderTypeLoadList event, Emitter<OrderTypeState> emit) async {
    emit(OrderTypeInProgress());
    try {
      final results = await _orderTypeRepository.getOrderTypeList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<OrderTypeModel> ordertype = (results.data as List)
            .map((ordertype) => OrderTypeModel.fromJson(ordertype))
            .toList();
        // print(ordertype.length);
        emit(OrderTypeLoadSuccess(ordertype: ordertype));
      } else {
        emit(const OrderTypeLoadFailed(message: 'OrderType Not Found'));
      }
    } catch (e) {
      emit(OrderTypeLoadFailed(message: e.toString()));
    }
  }

  void onOrderTypeDelete(
      OrderTypeDelete event, Emitter<OrderTypeState> emit) async {
    emit(OrderTypeDeleteInProgress());
    try {
      await _orderTypeRepository.deleteOrderType(event.guid);

      emit(OrderTypeDeleteSuccess());
    } catch (e) {
      emit(OrderTypeDeleteFailed(message: e.toString()));
    }
  }

  void onOrderTypeDeleteMany(
      OrderTypeDeleteMany event, Emitter<OrderTypeState> emit) async {
    emit(OrderTypeDeleteManyInProgress());
    try {
      await _orderTypeRepository.deleteOrderTypeMany(event.guid);

      emit(OrderTypeDeleteManySuccess());
    } catch (e) {
      emit(OrderTypeDeleteManyFailed(message: e.toString()));
    }
  }

  void onOrderTypeSave(
      OrderTypeSave event, Emitter<OrderTypeState> emit) async {
    emit(OrderTypeSaveInProgress());
    try {
      await _orderTypeRepository.saveOrderType(event.ordertypemodel);
      emit(OrderTypeSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(OrderTypeSaveFailed(message: error['message']));
    }
  }

  void onOrderTypeUpdate(
      OrderTypeUpdate event, Emitter<OrderTypeState> emit) async {
    emit(OrderTypeUpdateInProgress());
    try {
      await _orderTypeRepository.updateOrderType(
          event.guid, event.ordertypemodel);
      emit(OrderTypeUpdateSuccess());
    } catch (e) {
      emit(OrderTypeUpdateFailed(message: e.toString()));
    }
  }

  void onOrderTypeGet(OrderTypeGet event, Emitter<OrderTypeState> emit) async {
    emit(OrderTypeGetInProgress());
    try {
      final result = await _orderTypeRepository.getOrderType(event.guid);
      if (result.success) {
        OrderTypeModel orderTyle = OrderTypeModel.fromJson(result.data);
        emit(OrderTypeGetSuccess(ordertype: orderTyle));
      } else {
        emit(const OrderTypeGetFailed(message: 'OrderType Not Found'));
      }
    } catch (e) {
      emit(OrderTypeGetFailed(message: e.toString()));
    }
  }
}
