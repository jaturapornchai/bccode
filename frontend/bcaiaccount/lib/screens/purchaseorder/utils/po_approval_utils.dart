import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/services/approval_api_service.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// PO Approval Utilities
/// Helper functions สำหรับจัดการระบบอนุมัติ PO
class POApprovalUtils {
  /// สร้าง items list สำหรับส่งไป LIFF display
  static List<Map<String, dynamic>> buildItemsListForApproval(
      List<TransactionDetailModel>? details) {
    final List<Map<String, dynamic>> itemsList = [];
    if (details == null) return itemsList;

    for (int i = 0; i < details.length; i++) {
      final detail = details[i];
      itemsList.add({
        'linenumber': i + 1,
        'itemcode': detail.itemcode,
        'itemname': global.activeLangName(detail.itemnames ?? []),
        'qty': detail.qty,
        'unitname': global.activeLangName(detail.unitnames ?? []),
        'price': detail.price,
        'sumamount': detail.sumamount,
      });
    }
    return itemsList;
  }

  /// ส่ง PO เข้าระบบอนุมัติ
  /// คืนค่า POApprovalSubmitResult
  static Future<POApprovalSubmitResult> submitForApproval({
    required TransactionModel screenData,
    required String purchaseTypeCode,
    required String purchaseTypeName,
    bool isModified = false,
    POApprovalStatusModel? existingStatus,
  }) async {
    if (purchaseTypeCode.isEmpty) {
      return POApprovalSubmitResult.error(global.language('please_select_purchase_type'));
    }

    try {
      final totalAmount = screenData.totalaftervat;
      final userCode = global.appConfig.getString("user") ?? '';
      final userName = userCode;
      const approvalLevel = 0;

      // ถ้ามี approval status อยู่แล้ว ถือว่าเป็นการแก้ไขเอกสาร
      final effectiveIsModified = isModified || (existingStatus != null);

      AppLogger.debug(
          '[POApprovalUtils] Submit - isModified: $isModified, hasExistingStatus: ${existingStatus != null}, effectiveIsModified: $effectiveIsModified');

      // Build items list for LIFF display
      final itemsList = buildItemsListForApproval(screenData.details);

      return await ApprovalApiService.submitPOApproval(
        docNo: screenData.docno,
        guidFixed: screenData.guidfixed ?? '',
        purchaseTypeCode: purchaseTypeCode,
        purchaseTypeName: purchaseTypeName,
        totalAmount: totalAmount,
        actionBy: userCode,
        actionByName: userName,
        approvalLevel: approvalLevel,
        isModified: effectiveIsModified,
        comment: (screenData.description?.isNotEmpty ?? false)
            ? screenData.description
            : null,
        docDatetime: screenData.docdatetime,
        custCode: screenData.custcode,
        custName: global.activeLangName(screenData.custnames ?? []),
        items: itemsList,
      );
    } catch (e) {
      AppLogger.error('[POApprovalUtils] Submit error: $e');
      return POApprovalSubmitResult.error('ส่งเข้าระบบอนุมัติผิดพลาด: $e');
    }
  }

  /// โหลดสถานะการอนุมัติของ PO
  static Future<POApprovalStatusResult> loadApprovalStatus(String docNo) async {
    if (docNo.isEmpty) {
      return POApprovalStatusResult.notFound();
    }

    try {
      AppLogger.debug(
          '[POApprovalUtils] Loading approval status for docNo: $docNo');
      return await ApprovalApiService.getPOApprovalStatus(docNo);
    } catch (e) {
      AppLogger.error('[POApprovalUtils] Load status error: $e');
      return POApprovalStatusResult.error('โหลดสถานะไม่สำเร็จ: $e');
    }
  }

  /// ตรวจสอบว่าสามารถพิมพ์ PO ได้หรือไม่
  /// (ต้องผ่านการอนุมัติแล้ว)
  static bool canPrintPO(POApprovalStatusModel? status) {
    if (status == null) return false;
    return status.status == POApprovalStatus.approved ||
        status.status == POApprovalStatus.autoApproved;
  }

  /// ตรวจสอบว่าสามารถแก้ไข PO ได้หรือไม่
  static bool canEditPO(POApprovalStatusModel? status, bool isNewDocument) {
    // เอกสารใหม่ แก้ไขได้เสมอ
    if (isNewDocument) return true;

    // ถ้าไม่มี approval status ถือว่าแก้ไขได้
    if (status == null) return true;

    // ถ้าอนุมัติบางส่วนแล้ว ไม่ให้แก้ไข
    if (status.currentApprovedLevel > 0) return false;

    // ถูก reject แก้ไขได้
    if (status.status == POApprovalStatus.rejected) return true;

    // รออนุมัติ ไม่ให้แก้ไข
    if (status.status == POApprovalStatus.pending) return false;

    // อนุมัติแล้ว ไม่ให้แก้ไข
    if (status.status == POApprovalStatus.autoApproved ||
        status.status == POApprovalStatus.approved) {
      return false;
    }

    return true;
  }

  /// ตรวจสอบว่าต้องแสดง warning เมื่อแก้ไข PO หรือไม่
  static bool shouldShowEditWarning(POApprovalStatusModel? status) {
    if (status == null) return false;

    // ถ้าอนุมัติบางส่วนแล้ว แสดง warning
    if (status.currentApprovedLevel > 0) return true;

    switch (status.status) {
      case POApprovalStatus.pending:
      case POApprovalStatus.approved:
      case POApprovalStatus.autoApproved:
        return true;
      default:
        return false;
    }
  }

  /// ดึงข้อความ warning สำหรับการแก้ไข
  static String getEditWarningMessage(POApprovalStatusModel? status) {
    if (status == null) return '';

    if (status.currentApprovedLevel > 0) {
      return 'เอกสารถูกอนุมัติบางส่วนแล้ว (ระดับ ${status.currentApprovedLevel}) ไม่สามารถแก้ไขได้';
    }

    switch (status.status) {
      case POApprovalStatus.pending:
        return global.language('doc_pending_approval_cannot_edit');
      case POApprovalStatus.approved:
      case POApprovalStatus.autoApproved:
        return global.language('approved_cannot_edit');
      default:
        return '';
    }
  }
}
