import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:smlaicloud/services/pdf_service.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/global.dart' as global;

/// Controller สำหรับจัดการ Actions ของ PO (Toolbar, Save, Print)
///
/// รับผิดชอบ:
/// - จัดการการส่งอนุมัติ
/// - จัดการการพิมพ์หลังบันทึก
/// - จัดการ checkbox "พิมพ์หลังบันทึก"
class POActionController {
  static const String printAfterSaveKey = 'transaction_print_after_save_flag';

  final BuildContext context;
  final bool Function() isMounted;
  final VoidCallback onClearScreen;
  final void Function(String docNo) onShowSaveDialog;

  POActionController({
    required this.context,
    required this.isMounted,
    required this.onClearScreen,
    required this.onShowSaveDialog,
  });

  /// ส่ง PO เข้าระบบอนุมัติและจัดการพิมพ์ตามสถานะ
  /// - ถ้า auto_approved → พิมพ์อัตโนมัติทันที
  /// - ถ้าไม่ใช่ auto_approved → ใช้ flow เดิม (ตรวจสอบ checkbox)
  Future<void> submitAndHandlePrint({
    required String docNo,
    required global.TransactionTypeEnum transType,
    required Future<bool> Function() submitForApproval,
  }) async {
    AppLogger.debug('🖨️ [submitAndHandlePrint] START - docNo: $docNo');

    final isAutoApproved = await submitForApproval();
    AppLogger.debug('🖨️ [submitAndHandlePrint] isAutoApproved: $isAutoApproved');

    if (isAutoApproved && isMounted()) {
      AppLogger.debug('🖨️ [submitAndHandlePrint] Auto-approved, opening PDF...');

      // Clear SharedPreferences flag
      final prefs = await SharedPreferences.getInstance();
      await prefs.setBool(printAfterSaveKey, false);

      // แสดงข้อความแจ้งเตือนสำเร็จ
      _showSuccess('${global.language("save_success")} - ${global.language("docno")}: $docNo');

      // พิมพ์อัตโนมัติ
      await openPdfAfterSave(docNo, transType);
    } else {
      // ใช้ flow เดิม (ตรวจสอบ checkbox)
      await handleSaveSuccessWithPrintCheck(docNo, transType);
    }
  }

  /// ส่ง PO เข้าระบบอนุมัติและจัดการพิมพ์ (สำหรับ Update)
  Future<void> submitAndHandleUpdatePrint({
    required String docNo,
    required global.TransactionTypeEnum transType,
    required Future<bool> Function() submitForApproval,
  }) async {
    AppLogger.debug('🖨️ [submitAndHandleUpdatePrint] START - docNo: $docNo');

    final isAutoApproved = await submitForApproval();
    AppLogger.debug('🖨️ [submitAndHandleUpdatePrint] isAutoApproved: $isAutoApproved');

    if (isAutoApproved && isMounted()) {
      AppLogger.debug('🖨️ [submitAndHandleUpdatePrint] Auto-approved, opening PDF...');

      final prefs = await SharedPreferences.getInstance();
      await prefs.setBool(printAfterSaveKey, false);

      _showSuccess('${global.language("update_success")} - ${global.language("docno")}: $docNo');

      await openPdfAfterSave(docNo, transType);
    } else {
      await handleUpdateSuccessWithPrintCheck(docNo, transType);
    }
  }

  /// จัดการหลังบันทึกสำเร็จ - ตรวจสอบว่าจะพิมพ์หรือไม่
  Future<void> handleSaveSuccessWithPrintCheck(String docNo, global.TransactionTypeEnum transType) async {
    AppLogger.debug('🖨️ [handleSaveSuccessWithPrintCheck] START - docNo: $docNo');

    final prefs = await SharedPreferences.getInstance();
    final printAfterSave = prefs.getBool(printAfterSaveKey) ?? false;
    AppLogger.debug('🖨️ [handleSaveSuccessWithPrintCheck] printAfterSave: $printAfterSave');

    // Reset flag
    await prefs.setBool(printAfterSaveKey, false);

    if (printAfterSave && isMounted()) {
      AppLogger.debug('🖨️ [handleSaveSuccessWithPrintCheck] จะเปิด PDF...');
      _showSuccess('${global.language("save_success")} - ${global.language("docno")}: $docNo');
      await openPdfAfterSave(docNo, transType);
    } else {
      AppLogger.debug('🖨️ [handleSaveSuccessWithPrintCheck] แสดง dialog ปกติ');
      if (isMounted()) {
        onShowSaveDialog(docNo);
      }
    }
  }

  /// จัดการหลัง Update สำเร็จ
  Future<void> handleUpdateSuccessWithPrintCheck(String docNo, global.TransactionTypeEnum transType) async {
    AppLogger.debug('🖨️ [handleUpdateSuccessWithPrintCheck] START');

    final prefs = await SharedPreferences.getInstance();
    final printAfterSave = prefs.getBool(printAfterSaveKey) ?? false;
    AppLogger.debug('🖨️ [handleUpdateSuccessWithPrintCheck] printAfterSave: $printAfterSave');

    await prefs.setBool(printAfterSaveKey, false);

    if (printAfterSave && isMounted()) {
      AppLogger.debug('🖨️ [handleUpdateSuccessWithPrintCheck] จะเปิด PDF...');
      _showSuccess('${global.language("edit_success")} - ${global.language("docno")}: $docNo');
      await openPdfAfterSave(docNo, transType);
    } else {
      AppLogger.debug('🖨️ [handleUpdateSuccessWithPrintCheck] ไม่พิมพ์ - clear screen');
      if (isMounted()) {
        onClearScreen();
        _showInfo(global.language("edit_success"));
      }
    }
  }

  /// เปิด PDF หลังจากบันทึกเอกสารสำเร็จ
  Future<void> openPdfAfterSave(String docNo, global.TransactionTypeEnum transType, {bool isMultiCurrency = false}) async {
    AppLogger.debug('🖨️ [openPdfAfterSave] START - docNo: $docNo');

    if (!isMounted()) {
      AppLogger.debug('🖨️ [openPdfAfterSave] Widget not mounted');
      return;
    }

    final pdfService = PdfService();
    final collection = getCollectionName(transType);
    final title = getDocumentTitle(transType);

    AppLogger.debug('🖨️ [openPdfAfterSave] collection: $collection');

    if (collection.isNotEmpty) {
      try {
        await pdfService.generateAndOpenPdf(
          collection: collection,
          docNo: docNo,
          title: title,
          context: context,
          isMultiCurrency: isMultiCurrency,
        );
        AppLogger.debug('🖨️ [openPdfAfterSave] generateAndOpenPdf completed');
      } catch (e) {
        AppLogger.error('🖨️ [openPdfAfterSave] Error: $e');
        _showError('เปิด PDF ไม่สำเร็จ: $e');
      }
    }
  }

  /// ตั้งค่า flag พิมพ์หลังบันทึก
  Future<void> setPrintAfterSave(bool value) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(printAfterSaveKey, value);
    AppLogger.debug('🖨️ [setPrintAfterSave] value: $value');
  }

  /// อ่านค่า flag พิมพ์หลังบันทึก
  Future<bool> getPrintAfterSave() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getBool(printAfterSaveKey) ?? false;
  }

  /// ดึงชื่อ collection สำหรับ MongoDB
  String getCollectionName(global.TransactionTypeEnum type) {
    return "transactionPurchaseOrder";
  }

  /// ดึงชื่อเอกสารสำหรับ PDF title
  String getDocumentTitle(global.TransactionTypeEnum type) {
    return global.language("purchase_order");
  }

  // Helper methods for showing messages
  void _showSuccess(String message) {
    global.showSuccessSnackBar(context, message);
  }

  void _showInfo(String message) {
    global.showInfoSnackBar(context, message);
  }

  void _showError(String message) {
    global.showErrorSnackBar(context, message);
  }
}
