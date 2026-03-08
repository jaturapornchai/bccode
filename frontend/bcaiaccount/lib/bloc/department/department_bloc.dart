import 'dart:convert';

import 'package:smlaicloud/model/department_model.dart';
import 'package:smlaicloud/repositories/department_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'department_event.dart';
part 'department_state.dart';

class DepartmentBloc extends Bloc<DepartmentEvent, DepartmentState> {
  final DepartmentRepository _departmentRepository;

  DepartmentBloc({required DepartmentRepository departmentRepository})
      : _departmentRepository = departmentRepository,
        super(DepartmentInitial()) {
    on<DepartmentLoadList>(onDepartmentLoad);
    on<DepartmentSave>(onDepartmentSave);
    on<DepartmentUpdate>(onDepartmentUpdate);
    on<DepartmentDelete>(onDepartmentDelete);
    on<DepartmentDeleteMany>(onDepartmentDeleteMany);
    on<DepartmentGet>(onDepartmentGet);
  }

  void onDepartmentLoad(
      DepartmentLoadList event, Emitter<DepartmentState> emit) async {
    emit(DepartmentInProgress());

    try {
      final results = await _departmentRepository.getDepartmentList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<DepartmentModel> department = (results.data as List)
            .map((department) => DepartmentModel.fromJson(department))
            .toList();
        emit(DepartmentLoadSuccess(department: department));
      } else {
        emit(const DepartmentLoadFailed(message: 'Department Not Found'));
      }
    } catch (e) {
      emit(DepartmentLoadFailed(message: e.toString()));
    }
  }

  void onDepartmentDelete(
      DepartmentDelete event, Emitter<DepartmentState> emit) async {
    emit(DepartmentDeleteInProgress());
    try {
      await _departmentRepository.deleteDepartment(event.guid);

      emit(DepartmentDeleteSuccess());
    } catch (e) {
      emit(DepartmentDeleteFailed());
    }
  }

  void onDepartmentDeleteMany(
      DepartmentDeleteMany event, Emitter<DepartmentState> emit) async {
    emit(DepartmentDeleteManyInProgress());
    try {
      await _departmentRepository.deleteDepartmentMany(event.guid);

      emit(DepartmentDeleteManySuccess());
    } catch (e) {
      emit(DepartmentDeleteFailed());
    }
  }

  void onDepartmentSave(
      DepartmentSave event, Emitter<DepartmentState> emit) async {
    emit(DepartmentSaveInProgress());
    try {
      final result =
          await _departmentRepository.saveDepartment(event.department);
      if (result.success) {
        emit(DepartmentSaveSuccess(responsesID: result.id));
      }
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(DepartmentSaveFailed(message: error['message']));
    }
  }

  void onDepartmentUpdate(
      DepartmentUpdate event, Emitter<DepartmentState> emit) async {
    emit(DepartmentUpdateInProgress());
    try {
      await _departmentRepository.updateDepartment(
          event.guid, event.department);
      emit(DepartmentUpdateSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(DepartmentUpdateFailed(message: error['message']));
    }
  }

  void onDepartmentGet(
      DepartmentGet event, Emitter<DepartmentState> emit) async {
    emit(DepartmentGetInProgress());
    try {
      final result = await _departmentRepository.getDepartment(event.guid);
      if (result.success) {
        DepartmentModel department = DepartmentModel.fromJson(result.data);
        emit(DepartmentGetSuccess(department: department));
      } else {
        emit(const DepartmentGetFailed(message: 'Department Not Found'));
      }
    } catch (e) {
      emit(DepartmentGetFailed(message: e.toString()));
    }
  }
}
