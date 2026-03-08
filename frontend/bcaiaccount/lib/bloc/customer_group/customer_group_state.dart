part of 'customer_group_bloc.dart';

abstract class CustomerGroupState extends Equatable {
  const CustomerGroupState();

  @override
  List<Object> get props => [];
}

class CustomerGroupInitial extends CustomerGroupState {}

class CustomerGroupInProgress extends CustomerGroupState {}

class CustomerGroupLoadSuccess extends CustomerGroupState {
  final List<CustomerGroupModel> customerGroups;

  const CustomerGroupLoadSuccess({required this.customerGroups});

  CustomerGroupLoadSuccess copyWith({
    List<CustomerGroupModel>? customerGroups,
  }) =>
      CustomerGroupLoadSuccess(
          customerGroups: customerGroups ?? this.customerGroups);

  @override
  List<Object> get props => [customerGroups];
}

class CustomerGroupLoadFailed extends CustomerGroupState {
  final String message;

  const CustomerGroupLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CustomerGroupSaveInitial extends CustomerGroupState {}

class CustomerGroupSaveInProgress extends CustomerGroupState {}

class CustomerGroupSaveSuccess extends CustomerGroupState {}

class CustomerGroupSaveFailed extends CustomerGroupState {
  final String message;

  const CustomerGroupSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CustomerGroupDeleteInProgress extends CustomerGroupState {}

class CustomerGroupDeleteSuccess extends CustomerGroupState {}

class CustomerGroupDeleteFailed extends CustomerGroupState {}

class CustomerGroupDeleteManyInProgress extends CustomerGroupState {}

class CustomerGroupDeleteManySuccess extends CustomerGroupState {}

class CustomerGroupDeleteManyFailed extends CustomerGroupState {}

class CustomerGroupGetInProgress extends CustomerGroupState {}

class CustomerGroupGetSuccess extends CustomerGroupState {
  final CustomerGroupModel customerGroup;

  const CustomerGroupGetSuccess({required this.customerGroup});

  CustomerGroupGetSuccess copyWith({
    CustomerGroupModel? customerGroup,
  }) =>
      CustomerGroupGetSuccess(
          customerGroup: customerGroup ?? this.customerGroup);

  @override
  List<Object> get props => [customerGroup];
}

class CustomerGroupGetFailed extends CustomerGroupState {
  final String message;

  const CustomerGroupGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CustomerGroupUpdateInitial extends CustomerGroupState {}

class CustomerGroupUpdateInProgress extends CustomerGroupState {}

class CustomerGroupUpdateSuccess extends CustomerGroupState {}

class CustomerGroupUpdateFailed extends CustomerGroupState {
  final String message;

  const CustomerGroupUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
