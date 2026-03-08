import 'package:smlaicloud/model/global_model.dart';
import '../../global.dart' as global;

/// Model สำหรับการเลือกสาขาหลายๆ สาขา
/// ใช้สำหรับส่งข้อมูลกลับจาก BranchSearchScreen
class BranchSelectionModel {
  final List<SearchGuidCodeNameModel> selectedBranches;
  final bool isCancel;

  const BranchSelectionModel({
    required this.selectedBranches,
    required this.isCancel,
  });

  /// แปลงรายการสาขาที่เลือกเป็น String สำหรับ API
  /// รูปแบบ: "code1,code2,code3"
  String getBranchCodeString() {
    if (selectedBranches.isEmpty) return '';
    return selectedBranches.map((branch) => branch.code).join(',');
  }

  /// แปลงรายการสาขาที่เลือกเป็น String สำหรับแสดงผล
  /// รูปแบบ: "ชื่อสาขา1, ชื่อสาขา2, ชื่อสาขา3"
  String getBranchDisplayString() {
    if (selectedBranches.isEmpty) return global.language('kb_all_branches');
    if (selectedBranches.length == 1) {
      return selectedBranches.first.getDisplayName();
    }
    return '${selectedBranches.length} สาขาที่เลือก';
  }

  /// สร้าง BranchSelectionModel ที่ยกเลิก
  factory BranchSelectionModel.cancelled() {
    return const BranchSelectionModel(
      selectedBranches: [],
      isCancel: true,
    );
  }

  /// สร้าง BranchSelectionModel จากรายการสาขา
  factory BranchSelectionModel.fromBranches(
      List<SearchGuidCodeNameModel> branches) {
    return BranchSelectionModel(
      selectedBranches: branches,
      isCancel: false,
    );
  }
}

/// Extension สำหรับ SearchGuidCodeNameModel
extension SearchGuidCodeNameModelExtension on SearchGuidCodeNameModel {
  String getDisplayName() {
    if (names.isNotEmpty) {
      // หาชื่อภาษาไทยก่อน
      final thName = names.where((name) => name.code == 'th').firstOrNull;
      if (thName != null && thName.name.isNotEmpty) {
        return thName.name;
      }
      // ถ้าไม่มีใช้ชื่อแรก
      if (names.first.name.isNotEmpty) {
        return names.first.name;
      }
    }
    return code; // ถ้าไม่มีชื่อใช้ code
  }
}
