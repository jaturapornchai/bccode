part of 'database_info_cubit.dart';

abstract class DatabaseInfoState extends Equatable {
  const DatabaseInfoState();

  @override
  List<Object?> get props => [];
}

class DatabaseInfoInitial extends DatabaseInfoState {}

class DatabaseInfoLoading extends DatabaseInfoState {}

class DatabaseInfoLoaded extends DatabaseInfoState {
  final List<Map<String, dynamic>> data;
  final int totalRecords;

  const DatabaseInfoLoaded({required this.data, required this.totalRecords});

  @override
  List<Object?> get props => [data, totalRecords];
}

class DatabaseInfoError extends DatabaseInfoState {
  final String message;

  const DatabaseInfoError(this.message);

  @override
  List<Object?> get props => [message];
}
