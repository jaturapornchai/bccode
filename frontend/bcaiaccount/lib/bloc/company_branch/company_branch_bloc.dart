import 'dart:convert';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/repositories/company_branch_repository.dart';
import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

part 'company_branch_event.dart';
part 'company_branch_state.dart';

class CompanyBranchBloc extends Bloc<CompanyBranchEvent, CompanyBranchState> {
  final CompanyBranchRepository _companyBranchRepository;

  CompanyBranchBloc({required CompanyBranchRepository companyBranchRepository})
      : _companyBranchRepository = companyBranchRepository,
        super(CompanyBranchInitial()) {
    on<CompanyBranchLoadList>(onCompanyBranchLoad);
    on<CompanyBranchSave>(onCompanyBranchSave);
    on<CompanyBranchUpdate>(onCompanyBranchUpdate);
    on<CompanyBranchDelete>(onCompanyBranchDelete);
    on<CompanyBranchDeleteMany>(onCompanyBranchDeleteMany);
    on<CompanyBranchGet>(onCompanyBranchGet);
    on<CompanyBranchGetBycode>(onCompanyBranchGetBycode);
    on<CompanyBranchByBusinessTypeLoadList>(onCompanyBranchByBuisnessTypeLoad);
  }

  void onCompanyBranchLoad(
      CompanyBranchLoadList event, Emitter<CompanyBranchState> emit) async {
    emit(CompanyBranchInProgress());

    try {
      final results = await _companyBranchRepository.getBranchList(
        offset: event.offset,
        limit: event.limit,
        search: event.search,
      );

      if (results.success) {
        List<CompanyBranchModel> companyBranch = (results.data as List)
            .map((companyBranch) => CompanyBranchModel.fromJson(companyBranch))
            .toList();
        emit(CompanyBranchLoadSuccess(companyBranch: companyBranch));
      } else {
        emit(const CompanyBranchLoadFailed(message: 'CompanyBranch Not Found'));
      }
    } catch (e) {
      emit(CompanyBranchLoadFailed(message: e.toString()));
    }
  }

  void onCompanyBranchByBuisnessTypeLoad(
      CompanyBranchByBusinessTypeLoadList event,
      Emitter<CompanyBranchState> emit) async {
    emit(CompanyBranchByBusinessTypeInProgress());

    try {
      final results = await _companyBranchRepository.getBranchList(
        offset: event.offset,
        limit: event.limit,
        search: event.search,
        businesstypecode: event.businesstypecode,
      );

      if (results.success) {
        List<CompanyBranchModel> companyBranch = (results.data as List)
            .map((companyBranch) => CompanyBranchModel.fromJson(companyBranch))
            .toList();
        emit(CompanyBranchByBusinessTypeLoadSuccess(
            companyBranch: companyBranch));
      } else {
        emit(const CompanyBranchByBusinessTypeLoadFailed(
            message: 'CompanyBranch Not Found'));
      }
    } catch (e) {
      emit(CompanyBranchByBusinessTypeLoadFailed(message: e.toString()));
    }
  }

  void onCompanyBranchDelete(
      CompanyBranchDelete event, Emitter<CompanyBranchState> emit) async {
    emit(CompanyBranchDeleteInProgress());
    try {
      await _companyBranchRepository.deleteBranch(event.guid);

      emit(CompanyBranchDeleteSuccess());
    } catch (e) {
      emit(CompanyBranchDeleteFailed());
    }
  }

  void onCompanyBranchDeleteMany(
      CompanyBranchDeleteMany event, Emitter<CompanyBranchState> emit) async {
    emit(CompanyBranchDeleteManyInProgress());
    try {
      await _companyBranchRepository.deleteBranchMany(event.guid);

      emit(CompanyBranchDeleteManySuccess());
    } catch (e) {
      emit(CompanyBranchDeleteFailed());
    }
  }

  void onCompanyBranchSave(
      CompanyBranchSave event, Emitter<CompanyBranchState> emit) async {
    emit(CompanyBranchSaveInProgress());
    try {
      await _companyBranchRepository.saveBranch(event.companyBranch);
      emit(CompanyBranchSaveSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(CompanyBranchSaveFailed(message: error['message']));
    }
  }

  void onCompanyBranchUpdate(
      CompanyBranchUpdate event, Emitter<CompanyBranchState> emit) async {
    emit(CompanyBranchUpdateInProgress());
    try {
      await _companyBranchRepository.updateBranch(
          event.guid, event.companyBranch);
      emit(CompanyBranchUpdateSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(CompanyBranchUpdateFailed(message: error['message']));
    }
  }

  void onCompanyBranchGet(
      CompanyBranchGet event, Emitter<CompanyBranchState> emit) async {
    emit(CompanyBranchGetInProgress());

    if (event.guid.isEmpty) {
      emit(const CompanyBranchGetFailed(message: 'Branch guidfixed Not Found'));
      return;
    }
    try {
      final result = await _companyBranchRepository.getBranch(event.guid);
      if (result.success) {
        CompanyBranchModel companyBranch =
            CompanyBranchModel.fromJson(result.data);
        emit(CompanyBranchGetSuccess(companyBranch: companyBranch));
      } else {
        emit(const CompanyBranchGetFailed(message: 'CompanyBranch Not Found'));
      }
    } catch (e) {
      emit(CompanyBranchGetFailed(message: e.toString()));
    }
  }

  void onCompanyBranchGetBycode(
      CompanyBranchGetBycode event, Emitter<CompanyBranchState> emit) async {
    emit(CompanyBranchGetBycodeInProgress());
    try {
      final result = await _companyBranchRepository.getBranchBycode(event.code);
      if (result.success) {
        CompanyBranchModel companyBranch =
            CompanyBranchModel.fromJson(result.data);
        emit(CompanyBranchGetBycodeSuccess(companyBranch: companyBranch));
      } else {
        emit(const CompanyBranchGetBycodeFailed(
            message: 'CompanyBranch Not Found'));
      }
    } catch (e) {
      emit(CompanyBranchGetBycodeFailed(message: e.toString()));
    }
  }
}
