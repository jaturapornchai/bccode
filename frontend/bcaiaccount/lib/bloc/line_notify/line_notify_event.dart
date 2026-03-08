part of 'line_notify_bloc.dart';

abstract class LineNotifyEvent extends Equatable {
  const LineNotifyEvent();

  @override
  List<Object> get props => [];
}

class LineNotifyGet extends LineNotifyEvent {
  final String guid;

  const LineNotifyGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class LineNotifyLoadList extends LineNotifyEvent {
  final int limit;
  final int offset;
  final String search;

  const LineNotifyLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class LineNotifyDelete extends LineNotifyEvent {
  final String guid;

  const LineNotifyDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class LineNotifyDeleteMany extends LineNotifyEvent {
  final List<String> guid;

  const LineNotifyDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class LineNotifySave extends LineNotifyEvent {
  final LineNotifyModel lineNotify;

  const LineNotifySave({
    required this.lineNotify,
  });

  @override
  List<Object> get props => [lineNotify];
}

class LineNotifyUpdate extends LineNotifyEvent {
  final String guid;
  final LineNotifyModel lineNotify;

  const LineNotifyUpdate({
    required this.guid,
    required this.lineNotify,
  });

  @override
  List<Object> get props => [lineNotify];
}

class LineNotifyTest extends LineNotifyEvent {
  final String message;
  final String token;

  const LineNotifyTest({
    required this.message,
    required this.token,
  });

  @override
  List<Object> get props => [message, token];
}
