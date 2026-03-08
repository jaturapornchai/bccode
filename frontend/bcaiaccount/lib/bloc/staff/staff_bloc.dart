import 'dart:convert';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/staff_model.dart';
import 'package:smlaicloud/repositories/staff_repository.dart';

part 'staff_event.dart';
part 'staff_state.dart';

class StaffBloc extends Bloc<StaffEvent, StaffState> {
  final StaffRepository _staffRepository;

  StaffBloc({required StaffRepository staffRepository})
      : _staffRepository = staffRepository,
        super(StaffInitial()) {
    on<StaffLoadList>(onStaffLoad);
    on<StaffSave>(onStaffSave);
    on<StaffUpdate>(onStaffUpdate);
    on<StaffDelete>(onStaffDelete);
    on<StaffDeleteMany>(onStaffDeleteMany);
    on<StaffGet>(onStaffGet);
  }

  void onStaffLoad(StaffLoadList event, Emitter<StaffState> emit) async {
    emit(StaffInProgress());

    try {
      final results = await _staffRepository.getStaffList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<StaffModel> staffs = (results.data as List)
            .map((staffs) => StaffModel.fromJson(staffs))
            .toList();
        emit(StaffLoadSuccess(staffs: staffs));
      } else {
        emit(const StaffLoadFailed(message: 'Staff Group Not Found'));
      }
    } catch (e) {
      emit(StaffLoadFailed(message: e.toString()));
    }
  }

  void onStaffDelete(StaffDelete event, Emitter<StaffState> emit) async {
    emit(StaffDeleteInProgress());
    try {
      await _staffRepository.deleteStaff(event.guid);

      emit(StaffDeleteSuccess());
    } catch (e) {
      emit(StaffDeleteFailed());
    }
  }

  void onStaffDeleteMany(
      StaffDeleteMany event, Emitter<StaffState> emit) async {
    emit(StaffDeleteManyInProgress());
    try {
      await _staffRepository.deleteStaffMany(event.guid);

      emit(StaffDeleteManySuccess());
    } catch (e) {
      // emit(StaffDeleteFailure(message: e.toString()));
    }
  }

  void onStaffSave(StaffSave event, Emitter<StaffState> emit) async {
    emit(StaffSaveInProgress());
    try {
      await _staffRepository.saveStaff(event.staffModel);
      emit(StaffSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(StaffSaveFailed(message: error['message']));
    }
  }

  void onStaffUpdate(StaffUpdate event, Emitter<StaffState> emit) async {
    emit(StaffUpdateInProgress());
    try {
      await _staffRepository.updateStaff(event.guid, event.staffModel);
      emit(StaffUpdateSuccess());
    } catch (e) {
      emit(StaffUpdateFailed(message: e.toString()));
    }
  }

  void onStaffGet(StaffGet event, Emitter<StaffState> emit) async {
    emit(StaffGetInProgress());
    try {
      final result = await _staffRepository.getStaff(event.guid);
      if (result.success) {
        StaffModel staff = StaffModel.fromJson(result.data);
        emit(StaffGetSuccess(staff: staff));
      } else {
        emit(const StaffGetFailed(message: 'Staff Not Found'));
      }
    } catch (e) {
      // emit(StaffDeleteFailure(message: e.toString()));
    }
  }
}
