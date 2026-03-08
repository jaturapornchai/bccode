part of 'staff_bloc.dart';

abstract class StaffEvent extends Equatable {
  const StaffEvent();

  @override
  List<Object> get props => [];
}

class StaffGet extends StaffEvent {
  final String guid;

  const StaffGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class StaffLoadList extends StaffEvent {
  final int limit;
  final int offset;
  final String search;

  const StaffLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class StaffDelete extends StaffEvent {
  final String guid;

  const StaffDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class StaffDeleteMany extends StaffEvent {
  final List<String> guid;

  const StaffDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class StaffSave extends StaffEvent {
  final StaffModel staffModel;

  const StaffSave({
    required this.staffModel,
  });

  @override
  List<Object> get props => [staffModel];
}

class StaffUpdate extends StaffEvent {
  final String guid;
  final StaffModel staffModel;

  const StaffUpdate({
    required this.guid,
    required this.staffModel,
  });

  @override
  List<Object> get props => [staffModel];
}
