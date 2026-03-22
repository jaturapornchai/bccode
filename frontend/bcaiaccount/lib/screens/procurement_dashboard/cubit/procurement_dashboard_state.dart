part of 'procurement_dashboard_cubit.dart';

/// สถานะของ Procurement Dashboard
abstract class ProcurementDashboardState extends Equatable {
  const ProcurementDashboardState();

  @override
  List<Object?> get props => [];
}

/// สถานะเริ่มต้น
class ProcurementDashboardInitial extends ProcurementDashboardState {}

/// กำลังโหลดข้อมูล
class ProcurementDashboardLoading extends ProcurementDashboardState {}

/// โหลดข้อมูลสำเร็จ
class ProcurementDashboardLoaded extends ProcurementDashboardState {
  /// KPI แยกตาม key (pr/rfq/po) -> {doc_count, total_amount, avg_amount}
  final Map<String, Map<String, dynamic>> kpiData;

  /// KPI ช่วงก่อนหน้า — สำหรับคำนวณ trend % change
  final Map<String, Map<String, dynamic>> prevKpiData;

  /// ข้อมูลรายเดือน (monthly trend)
  final List<Map<String, dynamic>> monthlyData;

  /// สรุปสถานะเอกสาร (draft/pending/approved/rejected)
  final List<Map<String, dynamic>> statusData;

  /// Top 10 เจ้าหนี้ตามยอดสั่งซื้อ
  final List<Map<String, dynamic>> vendorData;

  /// จำนวนเอกสาร pending แยกตาม key {pr: 3, rfq: 0, po: 2}
  final Map<String, int> pendingCount;

  /// Top 20 สินค้าที่สั่งซื้อมากสุด (itemcode, itemname, total_qty, total_amount, avg_price)
  final List<Map<String, dynamic>> productData;

  const ProcurementDashboardLoaded({
    required this.kpiData,
    required this.prevKpiData,
    required this.monthlyData,
    required this.statusData,
    required this.vendorData,
    required this.pendingCount,
    required this.productData,
  });

  @override
  List<Object?> get props => [
        kpiData,
        prevKpiData,
        monthlyData,
        statusData,
        vendorData,
        pendingCount,
        productData,
      ];
}

/// เกิดข้อผิดพลาด
class ProcurementDashboardError extends ProcurementDashboardState {
  final String message;

  const ProcurementDashboardError(this.message);

  @override
  List<Object?> get props => [message];
}
