part of 'job_project_bloc.dart';

abstract class JobProjectEvent extends Equatable {
  const JobProjectEvent();

  @override
  List<Object> get props => [];
}

class JobProjectGet extends JobProjectEvent {
  final String guid;

  const JobProjectGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class JobProjectLoadList extends JobProjectEvent {
  final int limit;
  final int offset;
  final String search;

  const JobProjectLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class JobProjectDelete extends JobProjectEvent {
  final String guid;

  const JobProjectDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class JobProjectDeleteMany extends JobProjectEvent {
  final List<String> guid;

  const JobProjectDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class JobProjectSave extends JobProjectEvent {
  final JobProjectModel jobProject;

  const JobProjectSave({
    required this.jobProject,
  });

  @override
  List<Object> get props => [jobProject];
}

class JobProjectUpdate extends JobProjectEvent {
  final String guid;
  final JobProjectModel jobProject;

  const JobProjectUpdate({
    required this.guid,
    required this.jobProject,
  });

  @override
  List<Object> get props => [jobProject];
}
