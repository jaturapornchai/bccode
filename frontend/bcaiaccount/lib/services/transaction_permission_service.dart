import 'package:dio/dio.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Service สำหรับตรวจสอบสิทธิ์การแก้ไขเอกสาร Transaction
class TransactionPermissionService {
  static final Dio _dio = Dio();

  /// ตรวจสอบว่าผู้ใช้สามารถแก้ไขเอกสาร PO ได้หรือไม่
  ///
  /// เงื่อนไข:
  /// 1. เอกสารอนุมัติแล้ว (approved/autoApproved)
  /// 2. ยังไม่ถูกอ้างอิงโดยเอกสารอื่น (not referenced)
  /// 3. ผู้ใช้เป็นผู้สร้างเอกสาร (creator)
  static Future<CanEditResult> canEditPurchaseOrder({
    required TransactionModel doc,
    POApprovalStatusModel? approvalStatus,
    String? currentUserCode,
  }) async {
    try {
      // ใช้ usercode จาก global ถ้าไม่ได้ระบุ
      final userCode = currentUserCode ?? global.appConfig.getString("user") ?? '';

      // ถ้ามี approvalStatus ให้ใช้ status จาก approval
      if (approvalStatus != null) {
        // 1. ตรวจสอบสถานะ draft/rejected - แก้ไขได้เสมอ
        if (approvalStatus.canSubmit) {
          return CanEditResult(
            canEdit: true,
            reason: null,
          );
        }

        // 2. ตรวจสอบสถานะ pending - รออนุมัติ
        if (approvalStatus.status == POApprovalStatus.pending) {
          return CanEditResult(
            canEdit: false,
            reason: global.language('doc_pending_approval_cannot_edit'),
          );
        }

        // 3. ตรวจสอบสถานะ approved/autoApproved - ต้องเช็คเงื่อนไขพิเศษ
        if (approvalStatus.isCompleted) {
          // 3.1 ตรวจสอบว่าเป็นผู้สร้างหรือไม่
          // IMPORTANT: ต้องเช็คว่า createdBy ตรงกับ current user
          final createdBy = approvalStatus.createdBy;
          if (createdBy.isEmpty) {
            // ถ้าไม่มี creator - ไม่อนุญาตให้แก้ไข (safe mode)
            return CanEditResult(
              canEdit: false,
              reason: global.language('cannot_identify_doc_creator'),
            );
          }

          if (createdBy != userCode) {
            return CanEditResult(
              canEdit: false,
              reason:
                  'เอกสารนี้สร้างโดย ${approvalStatus.createdByName} คุณไม่สามารถแก้ไขได้',
            );
          }

          // 3.2 ตรวจสอบว่าถูกอ้างอิงหรือไม่
          final hasRef = await checkIfReferenced(
            docNo: doc.docno,
            shopId: global.getShopId(),
          );

          if (hasRef.isReferenced) {
            return CanEditResult(
              canEdit: false,
              reason:
                  'เอกสารถูกอ้างอิงโดย ${hasRef.referencedBy?.join(", ") ?? global.language("other_document")} แล้ว',
            );
          }

          // ✅ ผ่านทุกเงื่อนไข - อนุญาตให้แก้ไข
          return CanEditResult(
            canEdit: true,
            reason: null,
          );
        }
      }

      // ถ้าไม่มี approvalStatus - ใช้ logic แบบเก่า (ตรวจจาก iscancel)
      if (doc.iscancel) {
        return CanEditResult(
          canEdit: false,
          reason: 'เอกสารถูกยกเลิกแล้ว ไม่สามารถแก้ไขได้',
        );
      }

      // ตรวจสอบว่าเป็นผู้สร้างหรือไม่
      // IMPORTANT: ต้องเช็คว่า creatorcode ตรงกับ current user
      final creatorCode = doc.creatorcode ?? '';
      if (creatorCode.isEmpty) {
        // ถ้าไม่มี creator code - ไม่อนุญาตให้แก้ไข (safe mode)
        return CanEditResult(
          canEdit: false,
          reason: global.language('cannot_identify_doc_creator'),
        );
      }

      if (creatorCode != userCode) {
        return CanEditResult(
          canEdit: false,
          reason: 'เอกสารนี้สร้างโดย ${doc.creatorname ?? doc.creatorcode} คุณไม่สามารถแก้ไขได้',
        );
      }

      // ตรวจสอบว่าถูกอ้างอิงหรือไม่
      final hasRef = await checkIfReferenced(
        docNo: doc.docno,
        shopId: global.getShopId(),
      );

      if (hasRef.isReferenced) {
        return CanEditResult(
          canEdit: false,
          reason: 'เอกสารถูกอ้างอิงโดย ${hasRef.referencedBy?.join(", ") ?? global.language("other_document")} แล้ว',
        );
      }

      // ✅ อนุญาตให้แก้ไข
      return CanEditResult(
        canEdit: true,
        reason: null,
      );
    } catch (e) {
      AppLogger.error('[TransactionPermissionService] Error checking canEdit: $e');
      return CanEditResult(
        canEdit: false,
        reason: 'เกิดข้อผิดพลาดในการตรวจสอบสิทธิ์',
      );
    }
  }

  /// ตรวจสอบว่าเอกสารถูกอ้างอิงโดยเอกสารอื่นหรือไม่
  static Future<ReferenceCheckResult> checkIfReferenced({
    required String docNo,
    required String shopId,
  }) async {
    try {
      final url = global.goApiUrlPath('api/transaction/check-reference');
      final response = await _dio.get(
        url,
        queryParameters: {
          'shopid': shopId,
          'docno': docNo,
        },
      );

      if (response.statusCode == 200 && response.data['success'] == true) {
        final data = response.data['data'];
        return ReferenceCheckResult(
          isReferenced: data['is_referenced'] ?? false,
          referencedBy: data['referenced_by'] != null
              ? List<String>.from(data['referenced_by'])
              : null,
        );
      }

      // ถ้า API ไม่มี - สมมติว่าไม่ถูกอ้างอิง (safe mode)
      return ReferenceCheckResult(
        isReferenced: false,
        referencedBy: null,
      );
    } catch (e) {
      AppLogger.error('[TransactionPermissionService] Error checking references: $e');
      // ถ้า error - สมมติว่าถูกอ้างอิงแล้ว (safe mode - ห้ามแก้ไข)
      return ReferenceCheckResult(
        isReferenced: true,
        referencedBy: null,
      );
    }
  }
}

/// ผลลัพธ์การตรวจสอบสิทธิ์แก้ไข
class CanEditResult {
  final bool canEdit;
  final String? reason; // เหตุผลที่แก้ไขไม่ได้

  CanEditResult({
    required this.canEdit,
    this.reason,
  });
}

/// ผลลัพธ์การตรวจสอบการอ้างอิง
class ReferenceCheckResult {
  final bool isReferenced;
  final List<String>? referencedBy; // รายการเอกสารที่อ้างอิง เช่น ["GR001", "IV002"]

  ReferenceCheckResult({
    required this.isReferenced,
    this.referencedBy,
  });
}
