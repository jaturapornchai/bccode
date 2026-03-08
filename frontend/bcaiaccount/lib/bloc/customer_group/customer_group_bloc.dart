import 'dart:convert';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/customer_group_model.dart';
import 'package:smlaicloud/repositories/customer_group_repository.dart';

part 'customer_group_event.dart';
part 'customer_group_state.dart';

class CustomerGroupBloc extends Bloc<CustomerGroupEvent, CustomerGroupState> {
  final CustomerGroupRepository _customerGroupRepository;

  CustomerGroupBloc({required CustomerGroupRepository customerGroupRepository})
      : _customerGroupRepository = customerGroupRepository,
        super(CustomerGroupInitial()) {
    on<CustomerGroupLoadList>(onCustomerGroupLoad);
    on<CustomerGroupSave>(onCustomerGroupSave);
    on<CustomerGroupUpdate>(onCustomerGroupUpdate);
    on<CustomerGroupDelete>(customerGroupDelete);
    on<CustomerGroupDeleteMany>(customerGroupDeleteMany);
    on<CustomerGroupGet>(onCustomerGroupGet);
  }

  void onCustomerGroupLoad(
      CustomerGroupLoadList event, Emitter<CustomerGroupState> emit) async {
    emit(CustomerGroupInProgress());

    try {
      final results = await _customerGroupRepository.getCustomerGroupList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<CustomerGroupModel> customerGroups = (results.data as List)
            .map((customerGroup) => CustomerGroupModel.fromJson(customerGroup))
            .toList();
        emit(CustomerGroupLoadSuccess(customerGroups: customerGroups));
      } else {
        emit(
            const CustomerGroupLoadFailed(message: 'Customer Group Not Found'));
      }
    } catch (e) {
      emit(CustomerGroupLoadFailed(message: e.toString()));
    }
  }

  void customerGroupDelete(
      CustomerGroupDelete event, Emitter<CustomerGroupState> emit) async {
    emit(CustomerGroupDeleteInProgress());
    try {
      await _customerGroupRepository.deleteCustomerGroup(event.guid);

      emit(CustomerGroupDeleteSuccess());
    } catch (e) {
      emit(CustomerGroupDeleteFailed());
    }
  }

  void customerGroupDeleteMany(
      CustomerGroupDeleteMany event, Emitter<CustomerGroupState> emit) async {
    emit(CustomerGroupDeleteManyInProgress());
    try {
      await _customerGroupRepository.deleteCustomerGroupMany(event.guid);

      emit(CustomerGroupDeleteManySuccess());
    } catch (e) {
      emit(CustomerGroupDeleteFailed());
    }
  }

  void onCustomerGroupSave(
      CustomerGroupSave event, Emitter<CustomerGroupState> emit) async {
    emit(CustomerGroupSaveInProgress());
    try {
      await _customerGroupRepository.saveCustomerGroup(event.customerGroup);
      emit(CustomerGroupSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(CustomerGroupSaveFailed(message: error['message']));
    }
  }

  void onCustomerGroupUpdate(
      CustomerGroupUpdate event, Emitter<CustomerGroupState> emit) async {
    emit(CustomerGroupUpdateInProgress());
    try {
      await _customerGroupRepository.updateCustomerGroup(
          event.guid, event.customerGroup);
      emit(CustomerGroupUpdateSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(CustomerGroupUpdateFailed(message: error['message']));
    }
  }

  void onCustomerGroupGet(
      CustomerGroupGet event, Emitter<CustomerGroupState> emit) async {
    emit(CustomerGroupGetInProgress());
    try {
      final result =
          await _customerGroupRepository.getCustomerGroup(event.guid);
      if (result.success) {
        CustomerGroupModel customerGroup =
            CustomerGroupModel.fromJson(result.data);
        emit(CustomerGroupGetSuccess(customerGroup: customerGroup));
      } else {
        emit(const CustomerGroupGetFailed(message: 'CustomerGroup Not Found'));
      }
    } catch (e) {
      // emit(CustomerGroupDeleteFailure(message: e.toString()));
    }
  }
}
