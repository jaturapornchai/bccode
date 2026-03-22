import 'dart:convert';

import 'package:smlaicloud/model/job_project_model.dart';
import 'package:smlaicloud/repositories/job_project_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

part 'job_project_event.dart';
part 'job_project_state.dart';

class JobProjectBloc extends Bloc<JobProjectEvent, JobProjectState> {
  final JobProjectRepository _jobProjectRepository;

  JobProjectBloc({required JobProjectRepository jobProjectRepository})
      : _jobProjectRepository = jobProjectRepository,
        super(JobProjectInitial()) {
    on<JobProjectLoadList>(onJobProjectLoad);
    on<JobProjectSave>(onJobProjectSave);
    on<JobProjectUpdate>(onJobProjectUpdate);
    on<JobProjectDelete>(onJobProjectDelete);
    on<JobProjectDeleteMany>(onJobProjectDeleteMany);
    on<JobProjectGet>(onJobProjectGet);
  }

  void onJobProjectLoad(
      JobProjectLoadList event, Emitter<JobProjectState> emit) async {
    emit(JobProjectInProgress());

    try {
      final results = await _jobProjectRepository.getJobProjectList(
          offset: event.offset, limit: event.limit, search: event.search);

      if (results.success) {
        List<JobProjectModel> jobProject = (results.data as List)
            .map((jobProject) => JobProjectModel.fromJson(jobProject))
            .toList();
        emit(JobProjectLoadSuccess(jobProject: jobProject));
      } else {
        emit(const JobProjectLoadFailed(message: 'JobProject Not Found'));
      }
    } catch (e) {
      emit(JobProjectLoadFailed(message: e.toString()));
    }
  }

  void onJobProjectDelete(
      JobProjectDelete event, Emitter<JobProjectState> emit) async {
    emit(JobProjectDeleteInProgress());
    try {
      await _jobProjectRepository.deleteJobProject(event.guid);

      emit(JobProjectDeleteSuccess());
    } catch (e) {
      emit(JobProjectDeleteFailed());
    }
  }

  void onJobProjectDeleteMany(
      JobProjectDeleteMany event, Emitter<JobProjectState> emit) async {
    emit(JobProjectDeleteManyInProgress());
    try {
      await _jobProjectRepository.deleteJobProjectMany(event.guid);

      emit(JobProjectDeleteManySuccess());
    } catch (e) {
      emit(JobProjectDeleteFailed());
    }
  }

  void onJobProjectSave(
      JobProjectSave event, Emitter<JobProjectState> emit) async {
    emit(JobProjectSaveInProgress());
    try {
      final result =
          await _jobProjectRepository.saveJobProject(event.jobProject);
      if (result.success) {
        emit(JobProjectSaveSuccess(responsesID: result.id));
      }
    } catch (e) {
      try {
        final error = jsonDecode(e.toString());
        emit(JobProjectSaveFailed(message: error['message']));
      } catch (_) {
        emit(JobProjectSaveFailed(message: e.toString()));
      }
    }
  }

  void onJobProjectUpdate(
      JobProjectUpdate event, Emitter<JobProjectState> emit) async {
    emit(JobProjectUpdateInProgress());
    try {
      await _jobProjectRepository.updateJobProject(
          event.guid, event.jobProject);
      emit(JobProjectUpdateSuccess());
    } catch (e) {
      try {
        final error = jsonDecode(e.toString());
        emit(JobProjectUpdateFailed(message: error['message']));
      } catch (_) {
        emit(JobProjectUpdateFailed(message: e.toString()));
      }
    }
  }

  void onJobProjectGet(
      JobProjectGet event, Emitter<JobProjectState> emit) async {
    emit(JobProjectGetInProgress());
    try {
      final result = await _jobProjectRepository.getJobProject(event.guid);
      if (result.success) {
        JobProjectModel jobProject = JobProjectModel.fromJson(result.data);
        emit(JobProjectGetSuccess(jobProject: jobProject));
      } else {
        emit(const JobProjectGetFailed(message: 'JobProject Not Found'));
      }
    } catch (e) {
      emit(JobProjectGetFailed(message: e.toString()));
    }
  }
}
