import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:smlaicloud/services/pdf_service.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_ui_utils.dart';
import 'package:smlaicloud/global.dart' as global;

/// Manager สำหรับจัดการ Workflow การ Save/Print/Approval ของใบเสนอราคา
///
/// รับผิดชอบ:
/// - จัดการ flow หลังบันทึกสำเร็จ (submit approval → print หรือแสดง dialog)
/// - จัดการ flow หลัง update สำเร็จ
/// - เปิด PDF หลังบันทึก
/// - แสดง dialog เมื่อ docno ซ้ำ
///
/// เหมือน PO: มีระบบอนุมัติ — ส่งอนุมัติหลังบันทึก
class QTWorkflowManager {
  static const String _printAfterSaveKey = 'transaction_print_after_save_flag';

  final BuildContext context;
  final bool Function() isMounted;
  final String Function() getDocNo;
  final String Function(global.TransactionTypeEnum) getCollectionName;
  final String Function(global.TransactionTypeEnum) getDocumentTitle;
  final Future<bool> Function() submitQTForApproval;
  final VoidCallback onClearScreen;
  final void Function(String docNo) onShowSaveDialog;
  final void Function() onRegenerateDocNoAndSave;
  final TabController editTabController;
  final bool Function() checkIsMultiCurrency;

  QTWorkflowManager({
    required this.context,
    required this.isMounted,
    required this.getDocNo,
    required this.getCollectionName,
    required this.getDocumentTitle,
    required this.submitQTForApproval,
    required this.onClearScreen,
    required this.onShowSaveDialog,
    required this.onRegenerateDocNoAndSave,
    required this.editTabController,
    required this.checkIsMultiCurrency,
  });

  /// ส่งใบเสนอราคาเข้าระบบอนุมัติและจัดการพิมพ์ตามสถานะ (สำหรับ Save ใหม่)
  /// - ถ้า auto_approved → พิมพ์อัตโนมัติทันที
  /// - ถ้าไม่ใช่ auto_approved → ใช้ flow เดิม (ตรวจสอบ checkbox จาก SharedPreferences)
  Future<void> submitAndHandlePrint(String docNo, global.TransactionTypeEnum transType) async {
    AppLogger.debug('[QTWorkflow] submitAndHandlePrint START - docNo: $docNo');

    // ส่งใบเสนอราคาเข้าระบบอนุมัติและรอผลลัพธ์
    final isAutoApproved = await submitQTForApproval();
    AppLogger.debug('[QTWorkflow] submitAndHandlePrint isAutoApproved: $isAutoApproved');

    if (isAutoApproved && isMounted()) {
      // กรณี auto_approved → พิมพ์อัตโนมัติทันที
      AppLogger.debug('[QTWorkflow] Auto-approved, opening PDF...');

      final prefs = await SharedPreferences.getInstance();
      await prefs.setBool(_printAfterSaveKey, false);

      if (isMounted()) {
        context.ui.showSuccess('${global.language("save_success")} - ${global.language("docno")}: $docNo');
      }

      await openPdfAfterSave(docNo, transType);
    } else {
      // กรณีไม่ใช่ auto_approved → ใช้ flow เดิม (ตรวจสอบ checkbox)
      await handleSaveSuccessWithPrintCheck(docNo, transType);
    }
  }

  /// ส่งใบเสนอราคาเข้าระบบอนุมัติและจัดการพิมพ์ตามสถานะ (สำหรับ Update)
  Future<void> submitAndHandleUpdatePrint(global.TransactionTypeEnum transType) async {
    final docNo = getDocNo();
    AppLogger.debug('[QTWorkflow] submitAndHandleUpdatePrint START - docNo: $docNo');

    final isAutoApproved = await submitQTForApproval();
    AppLogger.debug('[QTWorkflow] submitAndHandleUpdatePrint isAutoApproved: $isAutoApproved');

    if (isAutoApproved && isMounted()) {
      AppLogger.debug('[QTWorkflow] Auto-approved, opening PDF...');

      final prefs = await SharedPreferences.getInstance();
      await prefs.setBool(_printAfterSaveKey, false);

      if (isMounted()) {
        context.ui.showSuccess('${global.language("update_success")} - ${global.language("docno")}: $docNo');
      }

      await openPdfAfterSave(docNo, transType);
    } else {
      await handleUpdateSuccessWithPrintCheck(transType);
    }
  }

  /// จัดการหลังบันทึกสำเร็จ - อ่านค่าจาก SharedPreferences และตัดสินใจว่าจะพิมพ์หรือไม่
  Future<void> handleSaveSuccessWithPrintCheck(String docNo, global.TransactionTypeEnum transType) async {
    final prefs = await SharedPreferences.getInstance();
    final printAfterSave = prefs.getBool(_printAfterSaveKey) ?? false;

    // reset flag หลังจากอ่านค่าแล้ว
    await prefs.setBool(_printAfterSaveKey, false);

    if (printAfterSave && isMounted()) {
      // แสดงข้อความแจ้งเตือนสำเร็จ
      context.ui.showSuccess('${global.language("save_success")} - ${global.language("docno")}: $docNo');

      // เปิด PDF จาก backend
      await openPdfAfterSave(docNo, transType);
    } else {
      // แสดง dialog ปกติ (กรณีไม่เลือกพิมพ์)
      if (isMounted()) {
        onShowSaveDialog(docNo);
      }
    }
  }

  /// จัดการหลัง Update สำเร็จ - อ่านค่าจาก SharedPreferences และตัดสินใจว่าจะพิมพ์หรือไม่
  Future<void> handleUpdateSuccessWithPrintCheck(global.TransactionTypeEnum transType) async {
    final prefs = await SharedPreferences.getInstance();
    final printAfterSave = prefs.getBool(_printAfterSaveKey) ?? false;

    await prefs.setBool(_printAfterSaveKey, false);

    if (printAfterSave && isMounted()) {
      final docNo = getDocNo();
      context.ui.showSuccess('${global.language("edit_success")} - ${global.language("docno")}: $docNo');

      await openPdfAfterSave(docNo, transType);
    } else {
      if (isMounted()) {
        onClearScreen();
        context.ui.showInfo(global.language("edit_success"));
      }
    }
  }

  /// เปิด PDF หลังจากบันทึกเอกสารสำเร็จ
  Future<void> openPdfAfterSave(String docNo, global.TransactionTypeEnum transType) async {
    final isMultiCurrency = checkIsMultiCurrency();
    AppLogger.debug('[QTWorkflow] openPdfAfterSave - docNo: $docNo, isMultiCurrency: $isMultiCurrency');

    if (!isMounted()) return;

    PdfService pdfService = PdfService();
    String collection = getCollectionName(transType);

    if (collection.isNotEmpty) {
      try {
        await pdfService.generateAndOpenPdf(
          collection: collection,
          docNo: docNo,
          title: getDocumentTitle(transType),
          context: context,
          isMultiCurrency: isMultiCurrency,
        );
      } catch (e) {
        AppLogger.error('[QTWorkflow] เปิด PDF ไม่สำเร็จ: $e');
      }
    }

    // หลังจากปิด PDF แล้ว ให้ clear หน้าจอ
    if (isMounted()) {
      onClearScreen();
      editTabController.animateTo(0);
    }
  }

  /// แสดง dialog เมื่อ docno ซ้ำ
  void showDuplicateDocNoDialog(String currentDocNo) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext dialogContext) {
        return AlertDialog(
          title: Row(
            children: [
              Icon(Icons.warning_amber_rounded, color: Colors.orange.shade700, size: 28),
              SizedBox(width: 8),
              Text(global.language('duplicate_document_number')),
            ],
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('${global.language("document_number")} "$currentDocNo" ${global.language("already_exists_in_system")}'),
              SizedBox(height: 12),
              Text(global.language('want_to_generate_new_docno_and_save')),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(dialogContext).pop(),
              child: Text(global.language('cancel')),
            ),
            ElevatedButton(
              onPressed: () {
                Navigator.of(dialogContext).pop();
                onRegenerateDocNoAndSave();
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.blue,
                foregroundColor: Colors.white,
              ),
              child: Text(global.language('generate_new_docno_and_save')),
            ),
          ],
        );
      },
    );
  }
}
