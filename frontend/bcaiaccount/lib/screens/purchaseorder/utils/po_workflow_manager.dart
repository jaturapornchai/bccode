import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:smlaicloud/services/pdf_service.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_ui_utils.dart';
import 'package:smlaicloud/global.dart' as global;

/// Manager สำหรับจัดการ Workflow การ Save/Print/Approval ของ PO
///
/// รับผิดชอบ:
/// - จัดการ flow หลังบันทึกสำเร็จ (print หรือแสดง dialog)
/// - จัดการ flow หลัง update สำเร็จ
/// - เปิด PDF หลังบันทึก
/// - แสดง dialog เมื่อ docno ซ้ำ
class POWorkflowManager {
  static const String _printAfterSaveKey = 'transaction_print_after_save_flag';

  final BuildContext context;
  final bool Function() isMounted;
  final String Function() getDocNo;
  final String Function(global.TransactionTypeEnum) getCollectionName;
  final String Function(global.TransactionTypeEnum) getDocumentTitle;
  final Future<bool> Function() submitPOForApproval;
  final VoidCallback onClearScreen;
  final void Function(String docNo) onShowSaveDialog;
  final void Function() onRegenerateDocNoAndSave;
  final TabController editTabController;
  final bool Function() checkIsMultiCurrency;

  POWorkflowManager({
    required this.context,
    required this.isMounted,
    required this.getDocNo,
    required this.getCollectionName,
    required this.getDocumentTitle,
    required this.submitPOForApproval,
    required this.onClearScreen,
    required this.onShowSaveDialog,
    required this.onRegenerateDocNoAndSave,
    required this.editTabController,
    required this.checkIsMultiCurrency,
  });

  /// ส่ง PO เข้าระบบอนุมัติและจัดการพิมพ์ตามสถานะ (สำหรับ Save ใหม่)
  /// - ถ้า auto_approved → พิมพ์อัตโนมัติทันที
  /// - ถ้าไม่ใช่ auto_approved → ใช้ flow เดิม (ตรวจสอบ checkbox จาก SharedPreferences)
  Future<void> submitAndHandlePrint(String docNo, global.TransactionTypeEnum transType) async {
    AppLogger.debug('🖨️ [submitAndHandlePrint] START - docNo: $docNo');

    // ส่ง PO เข้าระบบอนุมัติและรอผลลัพธ์
    final isAutoApproved = await submitPOForApproval();
    AppLogger.debug('🖨️ [submitAndHandlePrint] isAutoApproved: $isAutoApproved');

    if (isAutoApproved && isMounted()) {
      // กรณี auto_approved → พิมพ์อัตโนมัติทันที
      AppLogger.debug('🖨️ [submitAndHandlePrint] Auto-approved, opening PDF...');

      // Clear SharedPreferences flag (ถ้ามี)
      final prefs = await SharedPreferences.getInstance();
      await prefs.setBool(_printAfterSaveKey, false);

      // แสดงข้อความแจ้งเตือนสำเร็จ
      if (isMounted()) {
        context.ui.showSuccess('${global.language("save_success")} - ${global.language("docno")}: $docNo');
      }

      // เปิด PDF จาก backend (พิมพ์อัตโนมัติ)
      await openPdfAfterSave(docNo, transType);
    } else {
      // กรณีไม่ใช่ auto_approved → ใช้ flow เดิม (ตรวจสอบ checkbox)
      await handleSaveSuccessWithPrintCheck(docNo, transType);
    }
  }

  /// ส่ง PO เข้าระบบอนุมัติและจัดการพิมพ์ตามสถานะ (สำหรับ Update)
  Future<void> submitAndHandleUpdatePrint(global.TransactionTypeEnum transType) async {
    final docNo = getDocNo();
    AppLogger.debug('🖨️ [submitAndHandleUpdatePrint] START - docNo: $docNo');

    // ส่ง PO เข้าระบบอนุมัติและรอผลลัพธ์
    final isAutoApproved = await submitPOForApproval();
    AppLogger.debug('🖨️ [submitAndHandleUpdatePrint] isAutoApproved: $isAutoApproved');

    if (isAutoApproved && isMounted()) {
      // กรณี auto_approved → พิมพ์อัตโนมัติทันที
      AppLogger.debug('🖨️ [submitAndHandleUpdatePrint] Auto-approved, opening PDF...');

      // Clear SharedPreferences flag (ถ้ามี)
      final prefs = await SharedPreferences.getInstance();
      await prefs.setBool(_printAfterSaveKey, false);

      // แสดงข้อความแจ้งเตือนสำเร็จ
      if (isMounted()) {
        context.ui.showSuccess('${global.language("update_success")} - ${global.language("docno")}: $docNo');
      }

      // เปิด PDF จาก backend (พิมพ์อัตโนมัติ)
      await openPdfAfterSave(docNo, transType);
    } else {
      // กรณีไม่ใช่ auto_approved → ใช้ flow เดิม (ตรวจสอบ checkbox)
      await handleUpdateSuccessWithPrintCheck(transType);
    }
  }

  /// จัดการหลังบันทึกสำเร็จ - อ่านค่าจาก SharedPreferences และตัดสินใจว่าจะพิมพ์หรือไม่
  Future<void> handleSaveSuccessWithPrintCheck(String docNo, global.TransactionTypeEnum transType) async {
    AppLogger.debug('🖨️ [handleSaveSuccessWithPrintCheck] START - docNo: $docNo');

    // อ่านค่าจาก SharedPreferences (ป้องกัน Hot Reload reset)
    final prefs = await SharedPreferences.getInstance();
    final printAfterSave = prefs.getBool(_printAfterSaveKey) ?? false;
    AppLogger.debug('🖨️ [handleSaveSuccessWithPrintCheck] printAfterSave from prefs: $printAfterSave');

    // reset flag หลังจากอ่านค่าแล้ว
    await prefs.setBool(_printAfterSaveKey, false);

    if (printAfterSave && isMounted()) {
      AppLogger.debug('🖨️ [handleSaveSuccessWithPrintCheck] จะเปิด PDF...');

      // แสดงข้อความแจ้งเตือนสำเร็จ
      context.ui.showSuccess('${global.language("save_success")} - ${global.language("docno")}: $docNo');

      // เปิด PDF จาก backend
      await openPdfAfterSave(docNo, transType);
    } else {
      AppLogger.debug('🖨️ [handleSaveSuccessWithPrintCheck] แสดง dialog ปกติ');
      // แสดง dialog ปกติ (กรณีไม่เลือกพิมพ์)
      if (isMounted()) {
        onShowSaveDialog(docNo);
      }
    }
  }

  /// จัดการหลัง Update สำเร็จ - อ่านค่าจาก SharedPreferences และตัดสินใจว่าจะพิมพ์หรือไม่
  Future<void> handleUpdateSuccessWithPrintCheck(global.TransactionTypeEnum transType) async {
    AppLogger.debug('🖨️ [handleUpdateSuccessWithPrintCheck] START');

    // อ่านค่าจาก SharedPreferences (ป้องกัน Hot Reload reset)
    final prefs = await SharedPreferences.getInstance();
    final printAfterSave = prefs.getBool(_printAfterSaveKey) ?? false;
    AppLogger.debug('🖨️ [handleUpdateSuccessWithPrintCheck] printAfterSave from prefs: $printAfterSave');

    // reset flag หลังจากอ่านค่าแล้ว
    await prefs.setBool(_printAfterSaveKey, false);

    if (printAfterSave && isMounted()) {
      AppLogger.debug('🖨️ [handleUpdateSuccessWithPrintCheck] จะเปิด PDF...');
      final docNo = getDocNo();

      // แสดงข้อความแจ้งเตือนสำเร็จ
      context.ui.showSuccess('${global.language("edit_success")} - ${global.language("docno")}: $docNo');

      // เปิด PDF จาก backend
      await openPdfAfterSave(docNo, transType);
    } else {
      AppLogger.debug('🖨️ [handleUpdateSuccessWithPrintCheck] ไม่พิมพ์ - clear screen');
      // ไม่พิมพ์ ให้ clear screen ปกติ
      if (isMounted()) {
        onClearScreen();
        context.ui.showInfo(global.language("edit_success"));
      }
    }
  }

  /// เปิด PDF หลังจากบันทึกเอกสารสำเร็จ
  Future<void> openPdfAfterSave(String docNo, global.TransactionTypeEnum transType) async {
    final isMultiCurrency = checkIsMultiCurrency();
    AppLogger.debug('🖨️ [openPdfAfterSave] START - docNo: $docNo, transType: $transType, isMultiCurrency: $isMultiCurrency');

    if (!isMounted()) {
      AppLogger.debug('🖨️ [openPdfAfterSave] Widget not mounted, returning');
      return;
    }

    PdfService pdfService = PdfService();
    String collection = getCollectionName(transType);
    AppLogger.debug('🖨️ [openPdfAfterSave] collection: $collection');

    if (collection.isNotEmpty) {
      AppLogger.debug('🖨️ [openPdfAfterSave] Calling generateAndOpenPdf...');
      try {
        await pdfService.generateAndOpenPdf(
          collection: collection,
          docNo: docNo,
          title: getDocumentTitle(transType),
          context: context,
          isMultiCurrency: isMultiCurrency,
        );
        AppLogger.debug('🖨️ [openPdfAfterSave] generateAndOpenPdf completed');
      } catch (e) {
        AppLogger.error('🖨️ [openPdfAfterSave] ERROR: $e');
      }
    } else {
      AppLogger.debug('🖨️ [openPdfAfterSave] collection is empty, skipping PDF');
    }

    // หลังจากปิด PDF แล้ว ให้ clear หน้าจอ
    if (isMounted()) {
      AppLogger.debug('🖨️ [openPdfAfterSave] Clearing screen data...');
      onClearScreen();
      editTabController.animateTo(0);
    }
    AppLogger.debug('🖨️ [openPdfAfterSave] END');
  }

  /// แสดง dialog เมื่อ docno ซ้ำ และให้ผู้ใช้เลือกสร้าง docno ใหม่แล้วบันทึกอีกครั้ง
  void showDuplicateDocNoDialog(String currentDocNo) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext dialogContext) {
        return AlertDialog(
          title: Row(
            children: [
              Icon(Icons.warning_amber_rounded, color: global.theme.warningHighlightTextColor, size: 28),
              SizedBox(width: 8),
              Text(global.language('duplicate_docno')),
            ],
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('${global.language("doc_already_exists")} "$currentDocNo"'),
              SizedBox(height: 12),
              Text(global.language('create_new_docno_confirm')),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () {
                Navigator.of(dialogContext).pop();
              },
              child: Text(global.language('cancel')),
            ),
            ElevatedButton(
              onPressed: () {
                Navigator.of(dialogContext).pop();
                onRegenerateDocNoAndSave();
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.infoHighlightTextColor,
                foregroundColor: global.theme.onPrimaryColor,
              ),
              child: Text(global.language('create_new_docno_and_save')),
            ),
          ],
        );
      },
    );
  }
}
