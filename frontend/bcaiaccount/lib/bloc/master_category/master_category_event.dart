part of 'master_category_bloc.dart';

abstract class MasterCategoryEvent extends Equatable {
  const MasterCategoryEvent();

  @override
  List<Object> get props => [];
}

class MasterCategoryGet extends MasterCategoryEvent {
  final String guid;

  const MasterCategoryGet({required this.guid});

  @override
  List<Object> get props => [guid];
}

class MasterCategoryLoadList extends MasterCategoryEvent {
  final int limit;
  final int offset;
  final String search;

  const MasterCategoryLoadList(
      {required this.offset, required this.limit, required this.search});

  @override
  List<Object> get props => [offset, limit, search];
}

class MasterCategoryGetByCode extends MasterCategoryEvent {
  final String code;

  const MasterCategoryGetByCode({required this.code});

  @override
  List<Object> get props => [code];
}

class MasterCategoryDelete extends MasterCategoryEvent {
  final String guid;

  const MasterCategoryDelete({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterCategoryDeleteMany extends MasterCategoryEvent {
  final List<String> guid;

  const MasterCategoryDeleteMany({
    required this.guid,
  });

  @override
  List<Object> get props => [guid];
}

class MasterCategorySave extends MasterCategoryEvent {
  final MasterCategoryModel categoryModel;

  const MasterCategorySave({
    required this.categoryModel,
  });

  @override
  List<Object> get props => [categoryModel];
}

class MasterCategoryUpdate extends MasterCategoryEvent {
  final String guid;
  final MasterCategoryModel categoryModel;

  const MasterCategoryUpdate({
    required this.guid,
    required this.categoryModel,
  });

  @override
  List<Object> get props => [guid, categoryModel];
}
