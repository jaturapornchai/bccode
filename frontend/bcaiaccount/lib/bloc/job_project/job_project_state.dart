part of 'job_project_bloc.dart';

abstract class JobProjectState extends Equatable {
  const JobProjectState();

  @override
  List<Object> get props => [];
}

class JobProjectInitial extends JobProjectState {}

class JobProjectInProgress extends JobProjectState {}

class JobProjectLoadSuccess extends JobProjectState {
  final List<JobProjectModel> jobProject;

  const JobProjectLoadSuccess({required this.jobProject});

  JobProjectLoadSuccess copyWith({
    List<JobProjectModel>? jobProject,
  }) =>
      JobProjectLoadSuccess(jobProject: jobProject ?? this.jobProject);

  @override
  List<Object> get props => [jobProject];
}

class JobProjectLoadFailed extends JobProjectState {
  final String message;

  const JobProjectLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class JobProjectSaveInitial extends JobProjectState {}

class JobProjectSaveInProgress extends JobProjectState {}

class JobProjectSaveSuccess extends JobProjectState {
  final String responsesID;

  const JobProjectSaveSuccess({
    required this.responsesID,
  });

  @override
  List<Object> get props => [responsesID];
}

class JobProjectSaveFailed extends JobProjectState {
  final String message;

  const JobProjectSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class JobProjectDeleteInProgress extends JobProjectState {}

class JobProjectDeleteSuccess extends JobProjectState {}

class JobProjectDeleteFailed extends JobProjectState {}

class JobProjectDeleteManyInProgress extends JobProjectState {}

class JobProjectDeleteManySuccess extends JobProjectState {}

class JobProjectDeleteManyFailed extends JobProjectState {}

class JobProjectGetInProgress extends JobProjectState {}

class JobProjectGetSuccess extends JobProjectState {
  final JobProjectModel jobProject;

  const JobProjectGetSuccess({required this.jobProject});

  JobProjectGetSuccess copyWith({
    JobProjectModel? jobProject,
  }) =>
      JobProjectGetSuccess(jobProject: jobProject ?? this.jobProject);

  @override
  List<Object> get props => [jobProject];
}

class JobProjectGetFailed extends JobProjectState {
  final String message;

  const JobProjectGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class JobProjectUpdateInitial extends JobProjectState {}

class JobProjectUpdateInProgress extends JobProjectState {}

class JobProjectUpdateSuccess extends JobProjectState {}

class JobProjectUpdateFailed extends JobProjectState {
  final String message;

  const JobProjectUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
