part of 'holiday_bloc.dart';

abstract class HolidayEvent extends Equatable {
  const HolidayEvent();

  @override
  List<Object> get props => [];
}

class HolidayGet extends HolidayEvent {
  final String guid;

  const HolidayGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class HolidayLoadList extends HolidayEvent {
  final int limit;
  final int offset;
  final String search;

  const HolidayLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class HolidayDelete extends HolidayEvent {
  final String guid;

  const HolidayDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class HolidayDeleteMany extends HolidayEvent {
  final List<String> guid;

  const HolidayDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class HolidaySave extends HolidayEvent {
  final HolidayModel holidayModel;

  const HolidaySave({
    required this.holidayModel,
  });

  @override
  List<Object> get props => [holidayModel];
}

class HolidayUpdate extends HolidayEvent {
  final String guid;
  final HolidayModel holidayModel;

  const HolidayUpdate({
    required this.guid,
    required this.holidayModel,
  });

  @override
  List<Object> get props => [holidayModel];
}
