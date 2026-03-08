import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/approval_model.dart';
import '../../../global.dart' as global;

/// PO Validation Result
class POValidationResult {
  final bool isValid;
  final String? errorMessage;

  POValidationResult.success()
      : isValid = true,
        errorMessage = null;

  POValidationResult.error(this.errorMessage) : isValid = false;
}

/// PO Validation Utilities
/// Helper functions สำหรับ validate ข้อมูล PO
class POValidationUtils {
  /// ตรวจสอบว่าเลือกประเภทการจัดซื้อแล้วหรือยัง
  static POValidationResult validatePurchaseType(String? purchaseTypeCode) {
    if (purchaseTypeCode == null || purchaseTypeCode.isEmpty) {
      return POValidationResult.error(global.language('please_select_purchase_type'));
    }
    return POValidationResult.success();
  }

  /// ตรวจสอบว่าเลือก Supplier แล้วหรือยัง
  static POValidationResult validateSupplier(TransactionModel screenData) {
    if (screenData.custcode.isEmpty) {
      return POValidationResult.error('กรุณาเลือกผู้ขาย/Supplier');
    }
    return POValidationResult.success();
  }

  /// ตรวจสอบว่ามีรายการสินค้าหรือไม่
  static POValidationResult validateProducts(TransactionModel screenData) {
    if (screenData.details == null || screenData.details!.isEmpty) {
      return POValidationResult.error(global.language('please_add_product_items'));
    }
    return POValidationResult.success();
  }

  /// ตรวจสอบว่ายอดรวมมากกว่า 0 หรือไม่
  static POValidationResult validateTotalAmount(TransactionModel screenData) {
    if (screenData.totalaftervat <= 0) {
      return POValidationResult.error(global.language('total_must_be_greater_than_zero'));
    }
    return POValidationResult.success();
  }

  /// ตรวจสอบ PO ก่อนบันทึก
  static POValidationResult validateBeforeSave(
    TransactionModel screenData, {
    String? purchaseTypeCode,
  }) {
    // ตรวจสอบ supplier
    final supplierResult = validateSupplier(screenData);
    if (!supplierResult.isValid) return supplierResult;

    // ตรวจสอบสินค้า
    final productsResult = validateProducts(screenData);
    if (!productsResult.isValid) return productsResult;

    // ตรวจสอบประเภทการจัดซื้อ (ถ้าระบุ)
    if (purchaseTypeCode != null) {
      final typeResult = validatePurchaseType(purchaseTypeCode);
      if (!typeResult.isValid) return typeResult;
    }

    return POValidationResult.success();
  }

  /// ตรวจสอบ PO ก่อนส่งอนุมัติ
  static POValidationResult validateBeforeSubmitApproval(
    TransactionModel screenData, {
    required String? purchaseTypeCode,
    POApprovalStatusModel? approvalStatus,
  }) {
    // ตรวจสอบพื้นฐานก่อน
    final basicResult = validateBeforeSave(
      screenData,
      purchaseTypeCode: purchaseTypeCode,
    );
    if (!basicResult.isValid) return basicResult;

    // ต้องมี docno (ต้องบันทึกก่อน)
    if (screenData.docno.isEmpty) {
      return POValidationResult.error(global.language('please_save_before_submit'));
    }

    // ต้องเลือกประเภทการจัดซื้อ
    final typeResult = validatePurchaseType(purchaseTypeCode);
    if (!typeResult.isValid) return typeResult;

    // ตรวจสอบสถานะการอนุมัติ
    if (approvalStatus != null) {
      // ถ้าอนุมัติบางส่วนแล้ว ไม่ให้ส่งใหม่
      if (approvalStatus.currentApprovedLevel > 0) {
        return POValidationResult.error(
            global.language('doc_partially_approved_cannot_resubmit'));
      }

      // ถ้าอนุมัติครบแล้ว ไม่ให้ส่งใหม่
      if (approvalStatus.status == POApprovalStatus.approved ||
          approvalStatus.status == POApprovalStatus.autoApproved) {
        return POValidationResult.error(global.language('document_approved'));
      }
    }

    return POValidationResult.success();
  }

  /// ตรวจสอบว่าสามารถยกเลิก PO ได้หรือไม่
  static POValidationResult validateBeforeCancel(
    TransactionModel screenData, {
    POApprovalStatusModel? approvalStatus,
  }) {
    // ยกเลิกแล้ว ไม่สามารถยกเลิกซ้ำ
    if (screenData.iscancel == true) {
      return POValidationResult.error(global.language('document_cancelled'));
    }

    // อ้างอิงไปแล้ว ไม่สามารถยกเลิก
    if (screenData.isref == true) {
      return POValidationResult.error(global.language('doc_referenced_cannot_cancel'));
    }

    // ตรวจสอบสถานะการอนุมัติ
    if (approvalStatus != null) {
      // รออนุมัติ ไม่ให้ยกเลิก (ต้องถอนการส่งอนุมัติก่อน)
      if (approvalStatus.status == POApprovalStatus.pending) {
        return POValidationResult.error(
            global.language('doc_pending_withdraw_first'));
      }
    }

    return POValidationResult.success();
  }
}
