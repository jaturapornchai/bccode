part of 'business_type_bloc.dart';

abstract class BusinessTypeEvent extends Equatable {
  const BusinessTypeEvent();

  @override
  List<Object> get props => [];
}

class BusinessTypeGet extends BusinessTypeEvent {
  final String guid;

  const BusinessTypeGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class BusinessTypeLoadList extends BusinessTypeEvent {
  final int limit;
  final int offset;
  final String search;

  const BusinessTypeLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class BusinessTypeDelete extends BusinessTypeEvent {
  final String guid;

  const BusinessTypeDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class BusinessTypeDeleteMany extends BusinessTypeEvent {
  final List<String> guid;

  const BusinessTypeDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class BusinessTypeSave extends BusinessTypeEvent {
  final BusinessTypeModel businessType;

  const BusinessTypeSave({
    required this.businessType,
  });

  @override
  List<Object> get props => [businessType];
}

class BusinessTypeUpdate extends BusinessTypeEvent {
  final String guid;
  final BusinessTypeModel businessType;

  const BusinessTypeUpdate({
    required this.guid,
    required this.businessType,
  });

  @override
  List<Object> get props => [businessType];
}
