import 'package:flutter/material.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/screens/quotation/quotation_list_screen.dart';
import 'package:split_view/split_view.dart';
import 'package:smlaicloud/global.dart' as global;

/// Widget สำหรับจัดการ Responsive Layout ของหน้าจอใบเสนอราคา
///
/// รับผิดชอบ:
/// - SplitView สำหรับแสดง Preview/Edit บนจอใหญ่
/// - TabBarView สำหรับจอเล็ก
/// - Side Panel สำหรับค้นหาเอกสาร (จอใหญ่)
/// - Handle สำหรับปรับขนาด Panel
class QTResponsiveLayout extends StatefulWidget {
  final SplitViewController splitViewController;
  final TabController tabController;
  final bool showPreview;
  final bool showDocSearchPanel;
  final double docSearchPanelWidth;
  final GlobalKey<QuotationListScreenState> qtListKey;
  final Widget Function() buildEditWidget;
  final Widget Function() buildPreviewWidget;
  final void Function(TransactionModel) onDocumentSelected;
  final VoidCallback onClosePanel;
  final void Function(double) onPanelWidthChanged;

  const QTResponsiveLayout({
    super.key,
    required this.splitViewController,
    required this.tabController,
    required this.showPreview,
    required this.showDocSearchPanel,
    required this.docSearchPanelWidth,
    required this.qtListKey,
    required this.buildEditWidget,
    required this.buildPreviewWidget,
    required this.onDocumentSelected,
    required this.onClosePanel,
    required this.onPanelWidthChanged,
  });

  @override
  State<QTResponsiveLayout> createState() => _QTResponsiveLayoutState();
}

class _QTResponsiveLayoutState extends State<QTResponsiveLayout> {
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
              gripColorActive: Colors.blue,
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

    // ถ้าเปิด side panel ค้นหาเอกสาร (เฉพาะจอใหญ่)
    if (widget.showDocSearchPanel && isLargeScreen) {
      const minWidth = 450.0;
      final maxWidth = constraints.maxWidth * 0.75;

      return Row(
        children: [
          // Side panel ค้นหาเอกสาร (ด้านซ้าย)
          SizedBox(
            width: widget.docSearchPanelWidth.clamp(minWidth, maxWidth),
            child: QuotationListScreen(
              key: widget.qtListKey,
              custcode: '',
              isEmbedded: true,
              onDocumentSelected: widget.onDocumentSelected,
              onClose: widget.onClosePanel,
            ),
          ),
          // Handle สำหรับปรับขนาด panel
          _buildResizeHandle(minWidth, maxWidth),
          // เนื้อหาหลัก
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
                color: Colors.grey[400],
                borderRadius: BorderRadius.circular(2),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
