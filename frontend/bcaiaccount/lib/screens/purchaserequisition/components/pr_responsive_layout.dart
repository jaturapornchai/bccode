import 'package:flutter/material.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/screens/purchaserequisition/purchaserequisition_list_screen.dart';
import 'package:split_view/split_view.dart';
import 'package:smlaicloud/global.dart' as global;

/// Widget สำหรับจัดการ Responsive Layout ของหน้าจอ PR (ใบขอซื้อ)
///
/// รับผิดชอบ:
/// - SplitView สำหรับแสดง Preview/Edit บนจอใหญ่
/// - TabBarView สำหรับจอเล็ก
/// - Side Panel สำหรับค้นหาเอกสาร (จอใหญ่)
class PRResponsiveLayout extends StatefulWidget {
  final SplitViewController splitViewController;
  final TabController tabController;
  final bool showPreview;
  final bool showDocSearchPanel;
  final double docSearchPanelWidth;
  final GlobalKey<PurchaseRequisitionListScreenState> prListKey;
  final Widget Function() buildEditWidget;
  final Widget Function() buildPreviewWidget;
  final void Function(TransactionModel) onDocumentSelected;
  final VoidCallback onClosePanel;
  final void Function(double) onPanelWidthChanged;

  const PRResponsiveLayout({
    super.key,
    required this.splitViewController,
    required this.tabController,
    required this.showPreview,
    required this.showDocSearchPanel,
    required this.docSearchPanelWidth,
    required this.prListKey,
    required this.buildEditWidget,
    required this.buildPreviewWidget,
    required this.onDocumentSelected,
    required this.onClosePanel,
    required this.onPanelWidthChanged,
  });

  @override
  State<PRResponsiveLayout> createState() => _PRResponsiveLayoutState();
}

class _PRResponsiveLayoutState extends State<PRResponsiveLayout> with global.ThemeRefreshMixin {
  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        return _buildMainContent(constraints);
      },
    );
  }

  Widget _buildMainContent(BoxConstraints constraints) {
    final isLargeScreen = constraints.maxWidth > 700;

    Widget mainContent;
    if (isLargeScreen) {
      mainContent = widget.showPreview
          ? SplitView(
              controller: widget.splitViewController,
              gripSize: 8,
              gripColor: global.theme.appBarColor,
              gripColorActive: global.theme.infoHighlightTextColor,
              viewMode: SplitViewMode.Horizontal,
              indicator: const SplitIndicator(viewMode: SplitViewMode.Horizontal),
              activeIndicator: const SplitIndicator(viewMode: SplitViewMode.Horizontal, isActive: true),
              children: [widget.buildPreviewWidget(), widget.buildEditWidget()],
            )
          : widget.buildEditWidget();
    } else {
      mainContent = TabBarView(
        physics: const NeverScrollableScrollPhysics(),
        controller: widget.tabController,
        children: widget.showPreview ? [widget.buildEditWidget(), widget.buildPreviewWidget()] : [widget.buildEditWidget()],
      );
    }

    // Side panel ค้นหาเอกสาร (เฉพาะจอใหญ่)
    if (widget.showDocSearchPanel && isLargeScreen) {
      const minWidth = 450.0;
      final maxWidth = constraints.maxWidth * 0.75;

      return Row(
        children: [
          SizedBox(
            width: widget.docSearchPanelWidth.clamp(minWidth, maxWidth),
            child: PurchaseRequisitionListScreen(
              key: widget.prListKey,
              custcode: '',
              isEmbedded: true,
              onDocumentSelected: widget.onDocumentSelected,
              onClose: widget.onClosePanel,
            ),
          ),
          _buildResizeHandle(minWidth, maxWidth),
          Expanded(child: mainContent),
        ],
      );
    }

    return mainContent;
  }

  Widget _buildResizeHandle(double minWidth, double maxWidth) {
    return MouseRegion(
      cursor: SystemMouseCursors.resizeColumn,
      child: GestureDetector(
        onHorizontalDragUpdate: (details) {
          double newWidth = widget.docSearchPanelWidth + details.delta.dx;
          newWidth = newWidth.clamp(minWidth, maxWidth);
          widget.onPanelWidthChanged(newWidth);
        },
        child: Container(
          width: 8,
          color: Colors.transparent,
          child: Center(
            child: Container(
              width: 4,
              height: 40,
              decoration: BoxDecoration(
                color: global.theme.dividerBorderColor,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
