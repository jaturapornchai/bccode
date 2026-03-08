import 'dart:convert';

import 'package:smlaicloud/model/business_type_model.dart';
import 'package:smlaicloud/repositories/business_type_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'business_type_event.dart';
part 'business_type_state.dart';

class BusinessTypeBloc extends Bloc<BusinessTypeEvent, BusinessTypeState> {
  final BusinessTypeRepository _businessTypeRepository;

  BusinessTypeBloc({required BusinessTypeRepository businessTypeRepository})
      : _businessTypeRepository = businessTypeRepository,
        super(BusinessTypeInitial()) {
    on<BusinessTypeLoadList>(onBusinessTypeLoad);
    on<BusinessTypeSave>(onBusinessTypeSave);
    on<BusinessTypeUpdate>(onBusinessTypeUpdate);
    on<BusinessTypeDelete>(onBusinessTypeDelete);
    on<BusinessTypeDeleteMany>(onBusinessTypeDeleteMany);
    on<BusinessTypeGet>(onBusinessTypeGet);
  }

  void onBusinessTypeLoad(
      BusinessTypeLoadList event, Emitter<BusinessTypeState> emit) async {
    emit(BusinessTypeInProgress());

    try {
      final results = await _businessTypeRepository.getBusinessTypeList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<BusinessTypeModel> businessType = (results.data as List)
            .map((businessType) => BusinessTypeModel.fromJson(businessType))
            .toList();
        emit(BusinessTypeLoadSuccess(businessType: businessType));
      } else {
        emit(const BusinessTypeLoadFailed(message: 'Customer Group Not Found'));
      }
    } catch (e) {
      emit(BusinessTypeLoadFailed(message: e.toString()));
    }
  }

  void onBusinessTypeDelete(
      BusinessTypeDelete event, Emitter<BusinessTypeState> emit) async {
    emit(BusinessTypeDeleteInProgress());
    try {
      await _businessTypeRepository.deleteBusinessType(event.guid);

      emit(BusinessTypeDeleteSuccess());
    } catch (e) {
      emit(BusinessTypeDeleteFailed());
    }
  }

  void onBusinessTypeDeleteMany(
      BusinessTypeDeleteMany event, Emitter<BusinessTypeState> emit) async {
    emit(BusinessTypeDeleteManyInProgress());
    try {
      await _businessTypeRepository.deleteBusinessTypeMany(event.guid);

      emit(BusinessTypeDeleteManySuccess());
    } catch (e) {
      emit(BusinessTypeDeleteFailed());
    }
  }

  void onBusinessTypeSave(
      BusinessTypeSave event, Emitter<BusinessTypeState> emit) async {
    emit(BusinessTypeSaveInProgress());
    try {
      await _businessTypeRepository.saveBusinessType(event.businessType);
      emit(BusinessTypeSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(BusinessTypeSaveFailed(message: error['message']));
    }
  }

  void onBusinessTypeUpdate(
      BusinessTypeUpdate event, Emitter<BusinessTypeState> emit) async {
    emit(BusinessTypeUpdateInProgress());
    try {
      await _businessTypeRepository.updateBusinessType(
          event.guid, event.businessType);
      emit(BusinessTypeUpdateSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(BusinessTypeUpdateFailed(message: error['message']));
    }
  }

  void onBusinessTypeGet(
      BusinessTypeGet event, Emitter<BusinessTypeState> emit) async {
    emit(BusinessTypeGetInProgress());
    try {
      final result = await _businessTypeRepository.getBusinessType(event.guid);
      if (result.success) {
        BusinessTypeModel businessType =
            BusinessTypeModel.fromJson(result.data);
        emit(BusinessTypeGetSuccess(businessType: businessType));
      } else {
        emit(const BusinessTypeGetFailed(message: 'BusinessType Not Found'));
      }
    } catch (e) {
      emit(BusinessTypeGetFailed(message: e.toString()));
    }
  }
}
