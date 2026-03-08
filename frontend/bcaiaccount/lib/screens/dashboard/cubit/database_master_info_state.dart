part of 'database_master_info_cubit.dart';

abstract class DatabaseMasterInfoState {
  const DatabaseMasterInfoState();
}

class DatabaseMasterInfoInitial extends DatabaseMasterInfoState {}

class DatabaseMasterInfoLoading extends DatabaseMasterInfoState {}

class DatabaseMasterInfoLoaded extends DatabaseMasterInfoState {
  final List<Map<String, dynamic>> data;
  final int totalRecords;

  const DatabaseMasterInfoLoaded({
    required this.data,
    required this.totalRecords,
  });
}

class DatabaseMasterInfoError extends DatabaseMasterInfoState {
  final String message;

  const DatabaseMasterInfoError(this.message);
}
