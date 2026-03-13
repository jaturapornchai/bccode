import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_dialogs.dart';
import 'package:smlaicloud/services/datahistory_api_service.dart';
import 'package:smlaicloud/services/approval_api_service.dart';
import 'package:smlaicloud/widgets/data_history_dialog.dart';
import 'package:smlaicloud/widgets/po_approval_history_dialog.dart';
import 'package:smlaicloud/services/pdf_service.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/global.dart' as global;

/// Widget สำหรับสร้าง AppBar Actions ของ PO Screen
///
/// รับผิดชอบ:
/// - ปุ่มออกใบกำกับเต็ม
/// - ปุ่มดูรูป slip
/// - ปุ่มลบเอกสาร
/// - ปุ่มยกเลิกเอกสาร
/// - ปุ่มดูประวัติการพิมพ์
/// - ปุ่มดูประวัติการแก้ไข
/// - ปุ่มเพิ่มใหม่
/// - ปุ่ม Toggle Preview
/// - ปุ่มค้นหาเอกสาร
/// - ปุ่มบันทึก
class POAppBarActions {
  static const String _printAfterSaveKey = 'transaction_print_after_save_flag';

  /// สร้าง list ของ AppBar actions สำหรับ PO Screen
  static List<Widget> build({
    required BuildContext context,
    required TransactionModel screenData,
    required TransactionModel screenDataTemp,
    required global.TransactionTypeEnum transactionType,
    required bool isLoading,
    required bool showPreview,
    required bool canEditPO,
    required bool canPrintPO,
    required String editDisabledReason,
    required String Function(global.TransactionTypeEnum) getCollectionName,
    required String Function(global.TransactionTypeEnum) getDocumentTitle,
    required TabController tabController,
    required VoidCallback onClearScreen,
    required VoidCallback onTogglePreview,
    required VoidCallback onSearchTrans,
    required VoidCallback onDeleteDoc,
    required void Function({required bool taxInvoice}) onSaveOrUpdate,
    VoidCallback? onManualCloseToggle,
  }) {
    final List<Widget> actions = [];
    final bool hasDoc = (screenData.guidfixed?.isNotEmpty ?? false) && screenData.docno.isNotEmpty;
    final disabledReason = _getDisabledReason(screenData);
    final noDocReason = hasDoc ? null : global.language('no_document_yet');

    // Loading indicator
    if (isLoading) {
      actions.add(_buildLoadingIndicator());
    }

    // ออกใบกำกับเต็ม (สำหรับ POS เท่านั้น — ซ่อนได้เพราะไม่เกี่ยวกับ PO)
    if (screenData.ispos == true && screenData.iscancel == false) {
      actions.add(_buildFullInvoiceButton(context, screenData, onSaveOrUpdate));
    }

    // ดู slip image (ซ่อนได้เพราะไม่เกี่ยวกับ PO)
    if (screenData.slipurl != null && screenData.slipurl != '') {
      actions.add(_buildSlipImageButton(context, screenData));
    }

    // ลบเอกสาร — แสดงเสมอ, disable ถ้าไม่มีเอกสาร/มีการอ้างอิง/ลบแล้ว
    final deleteReason = noDocReason ?? disabledReason;
    actions.add(_buildDeleteButton(context, screenData, onDeleteDoc,
        enabled: deleteReason == null, disabledReason: deleteReason));

    // ยกเลิกเอกสาร — แสดงเสมอ, disable ถ้าไม่มีเอกสาร/ยกเลิกแล้ว/มีการอ้างอิง
    final cancelReason = noDocReason ?? disabledReason;
    actions.add(_buildCancelButton(context, screenData, screenDataTemp, onSaveOrUpdate,
        enabled: cancelReason == null, disabledReason: cancelReason));

    // ปิด/เปิดเอกสารด้วยมือ — แสดงเสมอ, disable ถ้าไม่มีเอกสาร/ลบ/ยกเลิก
    final manualCloseReason = noDocReason ??
        (screenData.isdelete == true ? global.language('doc_deleted') : null) ??
        (screenData.iscancel ? global.language('doc_cancelled') : null);
    actions.add(_buildManualCloseButton(context, screenData, onManualCloseToggle,
        enabled: manualCloseReason == null, disabledReason: manualCloseReason));

    // ประวัติการพิมพ์
    actions.add(_buildPrintHistoryButton(context, screenData, transactionType, getCollectionName, getDocumentTitle,
        enabled: hasDoc, disabledReason: noDocReason));

    // ประวัติการแก้ไข
    actions.add(_buildEditHistoryButton(context, screenData,
        enabled: hasDoc, disabledReason: noDocReason));

    // ประวัติการอนุมัติ
    actions.add(_buildApprovalHistoryButton(context, screenData.docno,
        enabled: hasDoc, disabledReason: noDocReason));

    // ปุ่มเพิ่มใหม่
    actions.add(_buildAddNewButton(onClearScreen));

    // Toggle Preview
    actions.add(_buildTogglePreviewButton(showPreview, onTogglePreview));

    // ค้นหาเอกสาร
    actions.add(_buildSearchButton(onSearchTrans));

    // Switch tabs (mobile)
    final screenWidth = MediaQuery.of(context).size.width;
    if (screenWidth < 800 && showPreview) {
      actions.add(_buildSwitchTabButton(tabController));
    }

    // บันทึก — แสดงเสมอ, disable ถ้าลบ/ยกเลิก/มีการอ้างอิง
    final saveDisabledReason = disabledReason;
    actions.add(_buildSaveButton(
      context,
      screenData,
      canEditPO,
      canPrintPO,
      editDisabledReason,
      onSaveOrUpdate,
      enabled: saveDisabledReason == null,
      disabledReason: saveDisabledReason,
    ));

    return actions;
  }

  /// หาเหตุผลที่ปุ่มถูก disable (ถ้ามี)
  /// คืน null = ไม่ถูก disable
  static String? _getDisabledReason(TransactionModel screenData) {
    if (screenData.isdelete == true) return '${global.language("doc_deleted")} ${global.language("cannot_process")}';
    if (screenData.iscancel) return '${global.language("doc_cancelled")} ${global.language("cannot_process")}';
    if (screenData.isclosedmanual == true) return '${global.language("doc_closed_manual")} ${screenData.closedmanualByName ?? ""}';
    if (screenData.isref == true) return '${global.language("doc_referenced")} ${global.language("cannot_process")}';
    if (screenData.isclosed == true) return '${global.language("doc_fully_received")} ${global.language("cannot_process")}';
    return null;
  }

  /// สร้างปุ่ม action ที่มี icon + label รวมกัน
  /// รองรับ disabled state - ถ้า enabled=false จะแสดง icon จางและกดแสดง SnackBar
  static Widget _buildActionButton({
    required Widget icon,
    required String label,
    required VoidCallback? onPressed,
    bool enabled = true,
    String? disabledReason,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 2.0),
      child: Builder(
        builder: (context) => Opacity(
          opacity: enabled ? 1.0 : 0.35,
          child: InkWell(
            onTap: enabled
                ? onPressed
                : () {
                    if (disabledReason != null) {
                      global.showWarningSnackBar(context, disabledReason);
                    }
                  },
            borderRadius: BorderRadius.circular(8),
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 4.0, vertical: 4.0),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  icon,
                  const SizedBox(height: 2),
                  Text(
                    label,
                    style: TextStyle(fontSize: 9, color: global.theme.onPrimaryColor.withValues(alpha: 0.7)),
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  static Widget _buildLoadingIndicator() {
    return Padding(
      padding: const EdgeInsets.only(right: 12.0),
      child: Center(
        child: SizedBox(
          width: 20,
          height: 20,
          child: CircularProgressIndicator(
            strokeWidth: 2,
            valueColor: AlwaysStoppedAnimation<Color>(global.theme.onPrimaryColor),
          ),
        ),
      ),
    );
  }

  static Widget _buildFullInvoiceButton(
    BuildContext context,
    TransactionModel screenData,
    void Function({required bool taxInvoice}) onSaveOrUpdate,
  ) {
    return _buildActionButton(
      icon: Icon(Icons.receipt_long_outlined, size: 22.0, color: global.theme.onPrimaryColor),
      label: global.language('full_invoice'),
      onPressed: () async {
        final result = await showAlertConfirmFullInvoiceDialog(context, screenData.docno);
        if (result != null && result) {
          onSaveOrUpdate(taxInvoice: true);
        }
      },
    );
  }

  static Widget _buildSlipImageButton(BuildContext context, TransactionModel screenData) {
    return _buildActionButton(
      icon: Icon(Icons.image_outlined, size: 22.0, color: global.theme.onPrimaryColor),
      label: global.language('view_slip'),
      onPressed: () {
        showDialog(
          context: context,
          builder: (BuildContext ctx) {
            return AlertDialog(content: Image.network(screenData.slipurl!));
          },
        );
      },
    );
  }

  static Widget _buildDeleteButton(
    BuildContext context,
    TransactionModel screenData,
    VoidCallback onDeleteDoc, {
    bool enabled = true,
    String? disabledReason,
  }) {
    return _buildActionButton(
      icon: Icon(Icons.delete, size: 22.0, color: global.theme.onPrimaryColor),
      label: global.language('delete'),
      enabled: enabled,
      disabledReason: disabledReason,
      onPressed: () async {
        final result = await showAlertConfirmDeleteDialog(context, screenData.docno);
        if (result != null && result) {
          onDeleteDoc();
        }
      },
    );
  }

  static Widget _buildCancelButton(
    BuildContext context,
    TransactionModel screenData,
    TransactionModel screenDataTemp,
    void Function({required bool taxInvoice}) onSaveOrUpdate, {
    bool enabled = true,
    String? disabledReason,
  }) {
    return _buildActionButton(
      icon: Icon(Icons.cancel_rounded, size: 22.0, color: global.theme.onPrimaryColor),
      label: global.language('cancel'),
      enabled: enabled,
      disabledReason: disabledReason,
      onPressed: () async {
        if (screenData.details!.isNotEmpty) {
          final result = await showAlertConfirmCancelDialog(context);
          if (result != null && result.isNotEmpty) {
            screenData.iscancel = true;
            screenData.cancelreason = result;
            screenData.cancelusercode = global.profileData.username;
            screenData.cancelusername = global.profileData.name;
            screenData.canceldatetime = DateTime.now().toLocal().toIso8601String();
            screenData.canceltime = DateTime.now().toLocal().toIso8601String();

            screenDataTemp.iscancel = true;
            screenDataTemp.cancelreason = result;
            screenDataTemp.cancelusercode = global.profileData.username;
            screenDataTemp.cancelusername = global.profileData.name;
            screenDataTemp.canceldatetime = DateTime.now().toLocal().toIso8601String();
            screenDataTemp.canceltime = DateTime.now().toLocal().toIso8601String();

            onSaveOrUpdate(taxInvoice: false);
          }
        }
      },
    );
  }

  static Widget _buildPrintHistoryButton(
    BuildContext context,
    TransactionModel screenData,
    global.TransactionTypeEnum transactionType,
    String Function(global.TransactionTypeEnum) getCollectionName,
    String Function(global.TransactionTypeEnum) getDocumentTitle, {
    bool enabled = true,
    String? disabledReason,
  }) {
    return _HistoryBadgeButton(
      key: ValueKey('print_history_${screenData.docno}'),
      icon: Icons.print,
      label: global.language('print_history'),
      badgeColor: global.theme.negativeHighlightTextColor,
      enabled: enabled,
      disabledReason: disabledReason,
      countLoader: () async {
        if (screenData.docno.isEmpty) return 0;
        final pdfService = PdfService();
        return pdfService.getPrintHistoryCount(
          collection: getCollectionName(transactionType),
          docNo: screenData.docno,
        );
      },
      onPressed: () {
        final pdfService = PdfService();
        pdfService.showPrintHistoryDialog(
          context: context,
          collection: getCollectionName(transactionType),
          docNo: screenData.docno,
          title: getDocumentTitle(transactionType),
        );
      },
    );
  }

  static Widget _buildEditHistoryButton(BuildContext context, TransactionModel screenData, {
    bool enabled = true,
    String? disabledReason,
  }) {
    return _HistoryBadgeButton(
      key: ValueKey('edit_history_${screenData.docno}'),
      icon: Icons.edit_note,
      label: global.language('edit_history'),
      badgeColor: global.theme.warningHighlightTextColor,
      enabled: enabled,
      disabledReason: disabledReason,
      countLoader: () async {
        if (screenData.docno.isEmpty) return 0;
        final history = await DataHistoryApiService.getPOHistory(screenData.docno);
        return history.length;
      },
      onPressed: () {
        showDataHistoryDialog(
          context,
          docNo: screenData.docno,
          title: global.language('edit_history'),
          creatorCode: screenData.creatorcode,
          creatorName: screenData.creatorname,
          createdAt: screenData.createdat,
        );
      },
    );
  }

  static Widget _buildApprovalHistoryButton(BuildContext context, String docNo, {
    bool enabled = true,
    String? disabledReason,
  }) {
    return _HistoryBadgeButton(
      key: ValueKey('approval_history_$docNo'),
      icon: Icons.approval,
      label: global.language('approval_history'),
      badgeColor: global.theme.positiveHighlightTextColor,
      enabled: enabled,
      disabledReason: disabledReason,
      countLoader: () async {
        if (docNo.isEmpty) return 0;
        final result = await ApprovalApiService.getApprovalTimeline(docNo);
        return result.isSuccess ? result.timeline.length : 0;
      },
      onPressed: () {
        POApprovalHistoryDialog.show(context, docNo);
      },
    );
  }

  static Widget _buildManualCloseButton(
    BuildContext context,
    TransactionModel screenData,
    VoidCallback? onManualCloseToggle, {
    bool enabled = true,
    String? disabledReason,
  }) {
    final isCurrentlyClosed = screenData.isclosedmanual == true;
    final action = isCurrentlyClosed ? 'open' : 'close';
    final icon = isCurrentlyClosed ? Icons.lock_open : Icons.lock;
    final label = isCurrentlyClosed ? global.language('open_document') : global.language('close_document');
    final color = isCurrentlyClosed ? global.theme.positiveHighlightTextColor : Colors.purple.shade700;

    return _buildActionButton(
      icon: Icon(icon, size: 22.0, color: color),
      label: label,
      enabled: enabled,
      disabledReason: disabledReason,
      onPressed: () async {
        // แสดง dialog ให้กรอกเหตุผล
        final reason = await _showManualCloseReasonDialog(context, isCurrentlyClosed);
        if (reason == null || reason.isEmpty) return;

        // เรียก API
        try {
          final rawResponse = await global.goApiPost(
            global.goApiUrlPath('api/purchase-order/manual-close'),
            {
              'shop_id': global.getShopId(),
              'docno': screenData.docno,
              'action': action,
              'action_by_code': global.profileData.username,
              'action_by_name': global.profileData.name,
              'reason': reason,
            },
          );
          final response = rawResponse as Map<String, dynamic>?;

          if (response != null && response['status'] == 'success') {
            // อัปเดต screenData
            if (action == 'close') {
              screenData.isclosedmanual = true;
              screenData.closedmanualByCode = global.profileData.username;
              screenData.closedmanualByName = global.profileData.name;
              screenData.closedmanualAt = DateTime.now().toUtc().toIso8601String();
              screenData.closedmanualReason = reason;
            } else {
              screenData.isclosedmanual = false;
              screenData.closedmanualByCode = null;
              screenData.closedmanualByName = null;
              screenData.closedmanualAt = null;
              screenData.closedmanualReason = null;
            }

            if (context.mounted) {
              global.showSuccessSnackBar(context, action == 'close' ? global.language('close_document_success') : global.language('open_document_success'));
            }

            // เรียก callback เพื่อ refresh UI
            onManualCloseToggle?.call();
          } else {
            final errorMsg = response?['message'] ?? global.language('error_occurred');
            if (context.mounted) {
              global.showErrorSnackBar(context, '$errorMsg');
            }
          }
        } catch (e) {
          AppLogger.error('[ManualClose] Error: $e');
          if (context.mounted) {
            global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
          }
        }
      },
    );
  }

  /// แสดง dialog ให้กรอกเหตุผลการปิด/เปิดเอกสาร
  static Future<String?> _showManualCloseReasonDialog(BuildContext context, bool isCurrentlyClosed) {
    final controller = TextEditingController();
    final title = isCurrentlyClosed ? global.language('open_document') : global.language('manual_close_document');
    final hint = isCurrentlyClosed ? global.language('reason_open_document') : global.language('reason_close_document');

    return showDialog<String>(
      context: context,
      builder: (dialogContext) {
        return AlertDialog(
          title: Row(
            children: [
              Icon(
                isCurrentlyClosed ? Icons.lock_open : Icons.lock,
                color: isCurrentlyClosed ? global.theme.positiveHighlightTextColor : Colors.purple,
              ),
              const SizedBox(width: 8),
              Text(title),
            ],
          ),
          content: TextField(
            controller: controller,
            autofocus: true,
            maxLines: 3,
            decoration: InputDecoration(
              hintText: hint,
              border: const OutlineInputBorder(),
              labelText: global.language('reason_required'),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(dialogContext).pop(null),
              child: Text(global.language('cancel')),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: isCurrentlyClosed ? global.theme.positiveHighlightTextColor : Colors.purple,
                foregroundColor: global.theme.onPrimaryColor,
              ),
              onPressed: () {
                final reason = controller.text.trim();
                if (reason.isEmpty) {
                  global.showWarningSnackBar(dialogContext, global.language('please_enter_reason'));
                  return;
                }
                Navigator.of(dialogContext).pop(reason);
              },
              child: Text(isCurrentlyClosed ? global.language('open_document') : global.language('close_document')),
            ),
          ],
        );
      },
    );
  }

  static Widget _buildAddNewButton(VoidCallback onClearScreen) {
    return _buildActionButton(
      icon: Icon(Icons.add, size: 22.0, color: global.theme.onPrimaryColor),
      label: global.language('add_new'),
      onPressed: onClearScreen,
    );
  }

  static Widget _buildTogglePreviewButton(bool showPreview, VoidCallback onToggle) {
    return _buildActionButton(
      icon: Icon(
        showPreview ? Icons.visibility_off : Icons.visibility,
        size: 22.0,
        color: global.theme.onPrimaryColor,
      ),
      label: showPreview ? global.language('hide_preview') : global.language('show_preview'),
      onPressed: onToggle,
    );
  }

  static Widget _buildSearchButton(VoidCallback onSearch) {
    return _buildActionButton(
      icon: Icon(Icons.list_alt, size: 22.0, color: global.theme.onPrimaryColor),
      label: global.language('search'),
      onPressed: onSearch,
    );
  }

  static Widget _buildSwitchTabButton(TabController tabController) {
    if (tabController.index == 0) {
      return _buildActionButton(
        icon: Icon(Icons.file_open, size: 22.0, color: global.theme.onPrimaryColor),
        label: global.language('view_document'),
        onPressed: () => tabController.animateTo(1),
      );
    } else {
      return _buildActionButton(
        icon: Icon(Icons.text_fields, size: 22.0, color: global.theme.onPrimaryColor),
        label: global.language('edit'),
        onPressed: () => tabController.animateTo(0),
      );
    }
  }

  static Widget _buildSaveButton(
    BuildContext context,
    TransactionModel screenData,
    bool canEditPO,
    bool canPrintPO,
    String editDisabledReason,
    void Function({required bool taxInvoice}) onSaveOrUpdate, {
    bool enabled = true,
    String? disabledReason,
  }) {
    return _buildActionButton(
      icon: Icon(Icons.save, size: 22.0, color: global.theme.onPrimaryColor),
      label: global.language('save'),
      enabled: enabled,
      disabledReason: disabledReason,
      onPressed: () async {
        if (screenData.details!.isNotEmpty) {
          // ตรวจสอบว่า PO สามารถแก้ไขได้หรือไม่
          if (!canEditPO) {
            global.showWarningSnackBar(context, editDisabledReason);
            return;
          }
          // ใช้ dialog แบบใหม่พร้อม checkbox เลือกพิมพ์
          final result = await showEnhancedSaveConfirmDialog(
            context,
            screenData.guidfixed ?? '',
            showPrintOption: canPrintPO,
          );
          if (result != null && result.confirmed) {
            // บันทึก flag ลง SharedPreferences
            final prefs = await SharedPreferences.getInstance();
            await prefs.setBool(_printAfterSaveKey, result.printAfterSave);
            AppLogger.debug('[Save Button] printAfterSave saved: ${result.printAfterSave}');
            onSaveOrUpdate(taxInvoice: false);
          }
        }
      },
    );
  }
}

/// ปุ่มประวัติพร้อม badge แสดงจำนวนครั้ง
/// ใช้ร่วมกันสำหรับประวัติพิมพ์, แก้ไข, อนุมัติ
class _HistoryBadgeButton extends StatefulWidget {
  final IconData icon;
  final String label;
  final Color badgeColor;
  final bool enabled;
  final String? disabledReason;
  final Future<int> Function() countLoader;
  final VoidCallback onPressed;

  const _HistoryBadgeButton({
    super.key,
    required this.icon,
    required this.label,
    required this.badgeColor,
    required this.countLoader,
    required this.onPressed,
    this.enabled = true,
    this.disabledReason,
  });

  @override
  State<_HistoryBadgeButton> createState() => _HistoryBadgeButtonState();
}

class _HistoryBadgeButtonState extends State<_HistoryBadgeButton> with global.ThemeRefreshMixin {
  int _count = 0;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadCount();
  }

  @override
  void didUpdateWidget(_HistoryBadgeButton oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.key != widget.key) {
      _loadCount();
    }
  }

  Future<void> _loadCount() async {
    if (!widget.enabled) {
      if (mounted) setState(() => _isLoading = false);
      return;
    }

    try {
      final count = await widget.countLoader();
      if (mounted) {
        setState(() {
          _count = count;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 2.0),
      child: Opacity(
        opacity: widget.enabled ? 1.0 : 0.35,
        child: InkWell(
          onTap: widget.enabled
              ? () {
                  widget.onPressed();
                  // โหลดจำนวนใหม่หลังปิด dialog
                  Future.delayed(const Duration(milliseconds: 500), () {
                    if (mounted) _loadCount();
                  });
                }
              : () {
                  if (widget.disabledReason != null) {
                    global.showWarningSnackBar(context, widget.disabledReason!);
                  }
                },
          borderRadius: BorderRadius.circular(8),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 4.0, vertical: 4.0),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Stack(
                  clipBehavior: Clip.none,
                  children: [
                    Icon(widget.icon, size: 20.0, color: global.theme.onPrimaryColor),
                    if (!_isLoading && _count > 0)
                      Positioned(
                        right: -8,
                        top: -8,
                        child: Container(
                          padding: const EdgeInsets.all(2),
                          decoration: BoxDecoration(
                            color: widget.badgeColor,
                            shape: BoxShape.circle,
                            border: Border.all(color: global.theme.onPrimaryColor, width: 1.5),
                          ),
                          constraints: const BoxConstraints(minWidth: 16, minHeight: 16),
                          child: Text(
                            _count > 99 ? '99+' : _count.toString(),
                            style: TextStyle(color: global.theme.onPrimaryColor, fontSize: 8, fontWeight: FontWeight.bold),
                            textAlign: TextAlign.center,
                          ),
                        ),
                      ),
                  ],
                ),
                const SizedBox(height: 2),
                Text(
                  widget.label,
                  style: TextStyle(fontSize: 9, color: global.theme.onPrimaryColor.withValues(alpha: 0.7)),
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
