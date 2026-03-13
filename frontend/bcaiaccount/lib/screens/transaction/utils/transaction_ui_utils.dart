// ignore_for_file: deprecated_member_use

import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

/// ============================================
/// Transaction UI Utilities - รวมศูนย์ UI Helpers
/// ============================================
/// 
/// ใช้แทนการเรียก ScaffoldMessenger, SnackBar, Dialog กระจายทั่วไฟล์
/// เพื่อให้ UI Consistent และ maintain ได้ง่าย

class TransactionUIUtils {
  final BuildContext context;

  TransactionUIUtils(this.context);

  // ============================================
  // SnackBar Methods
  // ============================================

  /// แสดง SnackBar สำเร็จ
  void showSuccess(String message, {IconData icon = Icons.check_circle}) {
    _showSnackBar(
      message: message,
      icon: icon,
      backgroundColor: global.theme.positiveHighlightTextColor,
      textColor: global.theme.onPrimaryColor,
    );
  }

  /// แสดง SnackBar ข้อผิดพลาด
  void showError(String message, {IconData icon = Icons.error}) {
    _showSnackBar(
      message: message,
      icon: icon,
      backgroundColor: global.theme.negativeHighlightTextColor,
      textColor: global.theme.onPrimaryColor,
    );
  }

  /// แสดง SnackBar คำเตือน
  void showWarning(String message, {IconData icon = Icons.warning_amber}) {
    _showSnackBar(
      message: message,
      icon: icon,
      backgroundColor: global.theme.warningHighlightColor,
      textColor: global.theme.warningHighlightTextColor,
      duration: Duration(seconds: 5),
      action: SnackBarAction(
        label: global.language('understood'),
        textColor: global.theme.warningHighlightTextColor,
        onPressed: () {},
      ),
    );
  }

  /// แสดง SnackBar ข้อมูล
  void showInfo(String message, {IconData icon = Icons.info}) {
    _showSnackBar(
      message: message,
      icon: icon,
      backgroundColor: global.theme.infoHighlightTextColor,
      textColor: global.theme.onPrimaryColor,
    );
  }

  void _showSnackBar({
    required String message,
    required IconData icon,
    required Color backgroundColor,
    required Color textColor,
    Duration duration = const Duration(seconds: 3),
    SnackBarAction? action,
  }) {
    if (!context.mounted) return;

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            Icon(icon, color: textColor),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                message,
                style: TextStyle(color: textColor),
              ),
            ),
          ],
        ),
        backgroundColor: backgroundColor,
        duration: duration,
        action: action,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      ),
    );
  }

  // ============================================
  // Dialog Methods
  // ============================================

  /// แสดง Dialog ยืนยัน
  Future<bool?> showConfirmDialog({
    required String title,
    required String message,
    String? confirmText,
    String? cancelText,
    Color? confirmColor,
  }) async {
    return showDialog<bool?>(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(title),
          content: Text(message),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: Text(cancelText ?? global.language("cancel")),
            ),
            TextButton(
              onPressed: () => Navigator.pop(context, true),
              style: TextButton.styleFrom(foregroundColor: confirmColor ?? global.theme.negativeHighlightTextColor),
              child: Text(confirmText ?? global.language("confirm")),
            ),
          ],
        );
      },
    );
  }

  /// แสดง Dialog ยืนยันการลบ
  Future<bool?> showDeleteConfirm(String docno) async {
    return showConfirmDialog(
      title: global.language("alert"),
      message: '${global.language("confirm_delete")} ${docno.isNotEmpty ? '"$docno"' : ''}?',
      confirmText: global.language("delete"),
      confirmColor: global.theme.negativeHighlightTextColor,
    );
  }

  /// แสดง Dialog กรอกข้อความ
  Future<String?> showInputDialog({
    required String title,
    String? initialValue,
    String? labelText,
    int maxLines = 1,
    TextInputType? keyboardType,
  }) async {
    final controller = TextEditingController(text: initialValue);

    return showDialog<String?>(
      context: context,
      barrierDismissible: true,
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(title),
          content: SizedBox(
            width: global.isMobileScreen(context) ? 350 : 400,
            child: TextField(
              autofocus: true,
              controller: controller,
              maxLines: maxLines,
              keyboardType: keyboardType,
              decoration: InputDecoration(
                labelText: labelText,
                border: const OutlineInputBorder(),
                suffixIcon: IconButton(
                  icon: Icon(Icons.clear),
                  onPressed: () => controller.clear(),
                ),
              ),
              onSubmitted: (value) => Navigator.pop(context, controller.text),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, initialValue),
              child: Text(global.language("cancel")),
            ),
            TextButton(
              onPressed: () => Navigator.pop(context, controller.text),
              child: Text(global.language("update")),
            ),
          ],
        );
      },
    );
  }

  /// แสดง Dialog กรอกตัวเลข (ใช้ global.showNumericInputDialog)
  Future<String?> showNumericInput({
    required String title,
    String? initialValue,
    int maxLength = 12,
    int decimalPlaces = 2,
    bool allowNegative = false,
  }) async {
    return global.showNumericInputDialog(
      context,
      title: title,
      initialValue: initialValue ?? '',
      maxLength: maxLength,
      decimalPlaces: decimalPlaces,
      allowNegative: allowNegative,
    );
  }

  /// แสดง Dialog Loading
  void showLoading({String? message}) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext context) {
        return Dialog(
          child: Padding(
            padding: EdgeInsets.all(20),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                const CircularProgressIndicator(),
                SizedBox(width: 16),
                Text(message ?? '${global.language('loading')}...'),
              ],
            ),
          ),
        );
      },
    );
  }

  /// ปิด Dialog Loading
  void hideLoading() {
    if (Navigator.canPop(context)) {
      Navigator.pop(context);
    }
  }

  // ============================================
  // Generic Selection Dialog
  // ============================================

  /// Generic Selection Dialog สำหรับเลือกรายการต่างๆ
  Future<T?> showSelectionDialog<T>({
    required String title,
    required List<T> items,
    required String Function(T) getTitle,
    required String Function(T) getSubtitle,
    IconData leadingIcon = Icons.check_circle_outline,
    Color? headerColor,
  }) async {
    final Color effectiveHeaderColor = headerColor ?? global.theme.primaryColor;
    return showDialog<T?>(
      context: context,
      barrierDismissible: true,
      builder: (BuildContext context) {
        return Dialog(
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
          child: Container(
            width: global.isMobileScreen(context) ? 350 : 420,
            constraints: const BoxConstraints(maxHeight: 500),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Header
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: effectiveHeaderColor,
                    borderRadius: const BorderRadius.only(
                      topLeft: Radius.circular(16),
                      topRight: Radius.circular(16),
                    ),
                  ),
                  child: Row(
                    children: [
                      Icon(leadingIcon, color: global.theme.onPrimaryColor, size: 24),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          title,
                          style: TextStyle(
                            color: global.theme.onPrimaryColor,
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
                // Content
                Flexible(
                  child: ListView.separated(
                    shrinkWrap: true,
                    padding: const EdgeInsets.symmetric(vertical: 8),
                    itemCount: items.length,
                    separatorBuilder: (context, index) => const Divider(height: 1),
                    itemBuilder: (BuildContext context, int index) {
                      final item = items[index];
                      return Material(
                        color: Colors.transparent,
                        child: InkWell(
                          onTap: () => Navigator.pop(context, item),
                          child: Padding(
                            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                            child: Row(
                              children: [
                                Container(
                                  padding: const EdgeInsets.all(8),
                                  decoration: BoxDecoration(
                                    color: effectiveHeaderColor.withValues(alpha: 0.1),
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  child: Icon(leadingIcon, color: effectiveHeaderColor, size: 20),
                                ),
                                const SizedBox(width: 12),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        getTitle(item),
                                        style: TextStyle(
                                          fontWeight: FontWeight.bold,
                                          fontSize: 14,
                                          color: effectiveHeaderColor,
                                        ),
                                      ),
                                      const SizedBox(height: 2),
                                      Text(
                                        getSubtitle(item),
                                        style: TextStyle(
                                          fontSize: 13,
                                          color: global.theme.textSecondaryColor,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                                Icon(Icons.chevron_right, color: global.theme.iconSecondaryColor),
                              ],
                            ),
                          ),
                        ),
                      );
                    },
                  ),
                ),
                // Footer
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    border: Border(top: BorderSide(color: global.theme.dividerBorderColor)),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      TextButton(
                        onPressed: () => Navigator.pop(context, null),
                        child: Text(global.language("close")),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

/// Extension สำหรับ BuildContext เพื่อใช้งานง่ายขึ้น
extension TransactionUIExtension on BuildContext {
  TransactionUIUtils get ui => TransactionUIUtils(this);
}
