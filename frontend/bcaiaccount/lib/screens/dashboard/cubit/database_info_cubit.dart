import 'package:equatable/equatable.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/utils/util.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import '../../../global.dart' as global;

part 'database_info_state.dart';

class DatabaseInfoCubit extends Cubit<DatabaseInfoState> {
  DatabaseInfoCubit() : super(DatabaseInfoInitial());

  /// ดึงรายชื่อ table ทั้งหมดจาก PostgreSQL
  Future<void> loadDocCountByTransFlag() async {
    try {
      emit(DatabaseInfoLoading());

      // Query เพื่อดึงจำนวนเอกสารแยกตาม trans_flag พร้อมข้อมูลเพิ่มเติม
      final query = """
        WITH latest_docs AS (
          SELECT DISTINCT ON (transflag)
            transflag,
            docno,
            docdatetime,
            totalamount
          FROM doc
          ORDER BY transflag, docdatetime DESC, id DESC
        ),
        doc_summary AS (
          SELECT 
            transflag,
            COUNT(*) as xcount
          FROM doc
          GROUP BY transflag
        )
        SELECT 
          ds.transflag,
          ds.xcount,
          ld.docdatetime as last_doc_date,
          ld.docno as last_doc_no,
          ld.totalamount as total_amount
        FROM doc_summary ds
        LEFT JOIN latest_docs ld ON ds.transflag = ld.transflag
        ORDER BY ds.transflag
      """;

      final response = await pgSqlSelectGroup([query]);
      final rows = response["results"]?[0]?["rows"] as List?;

      if (rows == null) {
        emit( DatabaseInfoError(global.language('cannot_fetch_data')));
        return;
      }

      // แปลงผลลัพธ์เป็น List<Map<String, dynamic>>
      final data = rows.map((row) {
        return {
          'trans_flag': int.tryParse(row['transflag']?.toString() ?? '0') ?? 0,
          'xcount': int.tryParse(row['xcount']?.toString() ?? '0') ?? 0,
          'last_doc_date': row['last_doc_date']?.toString(),
          'last_doc_no': row['last_doc_no']?.toString(),
          'total_amount':
              double.tryParse(row['total_amount']?.toString() ?? '0') ?? 0.0,
        };
      }).toList();
      if (kDebugMode) {
        AppLogger.debug('Found ${data.length} records');
        for (var item in data) {
          AppLogger.debug(
            'transflag: ${item['trans_flag']}, count: ${item['xcount']}, last_date: ${item['last_doc_date']}, last_no: ${item['last_doc_no']}, total: ${item['total_amount']}',
          );
        }
      }

      emit(DatabaseInfoLoaded(data: data, totalRecords: data.length));
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error loading table names: $e');
      }
      emit(DatabaseInfoError('เกิดข้อผิดพลาด: $e'));
    }
  }

  /// รีเฟรชข้อมูล
  Future<void> refresh() async {
    await loadDocCountByTransFlag();
  }
}
