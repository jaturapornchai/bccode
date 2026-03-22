import 'package:equatable/equatable.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/utils/util.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import '../../../global.dart' as global;

part 'procurement_dashboard_state.dart';

/// Cubit สำหรับดึงข้อมูลสรุปจัดซื้อ (PR/RFQ/PO) จาก ClickHouse
/// รองรับ filter ตามวันที่, ประเภทเอกสาร, สถานะ
class ProcurementDashboardCubit extends Cubit<ProcurementDashboardState> {
  ProcurementDashboardCubit() : super(ProcurementDashboardInitial());

  /// โหลดข้อมูลสรุปจัดซื้อทั้งหมดจาก ClickHouse พร้อม filter
  Future<void> loadDashboard({
    DateTime? dateFrom,
    DateTime? dateTo,
    List<int> docTypes = const [21, 22, 6], // 21=PR, 22=RFQ, 6=PO
    List<String> statuses = const [],
  }) async {
    try {
      emit(ProcurementDashboardLoading());

      final db = global.clickHouseDatabaseName;

      // กำหนด default date range — ถ้าไม่ระบุใช้ 6 เดือนย้อนหลัง
      final now = DateTime.now();
      final effectiveDateTo = dateTo ?? now;
      final effectiveDateFrom =
          dateFrom ?? DateTime(now.year, now.month - 6, now.day);

      // คำนวณ previous period (ช่วงก่อนหน้า ระยะเท่ากัน)
      final duration = effectiveDateTo.difference(effectiveDateFrom);
      final prevDateTo =
          effectiveDateFrom.subtract(const Duration(days: 1));
      final prevDateFrom = prevDateTo.subtract(duration);

      // สร้าง WHERE clause สำหรับ current period
      final docWhere = _buildWhereClause(db, 'doc',
          dateFrom: effectiveDateFrom,
          dateTo: effectiveDateTo,
          docTypes: docTypes,
          statuses: statuses);

      // WHERE clause สำหรับ previous period (ไม่ filter status — เทียบภาพรวม)
      final prevDocWhere = _buildWhereClause(db, 'doc',
          dateFrom: prevDateFrom,
          dateTo: prevDateTo,
          docTypes: docTypes);

      // WHERE clause สำหรับ pending count (ไม่ filter status — แสดง pending ทั้งหมดเสมอ)
      final pendingWhere = _buildWhereClause(db, 'doc',
          dateFrom: effectiveDateFrom,
          dateTo: effectiveDateTo,
          docTypes: docTypes);

      // WHERE clause สำหรับ docdetail (top products)
      final detailWhere = _buildWhereClause(db, 'docdetail',
          dateFrom: effectiveDateFrom,
          dateTo: effectiveDateTo,
          docTypes: docTypes,
          statuses: statuses);

      // Query 1: KPI — จำนวนเอกสาร + ยอดรวม แยกตาม transflag
      final kpiQuery =
          "SELECT transflag, count() as doc_count, sum(totalamount) as total_amount, avg(totalamount) as avg_amount "
          "FROM $docWhere "
          "GROUP BY transflag ORDER BY transflag";

      // Query 2: Previous Period KPI — สำหรับคำนวณ trend % change
      final prevKpiQuery =
          "SELECT transflag, count() as doc_count, sum(totalamount) as total_amount, avg(totalamount) as avg_amount "
          "FROM $prevDocWhere "
          "GROUP BY transflag ORDER BY transflag";

      // Query 3: Monthly Trend — สรุปรายเดือน
      final monthlyQuery =
          "SELECT formatDateTime(docdatetime, '%Y-%m') as month, transflag, count() as doc_count, sum(totalamount) as total_amount "
          "FROM $docWhere "
          "GROUP BY month, transflag ORDER BY month, transflag";

      // Query 4: Status Summary — สถานะเอกสาร
      final statusQuery =
          "SELECT transflag, approval_status, count() as doc_count "
          "FROM $docWhere "
          "GROUP BY transflag, approval_status ORDER BY transflag";

      // Query 5: Top Vendors — เจ้าหนี้ที่ซื้อบ่อย/มากสุด
      final vendorQuery =
          "SELECT custcode, count() as doc_count, sum(totalamount) as total_amount "
          "FROM $docWhere AND custcode != '' "
          "GROUP BY custcode ORDER BY total_amount DESC LIMIT 10";

      // Query 6: Pending Count — จำนวน pending แยกตาม transflag (ไม่ filter status)
      final pendingQuery =
          "SELECT transflag, count() as doc_count "
          "FROM $pendingWhere AND approval_status = 'pending' "
          "GROUP BY transflag ORDER BY transflag";

      // Query 7: Top Products — สินค้าที่สั่งซื้อมากสุด จาก docdetail
      final productQuery =
          "SELECT itemcode, anyLast(itemname) as itemname, sum(qty) as total_qty, sum(sumamount) as total_amount, avg(price) as avg_price "
          "FROM $detailWhere "
          "GROUP BY itemcode ORDER BY total_amount DESC LIMIT 20";

      final response = await clickhouseSelectGroup([
        kpiQuery,
        prevKpiQuery,
        monthlyQuery,
        statusQuery,
        vendorQuery,
        pendingQuery,
        productQuery,
      ]);

      // แปลงผลลัพธ์ — รองรับทั้ง "rows" และ "data" key
      final kpiRows = _extractRows(response, 0);
      final prevKpiRows = _extractRows(response, 1);
      final monthlyRows = _extractRows(response, 2);
      final statusRows = _extractRows(response, 3);
      final vendorRows = _extractRows(response, 4);
      final pendingRows = _extractRows(response, 5);
      final productRows = _extractRows(response, 6);

      // สร้าง KPI data
      final kpiData = <String, Map<String, dynamic>>{};
      for (var row in kpiRows) {
        final flag = int.tryParse(row['transflag']?.toString() ?? '0') ?? 0;
        final key = _transFlagToKey(flag);
        kpiData[key] = {
          'doc_count':
              int.tryParse(row['doc_count']?.toString() ?? '0') ?? 0,
          'total_amount':
              double.tryParse(row['total_amount']?.toString() ?? '0') ?? 0.0,
          'avg_amount':
              double.tryParse(row['avg_amount']?.toString() ?? '0') ?? 0.0,
        };
      }

      // สร้าง Previous Period KPI data
      final prevKpiData = <String, Map<String, dynamic>>{};
      for (var row in prevKpiRows) {
        final flag = int.tryParse(row['transflag']?.toString() ?? '0') ?? 0;
        final key = _transFlagToKey(flag);
        prevKpiData[key] = {
          'doc_count':
              int.tryParse(row['doc_count']?.toString() ?? '0') ?? 0,
          'total_amount':
              double.tryParse(row['total_amount']?.toString() ?? '0') ?? 0.0,
          'avg_amount':
              double.tryParse(row['avg_amount']?.toString() ?? '0') ?? 0.0,
        };
      }

      // สร้าง Monthly trend data
      final monthlyData = <Map<String, dynamic>>[];
      for (var row in monthlyRows) {
        monthlyData.add({
          'month': row['month']?.toString() ?? '',
          'transflag':
              int.tryParse(row['transflag']?.toString() ?? '0') ?? 0,
          'doc_count':
              int.tryParse(row['doc_count']?.toString() ?? '0') ?? 0,
          'total_amount':
              double.tryParse(row['total_amount']?.toString() ?? '0') ?? 0.0,
        });
      }

      // สร้าง Status data
      final statusData = <Map<String, dynamic>>[];
      for (var row in statusRows) {
        statusData.add({
          'transflag':
              int.tryParse(row['transflag']?.toString() ?? '0') ?? 0,
          'approval_status': row['approval_status']?.toString() ?? '',
          'doc_count':
              int.tryParse(row['doc_count']?.toString() ?? '0') ?? 0,
        });
      }

      // สร้าง Vendor data
      final vendorData = <Map<String, dynamic>>[];
      for (var row in vendorRows) {
        vendorData.add({
          'custcode': row['custcode']?.toString() ?? '',
          'doc_count':
              int.tryParse(row['doc_count']?.toString() ?? '0') ?? 0,
          'total_amount':
              double.tryParse(row['total_amount']?.toString() ?? '0') ?? 0.0,
        });
      }

      // สร้าง Pending count data
      final pendingCount = <String, int>{};
      for (var row in pendingRows) {
        final flag = int.tryParse(row['transflag']?.toString() ?? '0') ?? 0;
        final key = _transFlagToKey(flag);
        pendingCount[key] =
            int.tryParse(row['doc_count']?.toString() ?? '0') ?? 0;
      }

      // สร้าง Product data
      final productData = <Map<String, dynamic>>[];
      for (var row in productRows) {
        productData.add({
          'itemcode': row['itemcode']?.toString() ?? '',
          'itemname': row['itemname']?.toString() ?? '',
          'total_qty':
              double.tryParse(row['total_qty']?.toString() ?? '0') ?? 0.0,
          'total_amount':
              double.tryParse(row['total_amount']?.toString() ?? '0') ?? 0.0,
          'avg_price':
              double.tryParse(row['avg_price']?.toString() ?? '0') ?? 0.0,
        });
      }

      if (kDebugMode) {
        AppLogger.debug(
            '[ProcurementDashboard] KPI: ${kpiData.length}, PrevKPI: ${prevKpiData.length}, '
            'Monthly: ${monthlyData.length}, Status: ${statusData.length}, '
            'Vendors: ${vendorData.length}, Pending: ${pendingCount.length}, '
            'Products: ${productData.length}');
      }

      emit(ProcurementDashboardLoaded(
        kpiData: kpiData,
        prevKpiData: prevKpiData,
        monthlyData: monthlyData,
        statusData: statusData,
        vendorData: vendorData,
        pendingCount: pendingCount,
        productData: productData,
      ));
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('[ProcurementDashboard] Error: $e');
      }
      emit(ProcurementDashboardError(
          '${global.language('error_occurred')}: $e'));
    }
  }

  /// สร้าง WHERE clause จาก filter parameters
  /// คืนค่า "{db}.{table} WHERE ..." สำหรับต่อท้าย FROM
  String _buildWhereClause(
    String db,
    String table, {
    DateTime? dateFrom,
    DateTime? dateTo,
    List<int> docTypes = const [21, 22, 6],
    List<String> statuses = const [],
  }) {
    final parts = [
      "shopid = '${global.getShopId()}'",
      'isdelete = 0',
    ];
    if (docTypes.isNotEmpty) {
      parts.add('transflag IN (${docTypes.join(',')})');
    }
    if (dateFrom != null) {
      parts.add("docdatetime >= '${_formatDate(dateFrom)}'");
    }
    if (dateTo != null) {
      parts.add("docdatetime <= '${_formatDate(dateTo)} 23:59:59'");
    }
    if (statuses.isNotEmpty) {
      final statusStr = statuses.map((s) => "'$s'").join(',');
      parts.add('approval_status IN ($statusStr)');
    }
    return '$db.$table WHERE ${parts.join(' AND ')}';
  }

  /// Format DateTime เป็น YYYY-MM-DD สำหรับ ClickHouse
  String _formatDate(DateTime d) =>
      '${d.year}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';

  /// ดึง rows จาก response — รองรับทั้ง "rows" และ "data" key
  List _extractRows(Map<String, dynamic> response, int index) {
    final result = response["results"]?[index];
    if (result == null) return [];
    return result["rows"] as List? ?? result["data"] as List? ?? [];
  }

  /// แปลง transflag เป็น key string
  String _transFlagToKey(int flag) {
    switch (flag) {
      case 21:
        return 'pr';
      case 22:
        return 'rfq';
      case 6:
        return 'po';
      default:
        return 'unknown';
    }
  }

  /// รีเฟรชข้อมูลด้วย filter เดิม
  Future<void> refresh({
    DateTime? dateFrom,
    DateTime? dateTo,
    List<int> docTypes = const [21, 22, 6],
    List<String> statuses = const [],
  }) async {
    await loadDashboard(
      dateFrom: dateFrom,
      dateTo: dateTo,
      docTypes: docTypes,
      statuses: statuses,
    );
  }
}
