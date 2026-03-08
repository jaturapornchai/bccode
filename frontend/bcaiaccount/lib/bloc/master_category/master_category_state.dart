part of 'master_category_bloc.dart';

abstract class MasterCategoryState extends Equatable {
  const MasterCategoryState();

  @override
  List<Object> get props => [];
}

class MasterCategoryInitial extends MasterCategoryState {}

class MasterCategoryInProgress extends MasterCategoryState {}

class MasterCategoryLoadSuccess extends MasterCategoryState {
  final List<MasterCategoryModel> categories;

  const MasterCategoryLoadSuccess({required this.categories});

  MasterCategoryLoadSuccess copyWith({
    List<MasterCategoryModel>? categories,
  }) =>
      MasterCategoryLoadSuccess(categories: categories ?? this.categories);

  @override
  List<Object> get props => [categories];
}

class MasterCategoryLoadFailed extends MasterCategoryState {
  final String message;

  const MasterCategoryLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterCategorySaveInitial extends MasterCategoryState {}

class MasterCategorySaveInProgress extends MasterCategoryState {}

class MasterCategorySaveSuccess extends MasterCategoryState {}

class MasterCategorySaveFailed extends MasterCategoryState {
  final String message;

  const MasterCategorySaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterCategoryDeleteInProgress extends MasterCategoryState {}

class MasterCategoryDeleteSuccess extends MasterCategoryState {}

class MasterCategoryDeleteFailed extends MasterCategoryState {}

class MasterCategoryDeleteManyInProgress extends MasterCategoryState {}

class MasterCategoryDeleteManySuccess extends MasterCategoryState {}

class MasterCategoryDeleteManyFailed extends MasterCategoryState {}

class MasterCategoryGetInProgress extends MasterCategoryState {}

class MasterCategoryGetSuccess extends MasterCategoryState {
  final MasterCategoryModel category;

  const MasterCategoryGetSuccess({required this.category});

  MasterCategoryGetSuccess copyWith({
    MasterCategoryModel? category,
  }) =>
      MasterCategoryGetSuccess(category: category ?? this.category);

  @override
  List<Object> get props => [category];
}

class MasterCategoryGetFailed extends MasterCategoryState {
  final String message;

  const MasterCategoryGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class MasterCategoryUpdateInitial extends MasterCategoryState {}

class MasterCategoryUpdateInProgress extends MasterCategoryState {}

class MasterCategoryUpdateSuccess extends MasterCategoryState {}

class MasterCategoryUpdateFailed extends MasterCategoryState {
  final String message;

  const MasterCategoryUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
