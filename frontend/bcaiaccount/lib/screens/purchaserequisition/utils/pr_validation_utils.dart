import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/approval_model.dart';
import '../../../global.dart' as global;

/// PR Validation Result
class PRValidationResult {
  final bool isValid;
  final String? errorMessage;

  PRValidationResult.success()
      : isValid = true,
        errorMessage = null;

  PRValidationResult.error(this.errorMessage) : isValid = false;
}

/// PR Validation Utilities
/// Helper functions สำหรับ validate ข้อมูล PR (ใบขอซื้อ)
class PRValidationUtils {
  /// ตรวจสอบว่าเลือก Supplier แล้วหรือยัง
  static PRValidationResult validateSupplier(TransactionModel screenData) {
    if (screenData.custcode.isEmpty) {
      return PRValidationResult.error('กรุณาเลือกผู้ขาย/Supplier');
    }
    return PRValidationResult.success();
  }

  /// ตรวจสอบว่ามีรายการสินค้าหรือไม่
  static PRValidationResult validateProducts(TransactionModel screenData) {
    if (screenData.details == null || screenData.details!.isEmpty) {
      return PRValidationResult.error(global.language('please_add_product_items'));
    }
    return PRValidationResult.success();
  }

  /// ตรวจสอบว่ามี purpose (วัตถุประสงค์) หรือไม่
  static PRValidationResult validatePurpose(String? purpose) {
    if (purpose == null || purpose.trim().isEmpty) {
      return PRValidationResult.error(global.language('purpose'));
    }
    return PRValidationResult.success();
  }

  /// ตรวจสอบ PR ก่อนบันทึก
  static PRValidationResult validateBeforeSave(TransactionModel screenData) {
    // ตรวจสอบสินค้า
    final productsResult = validateProducts(screenData);
    if (!productsResult.isValid) return productsResult;

    return PRValidationResult.success();
  }

  /// ตรวจสอบ PR ก่อนส่งอนุมัติ
  static PRValidationResult validateBeforeSubmitApproval(
    TransactionModel screenData, {
    POApprovalStatusModel? approvalStatus,
  }) {
    // ตรวจสอบพื้นฐาน
    final basicResult = validateBeforeSave(screenData);
    if (!basicResult.isValid) return basicResult;

    // ต้องมี docno (ต้องบันทึกก่อน)
    if (screenData.docno.isEmpty) {
      return PRValidationResult.error(global.language('please_save_before_submit'));
    }

    // ตรวจสอบสถานะการอนุมัติ
    if (approvalStatus != null) {
      if (approvalStatus.currentApprovedLevel > 0) {
        return PRValidationResult.error(
            global.language('doc_partially_approved_cannot_resubmit'));
      }

      if (approvalStatus.status == POApprovalStatus.approved ||
          approvalStatus.status == POApprovalStatus.autoApproved) {
        return PRValidationResult.error(global.language('document_approved'));
      }
    }

    return PRValidationResult.success();
  }

  /// ตรวจสอบว่าสามารถยกเลิก PR ได้หรือไม่
  static PRValidationResult validateBeforeCancel(
    TransactionModel screenData, {
    POApprovalStatusModel? approvalStatus,
  }) {
    if (screenData.iscancel == true) {
      return PRValidationResult.error(global.language('document_cancelled'));
    }

    if (screenData.isref == true) {
      return PRValidationResult.error(global.language('doc_referenced_cannot_cancel'));
    }

    if (approvalStatus != null) {
      if (approvalStatus.status == POApprovalStatus.pending) {
        return PRValidationResult.error(
            global.language('doc_pending_withdraw_first'));
      }
    }

    return PRValidationResult.success();
  }
}
