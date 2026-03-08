import 'package:smlaicloud/repositories/table_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/model/table_model.dart';

part 'table_event.dart';
part 'table_state.dart';

class TableBloc extends Bloc<TableEvent, TableState> {
  final TableRepository _tableRepository;

  TableBloc({required TableRepository tableRepository})
      : _tableRepository = tableRepository,
        super(TableInitial()) {
    on<TableLoadList>(onTableLoad);
    on<TableSave>(onTableSave);
    on<TableUpdate>(onTableUpdate);
    on<TableDelete>(onTableDelete);
    on<TableDeleteMany>(onTableDeleteMany);
    on<TableGet>(onTableGet);
    on<TableUpdateXorder>(onTableUpdateXorder);
  }

  void onTableLoad(TableLoadList event, Emitter<TableState> emit) async {
    emit(TableInProgress());

    try {
      final results = await _tableRepository.getTableList(
          offset: event.offset,
          limit: event.limit,
          search: event.search,
          groupNumber: event.groupNumber);

      if (results.success) {
        List<TableModel> tables = (results.data as List)
            .map((tables) => TableModel.fromJson(tables))
            .toList();
        emit(TableLoadSuccess(tables: tables));
      } else {
        emit(const TableLoadFailed(message: 'Table Group Not Found'));
      }
    } catch (e) {
      emit(TableLoadFailed(message: e.toString()));
    }
  }

  void onTableDelete(TableDelete event, Emitter<TableState> emit) async {
    emit(TableDeleteInProgress());
    try {
      await _tableRepository.deleteTable(event.guid);

      emit(TableDeleteSuccess());
    } catch (e) {
      // emit(TableDeleteFailure(message: e.toString()));
    }
  }

  void onTableDeleteMany(
      TableDeleteMany event, Emitter<TableState> emit) async {
    emit(TableDeleteManyInProgress());
    try {
      await _tableRepository.deleteTableMany(event.guid);

      emit(TableDeleteManySuccess());
    } catch (e) {
      // emit(TableDeleteFailure(message: e.toString()));
    }
  }

  void onTableSave(TableSave event, Emitter<TableState> emit) async {
    emit(TableSaveInProgress());
    try {
      await _tableRepository.saveTable(event.tableModel);
      emit(TableSaveSuccess());
    } catch (e) {
      emit(TableSaveFailed(message: e.toString()));
    }
  }

  void onTableUpdate(TableUpdate event, Emitter<TableState> emit) async {
    emit(TableUpdateInProgress());
    try {
      await _tableRepository.updateTable(event.guid, event.tableModel);
      emit(TableUpdateSuccess());
    } catch (e) {
      emit(TableUpdateFailed(message: e.toString()));
    }
  }

  void onTableGet(TableGet event, Emitter<TableState> emit) async {
    emit(TableGetInProgress());
    try {
      final result = await _tableRepository.getTable(event.guid);
      if (result.success) {
        TableModel Table = TableModel.fromJson(result.data);
        emit(TableGetSuccess(table: Table));
      } else {
        emit(const TableGetFailed(message: 'Table Not Found'));
      }
    } catch (e) {
      // emit(TableDeleteFailure(message: e.toString()));
    }
  }

  void onTableUpdateXorder(
      TableUpdateXorder event, Emitter<TableState> emit) async {
    emit(TableUpdateInProgress());
    try {
      await _tableRepository.updateTableXorder(event.tableModel);
      emit(TableUpdateSuccess());
    } catch (e) {
      emit(TableUpdateFailed(message: e.toString()));
    }
  }
}
