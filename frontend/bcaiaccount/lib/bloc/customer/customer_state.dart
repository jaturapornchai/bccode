part of 'customer_bloc.dart';

abstract class CustomerState extends Equatable {
  const CustomerState();

  @override
  List<Object> get props => [];
}

class CustomerInitial extends CustomerState {}

class CustomerInProgress extends CustomerState {}

class CustomerLoadSuccess extends CustomerState {
  final List<CustomerModel> customers;

  const CustomerLoadSuccess({required this.customers});

  CustomerLoadSuccess copyWith({
    List<CustomerModel>? customers,
  }) =>
      CustomerLoadSuccess(customers: customers ?? this.customers);

  @override
  List<Object> get props => [customers];
}

class CustomerLoadFailed extends CustomerState {
  final String message;

  const CustomerLoadFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CustomerSaveInitial extends CustomerState {}

class CustomerSaveInProgress extends CustomerState {}

class CustomerSaveSuccess extends CustomerState {}

class CustomerSaveFailed extends CustomerState {
  final String message;

  const CustomerSaveFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CustomerDeleteInProgress extends CustomerState {}

class CustomerDeleteSuccess extends CustomerState {}

class CustomerDeleteFailed extends CustomerState {}

class CustomerDeleteManyInProgress extends CustomerState {}

class CustomerDeleteManySuccess extends CustomerState {}

class CustomerDeleteManyFailed extends CustomerState {}

class CustomerGetInProgress extends CustomerState {}

class CustomerGetSuccess extends CustomerState {
  final CustomerModel customer;

  const CustomerGetSuccess({required this.customer});

  CustomerGetSuccess copyWith({
    CustomerModel? customer,
  }) =>
      CustomerGetSuccess(customer: customer ?? this.customer);

  @override
  List<Object> get props => [customer];
}

class CustomerGetFailed extends CustomerState {
  final String message;

  const CustomerGetFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}

class CustomerUpdateInitial extends CustomerState {}

class CustomerUpdateInProgress extends CustomerState {}

class CustomerUpdateSuccess extends CustomerState {}

class CustomerUpdateFailed extends CustomerState {
  final String message;

  const CustomerUpdateFailed({
    required this.message,
  });

  @override
  List<Object> get props => [message];
}
