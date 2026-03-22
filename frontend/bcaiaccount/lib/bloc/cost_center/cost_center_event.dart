part of 'cost_center_bloc.dart';

abstract class CostCenterEvent extends Equatable {
  const CostCenterEvent();

  @override
  List<Object> get props => [];
}

class CostCenterGet extends CostCenterEvent {
  final String guid;

  const CostCenterGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class CostCenterLoadList extends CostCenterEvent {
  final int limit;
  final int offset;
  final String search;

  const CostCenterLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [];
}

class CostCenterDelete extends CostCenterEvent {
  final String guid;

  const CostCenterDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class CostCenterDeleteMany extends CostCenterEvent {
  final List<String> guid;

  const CostCenterDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class CostCenterSave extends CostCenterEvent {
  final CostCenterModel costCenter;

  const CostCenterSave({
    required this.costCenter,
  });

  @override
  List<Object> get props => [costCenter];
}

class CostCenterUpdate extends CostCenterEvent {
  final String guid;
  final CostCenterModel costCenter;

  const CostCenterUpdate({
    required this.guid,
    required this.costCenter,
  });

  @override
  List<Object> get props => [costCenter];
}
