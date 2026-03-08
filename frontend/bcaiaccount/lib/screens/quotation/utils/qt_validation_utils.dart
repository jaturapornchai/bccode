import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/approval_model.dart';
import '../../../global.dart' as global;

/// Quotation Validation Result
class QTValidationResult {
  final bool isValid;
  final String? errorMessage;

  QTValidationResult.success()
      : isValid = true,
        errorMessage = null;

  QTValidationResult.error(this.errorMessage) : isValid = false;
}

/// Quotation Validation Utilities
/// Helper functions สำหรับ validate ข้อมูลใบเสนอราคา
class QTValidationUtils {
  /// ตรวจสอบว่าเลือกประเภทใบเสนอราคาแล้วหรือยัง
  static QTValidationResult validatePurchaseType(String? purchaseTypeCode) {
    if (purchaseTypeCode == null || purchaseTypeCode.isEmpty) {
      return QTValidationResult.error(global.language('please_select_quotation_type'));
    }
    return QTValidationResult.success();
  }

  /// ตรวจสอบว่าเลือกลูกค้าแล้วหรือยัง
  static QTValidationResult validateCustomer(TransactionModel screenData) {
    if (screenData.custcode.isEmpty) {
      return QTValidationResult.error(global.language('coupon_error_no_customer_selected'));
    }
    return QTValidationResult.success();
  }

  /// ตรวจสอบว่ามีรายการสินค้าหรือไม่
  static QTValidationResult validateProducts(TransactionModel screenData) {
    if (screenData.details == null || screenData.details!.isEmpty) {
      return QTValidationResult.error(global.language('please_add_product_items'));
    }
    return QTValidationResult.success();
  }

  /// ตรวจสอบว่ายอดรวมมากกว่า 0 หรือไม่
  static QTValidationResult validateTotalAmount(TransactionModel screenData) {
    if (screenData.totalaftervat <= 0) {
      return QTValidationResult.error(global.language('total_must_be_greater_than_zero'));
    }
    return QTValidationResult.success();
  }

  /// ตรวจสอบใบเสนอราคาก่อนบันทึก
  static QTValidationResult validateBeforeSave(
    TransactionModel screenData, {
    String? purchaseTypeCode,
  }) {
    // ตรวจสอบลูกค้า
    final customerResult = validateCustomer(screenData);
    if (!customerResult.isValid) return customerResult;

    // ตรวจสอบสินค้า
    final productsResult = validateProducts(screenData);
    if (!productsResult.isValid) return productsResult;

    // ตรวจสอบประเภทใบเสนอราคา (ถ้าระบุ)
    if (purchaseTypeCode != null) {
      final typeResult = validatePurchaseType(purchaseTypeCode);
      if (!typeResult.isValid) return typeResult;
    }

    return QTValidationResult.success();
  }

  /// ตรวจสอบใบเสนอราคาก่อนส่งอนุมัติ
  static QTValidationResult validateBeforeSubmitApproval(
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
      return QTValidationResult.error(global.language('please_save_before_submit'));
    }

    // ต้องเลือกประเภทใบเสนอราคา
    final typeResult = validatePurchaseType(purchaseTypeCode);
    if (!typeResult.isValid) return typeResult;

    // ตรวจสอบสถานะการอนุมัติ
    if (approvalStatus != null) {
      // ถ้าอนุมัติบางส่วนแล้ว ไม่ให้ส่งใหม่
      if (approvalStatus.currentApprovedLevel > 0) {
        return QTValidationResult.error(
            global.language('doc_partially_approved_cannot_resubmit'));
      }

      // ถ้าอนุมัติครบแล้ว ไม่ให้ส่งใหม่
      if (approvalStatus.status == POApprovalStatus.approved ||
          approvalStatus.status == POApprovalStatus.autoApproved) {
        return QTValidationResult.error(global.language('document_approved'));
      }
    }

    return QTValidationResult.success();
  }

  /// ตรวจสอบว่าสามารถยกเลิกใบเสนอราคาได้หรือไม่
  static QTValidationResult validateBeforeCancel(
    TransactionModel screenData, {
    POApprovalStatusModel? approvalStatus,
  }) {
    if (screenData.iscancel == true) {
      return QTValidationResult.error(global.language('document_cancelled'));
    }

    if (screenData.isref == true) {
      return QTValidationResult.error(global.language('doc_referenced_cannot_cancel'));
    }

    // ตรวจสอบสถานะการอนุมัติ
    if (approvalStatus != null) {
      if (approvalStatus.status == POApprovalStatus.pending) {
        return QTValidationResult.error(
            global.language('doc_pending_withdraw_first'));
      }
    }

    return QTValidationResult.success();
  }
}
