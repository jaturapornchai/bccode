part of 'kitchen_bloc.dart';

abstract class KitchenState extends Equatable {
  const KitchenState();

  @override
  List<Object> get props => [];
}

class KitchenInitial extends KitchenState {}

class KitchenInProgress extends KitchenState {}

class KitchenLoadSuccess extends KitchenState {
  final List<KitchenModel> kitchens;

  const KitchenLoadSuccess({required this.kitchens});

  KitchenLoadSuccess copyWith({
    List<KitchenModel>? kitchens,
  }) =>
      KitchenLoadSuccess(kitchens: kitchens ?? this.kitchens);

  @override
  List<Object> get props => [kitchens];
}

class KitchenLoadFailed extends KitchenState {
  final String message;

  const KitchenLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class KitchenProductsInProgress extends KitchenState {}

class KitchenProductsLoadSuccess extends KitchenState {
  final List<ProductInKitchenModel> kitchens;

  const KitchenProductsLoadSuccess({required this.kitchens});

  KitchenProductsLoadSuccess copyWith({
    List<ProductInKitchenModel>? kitchens,
  }) =>
      KitchenProductsLoadSuccess(kitchens: kitchens ?? this.kitchens);

  @override
  List<Object> get props => [kitchens];
}

class KitchenProductsLoadFailed extends KitchenState {
  final String message;

  const KitchenProductsLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class KitchenSaveInitial extends KitchenState {}

class KitchenSaveInProgress extends KitchenState {}

class KitchenSaveSuccess extends KitchenState {}

class KitchenSaveFailed extends KitchenState {
  final String message;

  const KitchenSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class KitchenDeleteInProgress extends KitchenState {}

class KitchenDeleteSuccess extends KitchenState {}

class KitchenDeleteFailed extends KitchenState {
  final String message;

  const KitchenDeleteFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class KitchenDeleteManyInProgress extends KitchenState {}

class KitchenDeleteManySuccess extends KitchenState {}

class KitchenDeleteManyFailed extends KitchenState {}

class KitchenGetInProgress extends KitchenState {}

class KitchenGetSuccess extends KitchenState {
  final KitchenModel kitchen;

  const KitchenGetSuccess({required this.kitchen});

  KitchenGetSuccess copyWith({
    KitchenModel? kitchen,
  }) =>
      KitchenGetSuccess(kitchen: kitchen ?? this.kitchen);

  @override
  List<Object> get props => [kitchen];
}

class KitchenGetFailed extends KitchenState {
  final String message;

  const KitchenGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class KitchenUpdateInitial extends KitchenState {}

class KitchenUpdateInProgress extends KitchenState {}

class KitchenUpdateSuccess extends KitchenState {}

class KitchenUpdateFailed extends KitchenState {
  final String message;

  const KitchenUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class KitchenDetailBarcodeGetInProgress extends KitchenState {}

class KitchenDetailBarcodeGetSuccess extends KitchenState {
  final GetKitchenModel kitchen;

  const KitchenDetailBarcodeGetSuccess({required this.kitchen});

  KitchenDetailBarcodeGetSuccess copyWith({
    GetKitchenModel? kitchen,
  }) =>
      KitchenDetailBarcodeGetSuccess(kitchen: kitchen ?? this.kitchen);

  @override
  List<Object> get props => [kitchen];
}

class KitchenDetailBarcodeGetFailed extends KitchenState {
  final String message;

  const KitchenDetailBarcodeGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
