import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/model/employee_model.dart';
import 'package:smlaicloud/model/master_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/employee_repository.dart';

part 'employee_event.dart';
part 'employee_state.dart';

class EmployeeBloc extends Bloc<EmployeeEvent, EmployeeState> {
  final EmployeeRepository _employeeRepository;

  EmployeeBloc({required EmployeeRepository employeeRepository})
      : _employeeRepository = employeeRepository,
        super(EmployeeInitial()) {
    on<EmployeeLoadList>(onEmployeeLoad);
    on<EmployeeSave>(onEmployeeSave);
    on<EmployeeUpdate>(onEmployeeUpdate);
    on<EmployeeDelete>(employeeDelete);
    on<EmployeeDeleteMany>(employeeDeleteMany);
    on<EmployeeGet>(onEmployeeGet);
    on<EmployeeWithImageSave>(onEmployeeWithImageSave);
    on<EmployeeWithImageUpdate>(onEmployeeWithImageUpdate);
  }

  void onEmployeeLoad(
      EmployeeLoadList event, Emitter<EmployeeState> emit) async {
    emit(EmployeeInProgress());
    try {
      final results = await _employeeRepository.getEmployeeList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<EmployeeModel> employees = (results.data as List)
            .map((employee) => EmployeeModel.fromJson(employee))
            .toList();
        // print(employees.length);
        emit(EmployeeLoadSuccess(employees: employees));
      } else {
        emit(const EmployeeLoadFailed(message: 'Employee Not Found'));
      }
    } catch (e) {
      emit(EmployeeLoadFailed(message: e.toString()));
    }
  }

  void employeeDelete(EmployeeDelete event, Emitter<EmployeeState> emit) async {
    emit(EmployeeDeleteInProgress());
    try {
      await _employeeRepository.deleteEmployee(event.guid);

      emit(EmployeeDeleteSuccess());
    } catch (e) {
      // emit(EmployeeDeleteFailure(message: e.toString()));
    }
  }

  void employeeDeleteMany(
      EmployeeDeleteMany event, Emitter<EmployeeState> emit) async {
    emit(EmployeeDeleteManyInProgress());
    try {
      await _employeeRepository.deleteEmployeeMany(event.guid);

      emit(EmployeeDeleteManySuccess());
    } catch (e) {
      // emit(EmployeeDeleteFailure(message: e.toString()));
    }
  }

  void onEmployeeSave(EmployeeSave event, Emitter<EmployeeState> emit) async {
    emit(EmployeeSaveInProgress());
    try {
      // print(event.employeeModel);
      await _employeeRepository.saveEmployee(event.employeeModel);
      emit(EmployeeSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(EmployeeSaveFailed(message: error['message']));
    }
  }

  void onEmployeeUpdate(
      EmployeeUpdate event, Emitter<EmployeeState> emit) async {
    emit(EmployeeUpdateInProgress());
    try {
      await _employeeRepository.updateEmployee(event.guid, event.employeeModel);
      emit(EmployeeUpdateSuccess());
    } catch (e) {
      emit(EmployeeUpdateFailed(message: e.toString()));
    }
  }

  void onEmployeeGet(EmployeeGet event, Emitter<EmployeeState> emit) async {
    emit(EmployeeGetInProgress());
    try {
      final result = await _employeeRepository.getEmployee(event.guid);
      if (result.success) {
        EmployeeModel employee = EmployeeModel.fromJson(result.data);
        emit(EmployeeGetSuccess(employee: employee));
      } else {
        emit(const EmployeeGetFailed(message: 'Employee Not Found'));
      }
    } catch (e) {
      // emit(EmployeeDeleteFailure(message: e.toString()));
    }
  }

  void onEmployeeWithImageSave(
      EmployeeWithImageSave event, Emitter<EmployeeState> emit) async {
    emit(EmployeeSaveInProgress());
    try {
      ApiResponse result = await _employeeRepository.uploadImage(
          event.imageFile, event.imageWeb!);
      if (result.success) {
        UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
        EmployeeModel employeeModel = event.employee;
        employeeModel.profilepicture = uploadImage.uri;
        await _employeeRepository.saveEmployee(employeeModel);
        emit(EmployeeSaveSuccess());
      } else {
        emit(EmployeeSaveFailed(message: result.message));
      }
    } catch (e) {
      emit(EmployeeSaveFailed(message: e.toString()));
    }
  }

  void onEmployeeWithImageUpdate(
      EmployeeWithImageUpdate event, Emitter<EmployeeState> emit) async {
    emit(EmployeeUpdateInProgress());
    try {
      ApiResponse result = await _employeeRepository.uploadImage(
          event.imageFile, event.imageWeb);
      if (result.success) {
        UploadImageModel uploadImage = UploadImageModel.fromJson(result.data);
        EmployeeModel employeeModel = event.employee;
        employeeModel.profilepicture = uploadImage.uri;
        await _employeeRepository.updateEmployee(event.guid, employeeModel);
        emit(EmployeeUpdateSuccess());
      } else {
        emit(EmployeeUpdateFailed(message: result.message));
      }
    } catch (e) {
      emit(EmployeeUpdateFailed(message: e.toString()));
    }
  }
}
