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

/// Widget สำหรับสร้าง AppBar Actions ของหน้าจอใบเสนอราคา
///
/// รับผิดชอบ:
/// - ปุ่มลบเอกสาร
/// - ปุ่มยกเลิกเอกสาร
/// - ปุ่มดูประวัติการพิมพ์
/// - ปุ่มดูประวัติการแก้ไข
/// - ปุ่มดูประวัติการอนุมัติ
/// - ปุ่มเพิ่มใหม่
/// - ปุ่ม Toggle Preview
/// - ปุ่มค้นหาเอกสาร
/// - ปุ่มบันทึก
///
/// ต่างจาก PO: ไม่มีปุ่มปิดเอกสารด้วยมือ, ไม่มีปุ่มออกใบกำกับเต็ม
class QTAppBarActions {
  static const String _printAfterSaveKey = 'transaction_print_after_save_flag';

  /// สร้าง list ของ AppBar actions สำหรับ Quotation Screen
  static List<Widget> build({
    required BuildContext context,
    required TransactionModel screenData,
    required TransactionModel screenDataTemp,
    required global.TransactionTypeEnum transactionType,
    required bool isLoading,
    required bool showPreview,
    required bool canEditQT,
    required bool canPrintQT,
    required String editDisabledReason,
    required String Function(global.TransactionTypeEnum) getCollectionName,
    required String Function(global.TransactionTypeEnum) getDocumentTitle,
    required TabController tabController,
    required VoidCallback onClearScreen,
    required VoidCallback onTogglePreview,
    required VoidCallback onSearchTrans,
    required VoidCallback onDeleteDoc,
    required void Function({required bool taxInvoice}) onSaveOrUpdate,
  }) {
    final List<Widget> actions = [];
    final bool hasDoc = (screenData.guidfixed?.isNotEmpty ?? false) && screenData.docno.isNotEmpty;
    final disabledReason = _getDisabledReason(screenData);
    final noDocReason = hasDoc ? null : global.language("no_document_yet");

    // Loading indicator
    if (isLoading) {
      actions.add(_buildLoadingIndicator());
    }

    // ดู slip image (ถ้ามี)
    if (screenData.slipurl != null && screenData.slipurl != '') {
      actions.add(_buildSlipImageButton(context, screenData));
    }

    // ลบเอกสาร
    final deleteReason = noDocReason ?? disabledReason;
    actions.add(_buildDeleteButton(context, screenData, onDeleteDoc,
        enabled: deleteReason == null, disabledReason: deleteReason));

    // ยกเลิกเอกสาร
    final cancelReason = noDocReason ?? disabledReason;
    actions.add(_buildCancelButton(context, screenData, screenDataTemp, onSaveOrUpdate,
        enabled: cancelReason == null, disabledReason: cancelReason));

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

    // บันทึก
    final saveDisabledReason = disabledReason;
    actions.add(_buildSaveButton(
      context,
      screenData,
      canEditQT,
      canPrintQT,
      editDisabledReason,
      onSaveOrUpdate,
      enabled: saveDisabledReason == null,
      disabledReason: saveDisabledReason,
    ));

    return actions;
  }

  static String? _getDisabledReason(TransactionModel screenData) {
    if (screenData.isdelete == true) return global.language("document_deleted_cannot_proceed");
    if (screenData.iscancel) return global.language("document_cancelled_cannot_proceed");
    if (screenData.isref == true) return global.language("document_referenced_cannot_proceed");
    return null;
  }

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
      key: ValueKey('qt_print_history_${screenData.docno}'),
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
      key: ValueKey('qt_edit_history_${screenData.docno}'),
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
      key: ValueKey('qt_approval_history_$docNo'),
      icon: Icons.approval,
      label: global.language("approval_history"),
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
    bool canEditQT,
    bool canPrintQT,
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
          // ตรวจสอบว่าใบเสนอราคาสามารถแก้ไขได้หรือไม่
          if (!canEditQT) {
            global.showWarningSnackBar(context, editDisabledReason);
            return;
          }
          // ใช้ dialog แบบใหม่พร้อม checkbox เลือกพิมพ์
          final result = await showEnhancedSaveConfirmDialog(
            context,
            screenData.guidfixed ?? '',
            showPrintOption: canPrintQT,
          );
          if (result != null && result.confirmed) {
            final prefs = await SharedPreferences.getInstance();
            await prefs.setBool(_printAfterSaveKey, result.printAfterSave);
            AppLogger.debug('[QT Save Button] printAfterSave saved: ${result.printAfterSave}');
            onSaveOrUpdate(taxInvoice: false);
          }
        }
      },
    );
  }
}

/// ปุ่มประวัติพร้อม badge แสดงจำนวนครั้ง
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
