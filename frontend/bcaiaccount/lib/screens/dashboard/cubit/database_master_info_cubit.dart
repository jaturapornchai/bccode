import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter/foundation.dart';
import 'package:smlaicloud/utils/util.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

part 'database_master_info_state.dart';

class DatabaseMasterInfoCubit extends Cubit<DatabaseMasterInfoState> {
  DatabaseMasterInfoCubit() : super(DatabaseMasterInfoInitial());

  /// ดึงจำนวนข้อมูลหลักแยกตามประเภท
  Future<void> loadMasterDataCount() async {
    try {
      emit(DatabaseMasterInfoLoading());

      // Query เพื่อดึงจำนวนข้อมูลหลักต่างๆ
      final queries = [
        "SELECT 'product' as table_name, COUNT(*) as xcount FROM product",
        "SELECT 'productbarcode' as table_name, COUNT(*) as xcount FROM productbarcode",
        "SELECT 'ic_warehouse' as table_name, COUNT(*) as xcount FROM ic_warehouse",
        "SELECT 'ic_shelf' as table_name, COUNT(*) as xcount FROM ic_shelf",
        "SELECT 'customer' as table_name, COUNT(*) as xcount FROM customer",
        "SELECT 'creditor' as table_name, COUNT(*) as xcount FROM creditor",
        "SELECT 'debtor' as table_name, COUNT(*) as xcount FROM debtor",
        "SELECT 'doc' as table_name, COUNT(*) as xcount FROM doc",
        "SELECT 'docdetail' as table_name, COUNT(*) as xcount FROM docdetail",
      ];

      final response = await pgSqlSelectGroup(queries);

      List<Map<String, dynamic>> data = [];

      // ประมวลผลข้อมูลจากแต่ละ query
      if (response["results"] != null) {
        for (var i = 0; i < queries.length; i++) {
          final rows = response["results"]?[i]?["rows"] as List?;
          if (rows != null && rows.isNotEmpty) {
            final row = rows.first;
            data.add({
              'table_name': row['table_name']?.toString() ?? '',
              'xcount': int.tryParse(row['xcount']?.toString() ?? '0') ?? 0,
            });
          }
        }
      }

      emit(DatabaseMasterInfoLoaded(data: data, totalRecords: data.length));
    } catch (e, stackTrace) {
      if (kDebugMode) {
        AppLogger.debug(
          'Error loading master data: $e'
          ' database_master_info_cubit.dart',
          stackTrace: stackTrace,
        );
      }
      emit(DatabaseMasterInfoError('เกิดข้อผิดพลาด: $e'));
    }
  }

  /// รีเฟรชข้อมูล
  Future<void> refresh() async {
    await loadMasterDataCount();
  }
}
