import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import '../../models/form_template.dart';

class FormToolbar extends StatelessWidget {
  final bool canUndo;
  final bool canRedo;
  final VoidCallback onUndo;
  final VoidCallback onRedo;
  final double zoomLevel;
  final Function(double) onZoomChanged;
  final bool showGrid;
  final bool snapToGrid;
  final bool showRulers;
  final VoidCallback onToggleGrid;
  final VoidCallback onToggleSnap;
  final VoidCallback onToggleRulers;
  final VoidCallback onAlignLeft;
  final VoidCallback onAlignCenter;
  final VoidCallback onAlignRight;
  final VoidCallback onAlignTop;
  final VoidCallback onAlignMiddle;
  final VoidCallback onAlignBottom;
  final VoidCallback onDistributeH;
  final VoidCallback onDistributeV;
  final VoidCallback onBringToFront;
  final VoidCallback onSendToBack;
  final bool hasSelection;
  final bool hasMultiSelection;
  final VoidCallback onPreviewToggle;
  final bool isPreviewMode;
  final VoidCallback onSave;
  final VoidCallback onLoad;
  final VoidCallback onShowTemplateGallery;
  final VoidCallback onShowFormInfo;
  final FormTemplate template;
  final VoidCallback? onToggleLeftPanel;
  final VoidCallback? onToggleRightPanel;
  final bool showLeftPanel;
  final bool showRightPanel;
  final VoidCallback? onShowJsonEditor;

  const FormToolbar({
    super.key,
    required this.canUndo,
    required this.canRedo,
    required this.onUndo,
    required this.onRedo,
    required this.zoomLevel,
    required this.onZoomChanged,
    required this.showGrid,
    required this.snapToGrid,
    required this.showRulers,
    required this.onToggleGrid,
    required this.onToggleSnap,
    required this.onToggleRulers,
    required this.onAlignLeft,
    required this.onAlignCenter,
    required this.onAlignRight,
    required this.onAlignTop,
    required this.onAlignMiddle,
    required this.onAlignBottom,
    required this.onDistributeH,
    required this.onDistributeV,
    required this.onBringToFront,
    required this.onSendToBack,
    required this.hasSelection,
    this.hasMultiSelection = false,
    required this.onPreviewToggle,
    required this.isPreviewMode,
    required this.onSave,
    required this.onLoad,
    required this.onShowTemplateGallery,
    required this.onShowFormInfo,
    required this.template,
    this.onToggleLeftPanel,
    this.onToggleRightPanel,
    this.showLeftPanel = true,
    this.showRightPanel = true,
    this.onShowJsonEditor,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 44,
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        border: Border(
          bottom: BorderSide(color: global.theme.dividerBorderColor),
        ),
      ),
      child: Row(
        children: [
          const SizedBox(width: 4),
          // Undo / Redo
          _buildIconButton(
            icon: Icons.undo,
            tooltip: global.language('fd_undo'),
            onPressed: canUndo ? onUndo : null,
          ),
          _buildIconButton(
            icon: Icons.redo,
            tooltip: global.language('fd_redo'),
            onPressed: canRedo ? onRedo : null,
          ),
          _buildDivider(),
          // Zoom
          _buildZoomControls(),
          _buildDivider(),
          // Grid / Snap / Ruler toggles
          _buildToggleButton(
            icon: Icons.grid_on,
            tooltip: global.language('fd_grid'),
            isActive: showGrid,
            onPressed: onToggleGrid,
          ),
          _buildToggleButton(
            icon: Icons.grid_goldenratio,
            tooltip: global.language('fd_snap_to_grid'),
            isActive: snapToGrid,
            onPressed: onToggleSnap,
          ),
          _buildToggleButton(
            icon: Icons.straighten,
            tooltip: global.language('fd_rulers'),
            isActive: showRulers,
            onPressed: onToggleRulers,
          ),
          _buildDivider(),
          // Alignment
          _buildIconButton(
            icon: Icons.align_horizontal_left,
            tooltip: global.language('fd_align_left'),
            onPressed: hasMultiSelection ? onAlignLeft : null,
          ),
          _buildIconButton(
            icon: Icons.align_horizontal_center,
            tooltip: global.language('fd_align_center'),
            onPressed: hasMultiSelection ? onAlignCenter : null,
          ),
          _buildIconButton(
            icon: Icons.align_horizontal_right,
            tooltip: global.language('fd_align_right'),
            onPressed: hasMultiSelection ? onAlignRight : null,
          ),
          _buildIconButton(
            icon: Icons.align_vertical_top,
            tooltip: global.language('fd_align_top'),
            onPressed: hasMultiSelection ? onAlignTop : null,
          ),
          _buildIconButton(
            icon: Icons.align_vertical_center,
            tooltip: global.language('fd_align_middle'),
            onPressed: hasMultiSelection ? onAlignMiddle : null,
          ),
          _buildIconButton(
            icon: Icons.align_vertical_bottom,
            tooltip: global.language('fd_align_bottom'),
            onPressed: hasMultiSelection ? onAlignBottom : null,
          ),
          _buildDivider(),
          // Distribute
          _buildIconButton(
            icon: Icons.horizontal_distribute,
            tooltip: global.language('fd_distribute_h'),
            onPressed: hasMultiSelection ? onDistributeH : null,
          ),
          _buildIconButton(
            icon: Icons.vertical_distribute,
            tooltip: global.language('fd_distribute_v'),
            onPressed: hasMultiSelection ? onDistributeV : null,
          ),
          _buildDivider(),
          // Z-order
          _buildIconButton(
            icon: Icons.flip_to_front,
            tooltip: global.language('fd_bring_to_front'),
            onPressed: hasSelection ? onBringToFront : null,
          ),
          _buildIconButton(
            icon: Icons.flip_to_back,
            tooltip: global.language('fd_send_to_back'),
            onPressed: hasSelection ? onSendToBack : null,
          ),
          _buildDivider(),
          const Spacer(),
          // Preview / Info / Templates / Load / Save
          _buildToggleButton(
            icon: Icons.preview,
            tooltip: global.language('form_design_preview'),
            isActive: isPreviewMode,
            onPressed: onPreviewToggle,
          ),
          _buildIconButton(
            icon: Icons.info_outline,
            tooltip: global.language('form_design_form_info'),
            onPressed: onShowFormInfo,
          ),
          _buildIconButton(
            icon: Icons.dashboard_customize,
            tooltip: global.language('fd_templates'),
            onPressed: onShowTemplateGallery,
          ),
          _buildIconButton(
            icon: Icons.folder_open,
            tooltip: global.language('fd_load'),
            onPressed: onLoad,
          ),
          _buildIconButton(
            icon: Icons.save,
            tooltip: global.language('fd_save'),
            onPressed: onSave,
          ),
          if (onShowJsonEditor != null) ...[
            _buildDivider(),
            _buildIconButton(
              icon: Icons.code,
              tooltip: global.language('fd_json_editor'),
              onPressed: onShowJsonEditor,
            ),
          ],
          _buildDivider(),
          _buildToggleButton(
            icon: Icons.view_sidebar,
            tooltip: global.language('fd_left_panel'),
            isActive: showLeftPanel,
            onPressed: onToggleLeftPanel ?? () {},
          ),
          _buildToggleButton(
            icon: Icons.view_sidebar_outlined,
            tooltip: global.language('fd_right_panel'),
            isActive: showRightPanel,
            onPressed: onToggleRightPanel ?? () {},
          ),
          const SizedBox(width: 4),
        ],
      ),
    );
  }

  Widget _buildIconButton({
    required IconData icon,
    required String tooltip,
    VoidCallback? onPressed,
  }) {
    return Tooltip(
      message: tooltip,
      child: SizedBox(
        width: 32,
        height: 32,
        child: IconButton(
          padding: EdgeInsets.zero,
          iconSize: 18,
          icon: Icon(icon),
          color: onPressed != null
              ? global.theme.iconColor
              : global.theme.iconSecondaryColor.withValues(alpha: 0.4),
          onPressed: onPressed,
        ),
      ),
    );
  }

  Widget _buildToggleButton({
    required IconData icon,
    required String tooltip,
    required bool isActive,
    required VoidCallback onPressed,
  }) {
    return Tooltip(
      message: tooltip,
      child: SizedBox(
        width: 32,
        height: 32,
        child: IconButton(
          padding: EdgeInsets.zero,
          iconSize: 18,
          icon: Icon(icon),
          color: isActive
              ? global.theme.primaryColor
              : global.theme.iconSecondaryColor,
          onPressed: onPressed,
        ),
      ),
    );
  }

  Widget _buildDivider() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      child: Container(
        width: 1,
        height: 24,
        color: global.theme.dividerBorderColor,
      ),
    );
  }

  Widget _buildZoomControls() {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        SizedBox(
          width: 120,
          child: SliderTheme(
            data: SliderThemeData(
              trackHeight: 2,
              thumbShape: const RoundSliderThumbShape(enabledThumbRadius: 6),
              activeTrackColor: global.theme.primaryColor,
              inactiveTrackColor: global.theme.dividerBorderColor,
              thumbColor: global.theme.primaryColor,
              overlayShape: const RoundSliderOverlayShape(overlayRadius: 12),
            ),
            child: Slider(
              value: zoomLevel.clamp(0.25, 2.0),
              min: 0.25,
              max: 2.0,
              onChanged: onZoomChanged,
            ),
          ),
        ),
        SizedBox(
          width: 40,
          child: Text(
            '${(zoomLevel * 100).round()}%',
            style: TextStyle(
              fontSize: 11,
              color: global.theme.textSecondaryColor,
            ),
          ),
        ),
      ],
    );
  }
}
