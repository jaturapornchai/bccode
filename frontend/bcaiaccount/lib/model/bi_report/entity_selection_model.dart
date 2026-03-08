import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/bi_report/branch_selection_model.dart'; // Import for extension
import '../../global.dart' as global;

/// Model สำหรับการเลือก Entity (Creditor หรือ Employee) หลายๆ รายการ
/// ใช้สำหรับส่งข้อมูลกลับจาก MultiEntitySearchScreen
class EntitySelectionModel {
  final List<SearchGuidCodeNameModel> selectedEntities;
  final bool isCancel;

  const EntitySelectionModel({
    required this.selectedEntities,
    required this.isCancel,
  });

  /// แปลงรายการ Entity ที่เลือกเป็น String สำหรับ API
  /// รูปแบบ: "code1,code2,code3"
  String getEntityCodeString() {
    if (selectedEntities.isEmpty) return '';
    return selectedEntities.map((entity) => entity.code).join(',');
  }

  /// แปลงรายการ Entity ที่เลือกเป็น String สำหรับแสดงผล
  /// รูปแบบ: "ชื่อ Entity1, ชื่อ Entity2, ชื่อ Entity3"
  String getEntityDisplayString() {
    if (selectedEntities.isEmpty) return global.language('all');
    if (selectedEntities.length == 1) {
      return selectedEntities.first.getDisplayName();
    }
    return '${selectedEntities.length} รายการที่เลือก';
  }

  /// แปลงรายการ Barcode ที่เลือกเป็น String สำหรับแสดงผล (สำหรับ barcode เลือกได้เพียงรายการเดียว)
  String getBarcodeDisplayString() {
    if (selectedEntities.isEmpty) return 'กดเพื่อเลือกบาร์โค้ด (ทั้งหมด)';
    return selectedEntities.first.getDisplayName();
  }

  /// สร้าง EntitySelectionModel ที่ยกเลิก
  factory EntitySelectionModel.cancelled() {
    return const EntitySelectionModel(
      selectedEntities: [],
      isCancel: true,
    );
  }

  /// สร้าง EntitySelectionModel จากรายการ Entity
  factory EntitySelectionModel.fromEntities(
      List<SearchGuidCodeNameModel> entities) {
    return EntitySelectionModel(
      selectedEntities: entities,
      isCancel: false,
    );
  }
}
